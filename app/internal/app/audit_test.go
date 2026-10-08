package app_test

import (
	"encoding/json"
	"maps"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/danielgtaylor/huma/v2"
	"github.com/stretchr/testify/require"

	"github.com/luiszkm/template-go/internal/app"
	"github.com/luiszkm/template-go/internal/features"
	"github.com/luiszkm/template-go/internal/platform/deps"
	"github.com/luiszkm/template-go/internal/platform/op"
	"github.com/luiszkm/template-go/internal/platform/testkit"
)

func assembledAPI(t *testing.T) huma.API {
	t.Helper()
	var api huma.API
	_, err := app.New(app.Options{Logger: testkit.DiscardLogger(), Web: site,
		Register: func(a huma.API, d deps.Deps) error { api = a; return features.Register(a, d) }})
	require.NoError(t, err)
	return api
}

func auditOperations(api huma.API) map[string]*huma.Operation {
	out := map[string]*huma.Operation{}
	for path, item := range api.OpenAPI().Paths {
		if !strings.HasPrefix(path, "/api/v1/audit") {
			continue
		}
		for method, o := range map[string]*huma.Operation{"GET": item.Get, "POST": item.Post, "PUT": item.Put, "PATCH": item.Patch, "DELETE": item.Delete} {
			if o != nil {
				out[method+" "+path] = o
			}
		}
	}
	return out
}

func TestAudit_ActionsOfAssembledServer(t *testing.T) {
	pool := testkit.MigratedDB(t)
	h, err := app.New(app.Options{Logger: testkit.DiscardLogger(), Web: site, DB: pool})
	require.NoError(t, err)
	caller := testkit.SignIn(t, pool, "audit:read")
	rec := testkit.Do(t, h, testkit.Request{Method: http.MethodGet, Path: "/api/v1/audit/actions", Cookie: caller.Cookie})
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	items := testkit.JSON[struct {
		Items []string `json:"items"`
	}](t, rec).Items
	for _, want := range []string{"user.created", "session.created", "role.created", "user.roles_changed"} {
		require.Contains(t, items, want)
	}
	require.True(t, slices.IsSorted(items))
}

func TestAudit_OperationAccess(t *testing.T) {
	pool := testkit.MigratedDB(t)
	h, err := app.New(app.Options{Logger: testkit.DiscardLogger(), Web: site, DB: pool})
	require.NoError(t, err)
	operations := auditOperations(assembledAPI(t))
	require.ElementsMatch(t, []string{"GET /api/v1/audit/events", "GET /api/v1/audit/events/{id}", "GET /api/v1/audit/actions"},
		slices.Collect(maps.Keys(operations)))

	var others []op.Permission
	for _, p := range op.Permissions(assembledAPI(t)) {
		if p != "audit:read" {
			others = append(others, p)
		}
	}
	caller := testkit.SignIn(t, pool, others...)
	paths := map[string]string{
		"GET /api/v1/audit/events":      "/api/v1/audit/events",
		"GET /api/v1/audit/events/{id}": "/api/v1/audit/events/1",
		"GET /api/v1/audit/actions":     "/api/v1/audit/actions",
	}
	for key, o := range operations {
		t.Run(key, func(t *testing.T) {
			require.Equal(t, op.Permission("audit:read"), o.Metadata[op.MetaPermission])
			require.Equal(t, http.StatusUnauthorized, testkit.Do(t, h, testkit.Request{Method: http.MethodGet, Path: paths[key]}).Code)
			require.Equal(t, http.StatusForbidden,
				testkit.Do(t, h, testkit.Request{Method: http.MethodGet, Path: paths[key], Cookie: caller.Cookie}).Code)
		})
	}
}

func TestAudit_ReadOnly(t *testing.T) {
	for key := range auditOperations(assembledAPI(t)) {
		method, _, _ := strings.Cut(key, " ")
		require.NotContains(t, []string{"POST", "PUT", "PATCH", "DELETE"}, method, key)
	}
}

func TestOpenAPI_AuditStatuses(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "..", "openapi.json"))
	require.NoError(t, err)
	var doc struct {
		Paths map[string]map[string]struct {
			Responses map[string]any `json:"responses"`
		} `json:"paths"`
	}
	require.NoError(t, json.Unmarshal(raw, &doc))
	want := map[string][]string{
		"get /api/v1/audit/events":      {"200", "401", "403", "422"},
		"get /api/v1/audit/events/{id}": {"200", "401", "403", "404", "422"},
		"get /api/v1/audit/actions":     {"200", "401", "403"},
	}
	for key, statuses := range want {
		method, path, _ := strings.Cut(key, " ")
		operation, ok := doc.Paths[path][method]
		require.True(t, ok, key)
		require.Equal(t, slices.Sorted(slices.Values(append(statuses, "500"))), slices.Sorted(maps.Keys(operation.Responses)), key)
	}
}
