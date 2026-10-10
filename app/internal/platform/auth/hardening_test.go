package auth_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"

	"github.com/luiszkm/template-go/internal/platform/auth"
	"github.com/luiszkm/template-go/internal/platform/httpx"
	"github.com/luiszkm/template-go/internal/platform/op"
	"github.com/luiszkm/template-go/internal/platform/testkit"
)

type countingQuerier struct {
	auth.Querier
	calls atomic.Int32
}

func (c *countingQuerier) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	c.calls.Add(1)
	return c.Querier.Exec(ctx, sql, args...)
}

func (c *countingQuerier) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	c.calls.Add(1)
	return c.Querier.Query(ctx, sql, args...)
}

func (c *countingQuerier) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	c.calls.Add(1)
	return c.Querier.QueryRow(ctx, sql, args...)
}

var errInjected = errors.New("injected: connection reset by peer")

type failingRow struct{}

func (failingRow) Scan(...any) error { return errInjected }

type failingQuerier struct{ auth.Querier }

func (failingQuerier) Exec(context.Context, string, ...any) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, errInjected
}

func (failingQuerier) Query(context.Context, string, ...any) (pgx.Rows, error) {
	return nil, errInjected
}

func (failingQuerier) QueryRow(context.Context, string, ...any) pgx.Row { return failingRow{} }

type syncBuf struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (s *syncBuf) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.buf.Write(p)
}

func (s *syncBuf) entries(t *testing.T) []map[string]any {
	t.Helper()
	s.mu.Lock()
	text := s.buf.String()
	s.mu.Unlock()
	var out []map[string]any
	for line := range strings.Lines(text) {
		var m map[string]any
		require.NoError(t, json.Unmarshal([]byte(line), &m), line)
		out = append(out, m)
	}
	return out
}

func sessionAged(t *testing.T, pool *pgxpool.Pool, age time.Duration) []byte {
	t.Helper()
	u := testkit.SignIn(t, pool)
	_, err := pool.Exec(t.Context(), `UPDATE sessions SET created_at = now() - make_interval(secs => $1) WHERE user_id = $2`,
		age.Seconds(), u.ID)
	require.NoError(t, err)
	return auth.HashToken(u.Cookie.Value)
}

func exists(t *testing.T, pool *pgxpool.Pool, hash []byte) bool {
	t.Helper()
	var n int
	require.NoError(t, pool.QueryRow(t.Context(), `SELECT count(*) FROM sessions WHERE token_hash = $1`, hash).Scan(&n))
	return n == 1
}

func sweep(ctx context.Context, q auth.Querier, every time.Duration, log *slog.Logger) <-chan struct{} {
	done := make(chan struct{})
	go func() {
		auth.SweepSessions(ctx, q, time.Hour, every, log)
		close(done)
	}()
	return done
}

func TestMiddleware_LookupFailure500Logged(t *testing.T) {
	pool := testkit.MigratedDB(t)
	logs := &syncBuf{}
	mux := http.NewServeMux()
	api := httpx.NewAPI(mux)
	auth.Install(api, failingQuerier{Querier: pool}, ttl)
	require.NoError(t, op.Register(api, op.Spec{ID: "authn", Method: http.MethodGet, Path: "/authn", Authenticated: true},
		func(context.Context, *empty) (*ok, error) { return &ok{}, nil }))
	h := httpx.Wrap(mux, httpx.Options{Logger: slog.New(slog.NewJSONHandler(logs, nil))})

	rec := testkit.Do(t, h, testkit.Request{Method: http.MethodGet, Path: "/authn",
		Cookie: &http.Cookie{Name: auth.CookieName, Value: "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"}})
	require.Equal(t, http.StatusInternalServerError, rec.Code)
	body := testkit.JSON[map[string]any](t, rec)
	require.Equal(t, "internal server error", body["detail"])
	require.NotContains(t, body, "errors")
	require.NotContains(t, rec.Body.String(), "connection reset")

	entries := logs.entries(t)
	require.Len(t, entries, 1)
	require.Equal(t, "ERROR", entries[0]["level"])
	require.Contains(t, entries[0]["error"], errInjected.Error())
}

func TestSweepSessions_DeletesOnlyExpired(t *testing.T) {
	pool := testkit.MigratedDB(t)
	past := sessionAged(t, pool, 61*time.Minute)
	long := sessionAged(t, pool, 13*time.Hour)
	inside := sessionAged(t, pool, 59*time.Minute)
	fresh := sessionAged(t, pool, 0)

	ctx, cancel := context.WithCancel(t.Context())
	done := sweep(ctx, pool, time.Hour, testkit.DiscardLogger())
	require.Eventually(t, func() bool { return !exists(t, pool, past) && !exists(t, pool, long) }, 5*time.Second, 10*time.Millisecond)
	cancel()
	<-done
	require.True(t, exists(t, pool, inside))
	require.True(t, exists(t, pool, fresh))
}

func TestSweepSessions_RunsAtStartAndEveryInterval(t *testing.T) {
	pool := testkit.MigratedDB(t)
	before := sessionAged(t, pool, 2*time.Hour)
	later := sessionAged(t, pool, 0)

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	start := &countingQuerier{Querier: pool}
	done := sweep(ctx, start, time.Hour, testkit.DiscardLogger())
	require.Eventually(t, func() bool { return !exists(t, pool, before) }, 2*time.Second, 5*time.Millisecond,
		"the first sweep runs at start, without waiting for the interval")
	cancel()
	<-done

	ctx, cancel = context.WithCancel(t.Context())
	defer cancel()
	ticking := &countingQuerier{Querier: pool}
	done = sweep(ctx, ticking, 20*time.Millisecond, testkit.DiscardLogger())
	require.Eventually(t, func() bool { return ticking.calls.Load() >= 1 }, time.Second, 5*time.Millisecond)
	_, err := pool.Exec(t.Context(), `UPDATE sessions SET created_at = now() - interval '2 hours' WHERE token_hash = $1`, later)
	require.NoError(t, err)
	require.Eventually(t, func() bool { return !exists(t, pool, later) }, 200*time.Millisecond, 5*time.Millisecond)
	require.GreaterOrEqual(t, ticking.calls.Load(), int32(2))
	cancel()
	<-done
}

func TestSweepSessions_FailureLogsAndRetries(t *testing.T) {
	pool := testkit.MigratedDB(t)
	logs := &syncBuf{}
	ctx, cancel := context.WithCancel(t.Context())
	done := sweep(ctx, failingQuerier{Querier: pool}, 10*time.Millisecond, slog.New(slog.NewJSONHandler(logs, nil)))

	warns := func() int {
		n := 0
		for _, e := range logs.entries(t) {
			if e["level"] == "WARN" && e["error"] == errInjected.Error() {
				n++
			}
		}
		return n
	}
	require.Eventually(t, func() bool { return warns() >= 2 }, 200*time.Millisecond, 5*time.Millisecond)
	select {
	case <-done:
		t.Fatal("SweepSessions returned after a failure")
	default:
	}
	cancel()
	<-done
}

func TestSweepSessions_ReturnsOnCancel(t *testing.T) {
	pool := testkit.MigratedDB(t)
	ctx, cancel := context.WithCancel(t.Context())
	done := sweep(ctx, pool, time.Hour, testkit.DiscardLogger())
	time.Sleep(20 * time.Millisecond)
	cancel()
	select {
	case <-done:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("SweepSessions did not return within 100ms of cancel")
	}
}

func TestLookup_SingleQuery(t *testing.T) {
	pool := testkit.MigratedDB(t)
	u := testkit.SignIn(t, pool, "a:read", "b:write")
	testkit.GrantPermissions(t, pool, u.ID, "b:write", "c:delete")

	q := &countingQuerier{Querier: pool}
	p, err := auth.Lookup(t.Context(), q, u.Cookie.Value, ttl)
	require.NoError(t, err)
	require.EqualValues(t, 1, q.calls.Load())
	require.Equal(t, u.ID, p.UserID)
	require.Equal(t, map[op.Permission]struct{}{"a:read": {}, "b:write": {}, "c:delete": {}}, p.Permissions)
}

func TestLookup_UserWithoutRoles(t *testing.T) {
	pool := testkit.MigratedDB(t)
	u := testkit.SignIn(t, pool)
	p, err := auth.Lookup(t.Context(), pool, u.Cookie.Value, ttl)
	require.NoError(t, err)
	require.Equal(t, u.ID, p.UserID)
	require.Empty(t, p.Permissions)

	h, _ := server(t, pool)
	require.Equal(t, http.StatusForbidden, get(t, h, "/perm", u.Cookie))
	require.Equal(t, http.StatusOK, get(t, h, "/authn", u.Cookie))

	_, err = auth.Lookup(t.Context(), pool, uuid.NewString(), ttl)
	require.ErrorIs(t, err, auth.ErrNoSession)
}
