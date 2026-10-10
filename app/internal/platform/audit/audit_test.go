package audit_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"

	"github.com/luiszkm/template-go/internal/platform/audit"
	"github.com/luiszkm/template-go/internal/platform/auth"
	"github.com/luiszkm/template-go/internal/platform/db"
	"github.com/luiszkm/template-go/internal/platform/httpx"
	"github.com/luiszkm/template-go/internal/platform/testkit"
)

func requestContext(t *testing.T, actor uuid.UUID, rid string) context.Context {
	t.Helper()
	var ctx context.Context
	h := httpx.RequestID(httpx.WithClientIP(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		ctx = auth.WithPrincipal(r.Context(), auth.Principal{UserID: actor})
	}), nil))
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/", nil)
	req.RemoteAddr = "203.0.113.7:5555"
	req.Header.Set(httpx.HeaderRequestID, rid)
	h.ServeHTTP(httptest.NewRecorder(), req)
	return ctx
}

func count(t *testing.T, pool *pgxpool.Pool, sql string, args ...any) int {
	t.Helper()
	var n int
	require.NoError(t, pool.QueryRow(t.Context(), sql, args...).Scan(&n))
	return n
}

func TestRecord_CommitsWithTx(t *testing.T) {
	pool := testkit.MigratedDB(t)
	actor := testkit.SignIn(t, pool).ID
	ctx := requestContext(t, actor, "req-28")

	require.NoError(t, db.WithTx(ctx, pool, func(tx pgx.Tx) error {
		return audit.Record(ctx, tx, audit.Event{Action: "thing.changed", ResourceType: "thing", ResourceID: "t-1",
			Before: map[string]string{"name": "a"}, After: map[string]string{"name": "b"}})
	}))

	var (
		gotActor                    uuid.UUID
		action, rtype, rid, request string
		before, after               string
		ip                          netip.Addr
	)
	require.NoError(t, pool.QueryRow(t.Context(), `
		SELECT actor_id, action, resource_type, resource_id, before::text, after::text, ip, request_id FROM audit_events`).
		Scan(&gotActor, &action, &rtype, &rid, &before, &after, &ip, &request))
	require.Equal(t, 1, count(t, pool, `SELECT count(*) FROM audit_events`))
	require.Equal(t, actor, gotActor)
	require.Equal(t, "thing.changed", action)
	require.Equal(t, "thing", rtype)
	require.Equal(t, "t-1", rid)
	require.JSONEq(t, `{"name":"a"}`, before)
	require.JSONEq(t, `{"name":"b"}`, after)
	require.Equal(t, netip.MustParseAddr("203.0.113.7"), ip)
	require.Equal(t, "req-28", request)
}

func TestRecord_RolledBackWithTx(t *testing.T) {
	pool := testkit.MigratedDB(t)
	ctx := requestContext(t, testkit.SignIn(t, pool).ID, "req-29")
	boom := errors.New("boom")

	err := db.WithTx(ctx, pool, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, `INSERT INTO roles (name) VALUES ('rolled-back')`); err != nil {
			return err
		}
		if err := audit.Record(ctx, tx, audit.Event{Action: "role.created", ResourceType: "role", ResourceID: "r"}); err != nil {
			return err
		}
		return boom
	})
	require.ErrorIs(t, err, boom)
	require.Zero(t, count(t, pool, `SELECT count(*) FROM audit_events`))
	require.Zero(t, count(t, pool, `SELECT count(*) FROM roles WHERE name = 'rolled-back'`))
}

func TestRecord_RedactsPasswordKeys(t *testing.T) {
	pool := testkit.MigratedDB(t)
	ctx := requestContext(t, testkit.SignIn(t, pool).ID, "req-30")
	type user struct {
		Email        string `json:"email"`
		PasswordHash string `json:"password_hash"`
	}
	before := user{Email: "a@x.com", PasswordHash: "$argon2id$secret"}
	after := map[string]any{
		"email": "a@x.com", "password": "p1", "current_password": "p2", "new_password": "p3",
		"nested": map[string]any{"password_hash": "h", "keep": 1},
		"list":   []any{map[string]any{"password": "p4", "keep": 2}},
	}
	require.NoError(t, db.WithTx(ctx, pool, func(tx pgx.Tx) error {
		return audit.Record(ctx, tx, audit.Event{Action: "user.updated", ResourceType: "user", ResourceID: "u", Before: before, After: after})
	}))
	var b, a string
	require.NoError(t, pool.QueryRow(t.Context(), `SELECT before::text, after::text FROM audit_events`).Scan(&b, &a))
	require.JSONEq(t, `{"email":"a@x.com"}`, b)
	require.JSONEq(t, `{"email":"a@x.com","nested":{"keep":1},"list":[{"keep":2}]}`, a)
}

func TestAuditEvents_AppendOnly(t *testing.T) {
	pool := testkit.MigratedDB(t)
	ctx := requestContext(t, testkit.SignIn(t, pool).ID, "req-31")
	require.NoError(t, db.WithTx(ctx, pool, func(tx pgx.Tx) error {
		return audit.Record(ctx, tx, audit.Event{Action: "a", ResourceType: "r", ResourceID: "1"})
	}))
	_, err := pool.Exec(t.Context(), `UPDATE audit_events SET action = 'x'`)
	require.ErrorContains(t, err, "append-only")
	_, err = pool.Exec(t.Context(), `DELETE FROM audit_events`)
	require.ErrorContains(t, err, "append-only")
	require.Equal(t, 1, count(t, pool, `SELECT count(*) FROM audit_events WHERE action = 'a'`))
}

func TestRecord_InvalidRequestIDReplaced(t *testing.T) {
	pool := testkit.MigratedDB(t)
	actor := testkit.SignIn(t, pool).ID
	var recordErr error
	h := httpx.RequestID(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		ctx := auth.WithPrincipal(r.Context(), auth.Principal{UserID: actor})
		recordErr = db.WithTx(ctx, pool, func(tx pgx.Tx) error {
			return audit.Record(ctx, tx, audit.Event{Action: "thing.changed", ResourceType: "thing", ResourceID: "t-24"})
		})
	}))
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/", nil)
	req.Header.Set(httpx.HeaderRequestID, "a b")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.NoError(t, recordErr)

	sent := rec.Header().Get(httpx.HeaderRequestID)
	_, err := uuid.Parse(sent)
	require.NoError(t, err, sent)
	var stored string
	require.NoError(t, pool.QueryRow(t.Context(), `SELECT request_id FROM audit_events WHERE resource_id = 't-24'`).Scan(&stored))
	require.Equal(t, sent, stored)
}
