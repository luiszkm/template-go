package app_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"

	"github.com/luiszkm/template-go/internal/app"
	"github.com/luiszkm/template-go/internal/platform/testkit"
)

const (
	spaPolicy = "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; " +
		"font-src 'self'; connect-src 'self'; object-src 'none'; base-uri 'self'; form-action 'self'; frame-ancestors 'none'"
	apiPolicy = "default-src 'none'; frame-ancestors 'none'"
)

type openAPIDoc struct {
	Paths map[string]map[string]struct {
		OperationID string         `json:"operationId"`
		Responses   map[string]any `json:"responses"`
	} `json:"paths"`
}

func serverWith(t *testing.T, pool *pgxpool.Pool, secure bool) http.Handler {
	t.Helper()
	h, err := app.New(app.Options{Logger: testkit.DiscardLogger(), Web: site, DB: pool, CookieSecure: secure})
	require.NoError(t, err)
	return h
}

func crossSiteLogin(t *testing.T, h http.Handler, fetchSite string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/api/v1/users/session",
		strings.NewReader(`{"email":"ana@x.com","password":"senha-longa-123"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Sec-Fetch-Site", fetchSite)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestCrossOrigin_RejectsCrossSiteMutation(t *testing.T) {
	pool := testkit.MigratedDB(t)
	h := serverWith(t, pool, false)
	for _, fetchSite := range []string{"cross-site", "same-site"} {
		rec := crossSiteLogin(t, h, fetchSite)
		require.Equal(t, http.StatusForbidden, rec.Code, fetchSite)
		require.Equal(t, "application/problem+json", rec.Header().Get("Content-Type"), fetchSite)
		require.Equal(t, "cross-origin request rejected", testkit.JSON[map[string]any](t, rec)["detail"], fetchSite)
	}
	var attempts int
	require.NoError(t, pool.QueryRow(t.Context(), `SELECT count(*) FROM login_attempts`).Scan(&attempts))
	require.Zero(t, attempts, "the login handler must not run")
}

func TestOpenAPI_MutationsDocument403(t *testing.T) {
	rec := get(newServer(t), "/api/openapi.json")
	require.Equal(t, http.StatusOK, rec.Code)
	doc := testkit.JSON[openAPIDoc](t, rec)
	mutations := 0
	for path, item := range doc.Paths {
		for method, o := range item {
			switch method {
			case "post", "put", "patch", "delete":
				mutations++
				require.Contains(t, o.Responses, "403", "%s %s (%s)", method, path, o.OperationID)
			}
		}
	}
	require.Positive(t, mutations)
}

func TestSecurityHeaders_OnEveryResponse(t *testing.T) {
	h := serverWith(t, testkit.MigratedDB(t), false)
	cases := []struct {
		name   string
		rec    *httptest.ResponseRecorder
		status int
	}{
		{"GET /", get(h, "/"), http.StatusOK},
		{"GET /api/does-not-exist", get(h, "/api/does-not-exist"), http.StatusNotFound},
		{"GET /healthz", get(h, "/healthz"), http.StatusOK},
		{"cross-site mutation", crossSiteLogin(t, h, "cross-site"), http.StatusForbidden},
	}
	for _, c := range cases {
		require.Equal(t, c.status, c.rec.Code, c.name)
		require.Equal(t, "nosniff", c.rec.Header().Get("X-Content-Type-Options"), c.name)
		require.Equal(t, "DENY", c.rec.Header().Get("X-Frame-Options"), c.name)
		require.Equal(t, "strict-origin-when-cross-origin", c.rec.Header().Get("Referrer-Policy"), c.name)
	}
}

func TestSecurityHeaders_SPAPolicy(t *testing.T) {
	h := newServer(t)
	for _, p := range []string{"/", "/users/123"} {
		rec := get(h, p)
		require.Equal(t, http.StatusOK, rec.Code, p)
		require.Equal(t, []string{spaPolicy}, rec.Header().Values("Content-Security-Policy"), p)
	}
}

func TestSecurityHeaders_APIPolicy(t *testing.T) {
	h := serverWith(t, testkit.MigratedDB(t), false)
	for _, p := range []string{"/api/does-not-exist", "/api/openapi.json", "/healthz", "/readyz"} {
		require.Equal(t, []string{apiPolicy}, get(h, p).Header().Values("Content-Security-Policy"), p)
	}
}

func TestSecurityHeaders_DocsKeepsHumaPolicy(t *testing.T) {
	rec := get(newServer(t), "/api/docs")
	require.Equal(t, http.StatusOK, rec.Code)
	policies := rec.Header().Values("Content-Security-Policy")
	require.Len(t, policies, 1)
	require.Contains(t, policies[0], "https://unpkg.com/")
}

func TestSecurityHeaders_HSTSFollowsCookieSecure(t *testing.T) {
	for _, p := range []string{"/", "/api/does-not-exist"} {
		require.Equal(t, "max-age=31536000", get(serverWith(t, nil, true), p).Header().Get("Strict-Transport-Security"), p)
		require.Empty(t, get(serverWith(t, nil, false), p).Header().Values("Strict-Transport-Security"), p)
	}
}
