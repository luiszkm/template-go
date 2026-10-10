package testkit

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"testing"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/luiszkm/template-go/internal/platform/auth"
	"github.com/luiszkm/template-go/internal/platform/deps"
	"github.com/luiszkm/template-go/internal/platform/httpx"
	"github.com/luiszkm/template-go/internal/platform/op"
)

type APIOptions struct {
	SessionTTL     time.Duration
	CookieSecure   bool
	TrustedProxies []netip.Prefix
	Logger         *slog.Logger
}

func NewAPI(t *testing.T, pool *pgxpool.Pool, o APIOptions) (huma.API, http.Handler, deps.Deps) {
	t.Helper()
	if o.SessionTTL == 0 {
		o.SessionTTL = 12 * time.Hour
	}
	if o.Logger == nil {
		o.Logger = DiscardLogger()
	}
	mux := http.NewServeMux()
	api := httpx.NewAPI(mux)
	auth.Install(api, pool, o.SessionTTL)
	d := deps.Deps{DB: pool, Logger: o.Logger, SessionTTL: o.SessionTTL, CookieSecure: o.CookieSecure}
	return api, httpx.Wrap(mux, httpx.Options{Logger: o.Logger, TrustedProxies: o.TrustedProxies, HSTS: o.CookieSecure}), d
}

func NewAPIWithoutDatabase(t *testing.T) (huma.API, http.Handler, deps.Deps) {
	t.Helper()
	mux := http.NewServeMux()
	api := httpx.NewAPI(mux)
	auth.Install(api, nil, 12*time.Hour)
	return api, httpx.Wrap(mux, httpx.Options{Logger: DiscardLogger()}), deps.Deps{Logger: DiscardLogger(), SessionTTL: 12 * time.Hour}
}

type User struct {
	ID     uuid.UUID
	Email  string
	Cookie *http.Cookie
}

func SignIn(t *testing.T, pool *pgxpool.Pool, perms ...op.Permission) User {
	t.Helper()
	ctx := t.Context()
	email := uuid.NewString() + "@test.local"
	var id uuid.UUID
	if err := pool.QueryRow(ctx, `INSERT INTO users (email, name, password_hash) VALUES ($1, 'Test', 'x') RETURNING id`, email).Scan(&id); err != nil {
		t.Fatalf("testkit: insert user: %v", err)
	}
	GrantPermissions(t, pool, id, perms...)
	token, err := auth.IssueSession(ctx, pool, id)
	if err != nil {
		t.Fatalf("testkit: %v", err)
	}
	return User{ID: id, Email: email, Cookie: &http.Cookie{Name: auth.CookieName, Value: token}} //nolint:gosec
}

func GrantPermissions(t *testing.T, pool *pgxpool.Pool, userID uuid.UUID, perms ...op.Permission) {
	t.Helper()
	if len(perms) == 0 {
		return
	}
	ctx := t.Context()
	var roleID uuid.UUID
	if err := pool.QueryRow(ctx, `INSERT INTO roles (name) VALUES ($1) RETURNING id`, "role-"+uuid.NewString()).Scan(&roleID); err != nil {
		t.Fatalf("testkit: insert role: %v", err)
	}
	for _, p := range perms {
		if _, err := pool.Exec(ctx, `INSERT INTO role_permissions (role_id, permission) VALUES ($1, $2)`, roleID, string(p)); err != nil {
			t.Fatalf("testkit: insert permission: %v", err)
		}
	}
	if _, err := pool.Exec(ctx, `INSERT INTO user_roles (user_id, role_id) VALUES ($1, $2)`, userID, roleID); err != nil {
		t.Fatalf("testkit: insert user role: %v", err)
	}
}

type Request struct {
	Method, Path string
	Body         any
	Cookie       *http.Cookie
	RemoteAddr   string
	Header       http.Header
}

func Do(t *testing.T, h http.Handler, r Request) *httptest.ResponseRecorder {
	t.Helper()
	var body bytes.Buffer
	if r.Body != nil {
		if err := json.NewEncoder(&body).Encode(r.Body); err != nil {
			t.Fatalf("testkit: encode body: %v", err)
		}
	}
	req := httptest.NewRequestWithContext(t.Context(), r.Method, r.Path, &body)
	if r.Body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for k, vs := range r.Header {
		for _, v := range vs {
			req.Header.Add(k, v)
		}
	}
	if r.Cookie != nil {
		req.AddCookie(r.Cookie)
	}
	if r.RemoteAddr != "" {
		req.RemoteAddr = r.RemoteAddr
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func JSON[T any](t *testing.T, rec *httptest.ResponseRecorder) T {
	t.Helper()
	var v T
	if err := json.Unmarshal(rec.Body.Bytes(), &v); err != nil {
		t.Fatalf("testkit: decode %q: %v", rec.Body.String(), err)
	}
	return v
}
