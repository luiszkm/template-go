package app_test

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/luiszkm/template-go/internal/app"
	"github.com/luiszkm/template-go/internal/features"
	"github.com/luiszkm/template-go/internal/platform/deps"
	"github.com/luiszkm/template-go/internal/platform/op"
	"github.com/luiszkm/template-go/internal/platform/testkit"
)

type route struct{ method, path string }

var someID = uuid.NewString()

var authenticatedRoutes = []route{
	{http.MethodDelete, "/api/v1/users/session"},
	{http.MethodGet, "/api/v1/users/me"},
	{http.MethodPut, "/api/v1/users/me/password"},
}

var permissionRoutes = []route{
	{http.MethodGet, "/api/v1/users"},
	{http.MethodPost, "/api/v1/users"},
	{http.MethodGet, "/api/v1/users/" + someID},
	{http.MethodPatch, "/api/v1/users/" + someID},
	{http.MethodPost, "/api/v1/users/" + someID + "/deactivate"},
	{http.MethodPost, "/api/v1/users/" + someID + "/activate"},
}

var bodies = map[route]any{
	{http.MethodPut, "/api/v1/users/me/password"}: map[string]string{"current_password": "x", "new_password": "senha-nova-1234"},
	{http.MethodPost, "/api/v1/users"}:            map[string]string{"email": "n@x.com", "name": "N", "password": "senha-nova-1234"},
	{http.MethodPatch, "/api/v1/users/" + someID}: map[string]string{"name": "N"},
}

func TestUsersRoutes_401WithoutSession(t *testing.T) {
	h, err := app.New(app.Options{Logger: testkit.DiscardLogger(), Web: site, DB: testkit.MigratedDB(t)})
	require.NoError(t, err)
	for _, r := range append(slices.Clone(authenticatedRoutes), permissionRoutes...) {
		rec := testkit.Do(t, h, testkit.Request{Method: r.method, Path: r.path, Body: bodies[r]})
		require.Equal(t, http.StatusUnauthorized, rec.Code, "%s %s", r.method, r.path)
		require.Equal(t, "application/problem+json", rec.Header().Get("Content-Type"))
	}
}

func TestUsersRoutes_403WithoutPermission(t *testing.T) {
	pool := testkit.MigratedDB(t)
	h, err := app.New(app.Options{Logger: testkit.DiscardLogger(), Web: site, DB: pool})
	require.NoError(t, err)
	noRole := testkit.SignIn(t, pool)
	for _, r := range permissionRoutes {
		rec := testkit.Do(t, h, testkit.Request{Method: r.method, Path: r.path, Body: bodies[r], Cookie: noRole.Cookie})
		require.Equal(t, http.StatusForbidden, rec.Code, "%s %s", r.method, r.path)
	}
}

func TestUsersOperations_AccessMarkers(t *testing.T) {
	var api huma.API
	_, err := app.New(app.Options{Logger: testkit.DiscardLogger(), Web: site,
		Register: func(a huma.API, d deps.Deps) error { api = a; return features.Register(a, d) }})
	require.NoError(t, err)

	want := map[string]string{
		"login":               "public",
		"logout":              "authenticated",
		"get-me":              "authenticated",
		"change-own-password": "authenticated",
		"list-users":          "users:read",
		"get-user":            "users:read",
		"create-user":         "users:create",
		"update-user":         "users:update",
		"deactivate-user":     "users:deactivate",
		"activate-user":       "users:activate",
	}
	got := map[string]string{}
	for _, item := range api.OpenAPI().Paths {
		for _, o := range []*huma.Operation{item.Get, item.Post, item.Put, item.Patch, item.Delete} {
			if o == nil || !strings.HasPrefix(o.Path, "/api/v1/users") {
				continue
			}
			switch {
			case o.Metadata[op.MetaPublic] == true:
				got[o.OperationID] = "public"
			case o.Metadata[op.MetaAuthenticated] == true:
				got[o.OperationID] = "authenticated"
			default:
				perm, _ := o.Metadata[op.MetaPermission].(op.Permission)
				got[o.OperationID] = string(perm)
			}
		}
	}
	require.Equal(t, want, got)
}

func TestOpenAPI_UsersStatuses(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "..", "openapi.json"))
	require.NoError(t, err)
	var doc struct {
		Paths map[string]map[string]struct {
			Responses map[string]any `json:"responses"`
		} `json:"paths"`
	}
	require.NoError(t, json.Unmarshal(raw, &doc))

	want := map[string][]string{
		"post /api/v1/users/session":         {"204", "401", "422", "429"},
		"delete /api/v1/users/session":       {"204", "401"},
		"get /api/v1/users/me":               {"200", "401"},
		"put /api/v1/users/me/password":      {"204", "401", "422"},
		"post /api/v1/users":                 {"201", "401", "403", "409", "422"},
		"get /api/v1/users":                  {"200", "401", "403", "422"},
		"get /api/v1/users/{id}":             {"200", "401", "403", "404", "422"},
		"patch /api/v1/users/{id}":           {"200", "401", "403", "404", "409", "422"},
		"post /api/v1/users/{id}/deactivate": {"204", "401", "403", "404", "409", "422"},
		"post /api/v1/users/{id}/activate":   {"204", "401", "403", "404", "422"},
	}
	for key, statuses := range want {
		method, path, _ := strings.Cut(key, " ")
		operation, ok := doc.Paths[path][method]
		require.True(t, ok, key)
		for _, status := range statuses {
			require.Contains(t, operation.Responses, status, key)
		}
	}
}
