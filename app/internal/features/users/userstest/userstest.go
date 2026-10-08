package userstest

import (
	"net/http"
	"strings"
	"testing"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/luiszkm/template-go/internal/features/users/password"
	"github.com/luiszkm/template-go/internal/platform/auth"
	"github.com/luiszkm/template-go/internal/platform/deps"
	"github.com/luiszkm/template-go/internal/platform/testkit"
)

type Register func(huma.API, deps.Deps) error

func Serve(t *testing.T, pool *pgxpool.Pool, o testkit.APIOptions, slices ...Register) http.Handler {
	t.Helper()
	api, h, d := testkit.NewAPI(t, pool, o)
	for _, register := range slices {
		if err := register(api, d); err != nil {
			t.Fatalf("userstest: register: %v", err)
		}
	}
	return h
}

func CreateUser(t *testing.T, pool *pgxpool.Pool, email, pw string) uuid.UUID {
	t.Helper()
	hash, err := password.Hash(pw)
	if err != nil {
		t.Fatalf("userstest: hash: %v", err)
	}
	return InsertUser(t, pool, email, hash)
}

func InsertUser(t *testing.T, pool *pgxpool.Pool, email, hash string) uuid.UUID {
	t.Helper()
	var id uuid.UUID
	err := pool.QueryRow(t.Context(), `INSERT INTO users (email, name, password_hash) VALUES ($1, $2, $3) RETURNING id`,
		email, strings.Split(email, "@")[0], hash).Scan(&id)
	if err != nil {
		t.Fatalf("userstest: insert user: %v", err)
	}
	return id
}

func Count(t *testing.T, pool *pgxpool.Pool, sql string, args ...any) int {
	t.Helper()
	var n int
	if err := pool.QueryRow(t.Context(), sql, args...).Scan(&n); err != nil {
		t.Fatalf("userstest: count: %v", err)
	}
	return n
}

func SessionCookie(t *testing.T, res *http.Response) *http.Cookie {
	t.Helper()
	for _, c := range res.Cookies() {
		if c.Name == auth.CookieName {
			return c
		}
	}
	t.Fatalf("userstest: no %s cookie in %v", auth.CookieName, res.Header.Values("Set-Cookie"))
	return nil
}
