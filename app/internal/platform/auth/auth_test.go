package auth_test

import (
	"context"
	"encoding/base64"
	"net/http"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"

	"github.com/luiszkm/template-go/internal/platform/auth"
	"github.com/luiszkm/template-go/internal/platform/op"
	"github.com/luiszkm/template-go/internal/platform/testkit"
)

const ttl = time.Hour

type empty struct{}

type ok struct {
	Body struct {
		OK bool `json:"ok"`
	}
}

func server(t *testing.T, pool *pgxpool.Pool) (http.Handler, *atomic.Int32) {
	t.Helper()
	api, h, _ := testkit.NewAPI(t, pool, testkit.APIOptions{SessionTTL: ttl})
	var calls atomic.Int32
	handler := func(context.Context, *empty) (*ok, error) { calls.Add(1); return &ok{}, nil }
	require.NoError(t, op.Register(api, op.Spec{ID: "public", Method: http.MethodGet, Path: "/public", Public: true}, handler))
	require.NoError(t, op.Register(api, op.Spec{ID: "authn", Method: http.MethodGet, Path: "/authn", Authenticated: true}, handler))
	require.NoError(t, op.Register(api, op.Spec{ID: "perm", Method: http.MethodGet, Path: "/perm", Permission: "x:y"}, handler))
	return h, &calls
}

func get(t *testing.T, h http.Handler, path string, c *http.Cookie) int {
	t.Helper()
	return testkit.Do(t, h, testkit.Request{Method: http.MethodGet, Path: path, Cookie: c}).Code
}

func TestNewToken_32RandomBytes(t *testing.T) {
	seen := map[string]bool{}
	for range 1000 {
		tok, err := auth.NewToken()
		require.NoError(t, err)
		require.Len(t, tok, 43)
		raw, err := base64.RawURLEncoding.DecodeString(tok)
		require.NoError(t, err)
		require.Len(t, raw, 32)
		require.False(t, seen[tok], "duplicate token")
		seen[tok] = true
	}
}

func TestMiddleware_Rejects401(t *testing.T) {
	pool := testkit.MigratedDB(t)
	h, calls := server(t, pool)

	expired := testkit.SignIn(t, pool, "x:y")
	_, err := pool.Exec(t.Context(), `UPDATE sessions SET created_at = now() - $1::interval WHERE user_id = $2`, (ttl + time.Second).String(), expired.ID)
	require.NoError(t, err)
	deactivated := testkit.SignIn(t, pool, "x:y")
	_, err = pool.Exec(t.Context(), `UPDATE users SET deactivated_at = now() WHERE id = $1`, deactivated.ID)
	require.NoError(t, err)

	cases := map[string]*http.Cookie{
		"no cookie":     nil,
		"unknown token": {Name: auth.CookieName, Value: "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"},
		"expired":       expired.Cookie,
		"deactivated":   deactivated.Cookie,
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			for _, path := range []string{"/authn", "/perm"} {
				rec := testkit.Do(t, h, testkit.Request{Method: http.MethodGet, Path: path, Cookie: c})
				require.Equal(t, http.StatusUnauthorized, rec.Code, path)
				require.Equal(t, "application/problem+json", rec.Header().Get("Content-Type"))
			}
		})
	}
	require.Zero(t, calls.Load(), "handler must not run")

	fresh := testkit.SignIn(t, pool, "x:y")
	_, err = pool.Exec(t.Context(), `UPDATE sessions SET created_at = now() - $1::interval WHERE user_id = $2`, (ttl - time.Minute).String(), fresh.ID)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, get(t, h, "/perm", fresh.Cookie), "a session under the TTL passes")
}

func TestMiddleware_Forbids403(t *testing.T) {
	pool := testkit.MigratedDB(t)
	h, calls := server(t, pool)
	for name, u := range map[string]testkit.User{
		"no role":         testkit.SignIn(t, pool),
		"other perm only": testkit.SignIn(t, pool, "other:thing"),
	} {
		rec := testkit.Do(t, h, testkit.Request{Method: http.MethodGet, Path: "/perm", Cookie: u.Cookie})
		require.Equal(t, http.StatusForbidden, rec.Code, name)
		require.Equal(t, "application/problem+json", rec.Header().Get("Content-Type"), name)
	}
	require.Zero(t, calls.Load(), "handler must not run")
}

func TestMiddleware_WildcardAndExactAllow(t *testing.T) {
	pool := testkit.MigratedDB(t)
	h, calls := server(t, pool)
	require.Equal(t, http.StatusOK, get(t, h, "/perm", testkit.SignIn(t, pool, auth.Wildcard).Cookie))
	require.Equal(t, http.StatusOK, get(t, h, "/perm", testkit.SignIn(t, pool, "x:y").Cookie))
	require.EqualValues(t, 2, calls.Load())
}

func TestMiddleware_AuthenticatedAndPublic(t *testing.T) {
	pool := testkit.MigratedDB(t)
	h, calls := server(t, pool)
	require.Equal(t, http.StatusOK, get(t, h, "/authn", testkit.SignIn(t, pool).Cookie))
	require.Equal(t, http.StatusOK, get(t, h, "/public", nil))
	require.EqualValues(t, 2, calls.Load())
}
