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
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/luiszkm/template-go/internal/app"
	"github.com/luiszkm/template-go/internal/features"
	"github.com/luiszkm/template-go/internal/platform/deps"
	"github.com/luiszkm/template-go/internal/platform/op"
	"github.com/luiszkm/template-go/internal/platform/testkit"
)

type rbacOperation struct {
	id         string
	method     string
	path       string
	permission op.Permission
	body       any
}

func rbacOperations(roleID, userID string) []rbacOperation {
	return []rbacOperation{
		{"list-permissions", http.MethodGet, "/api/v1/rbac/permissions", "rbac:read", nil},
		{"list-roles", http.MethodGet, "/api/v1/rbac/roles", "rbac:read", nil},
		{"get-role", http.MethodGet, "/api/v1/rbac/roles/" + roleID, "rbac:read", nil},
		{"create-role", http.MethodPost, "/api/v1/rbac/roles", "rbac:create", map[string]any{"name": "Nova", "permissions": []string{}}},
		{"update-role", http.MethodPatch, "/api/v1/rbac/roles/" + roleID, "rbac:update", map[string]any{"name": "Renomeado"}},
		{"delete-role", http.MethodDelete, "/api/v1/rbac/roles/" + roleID, "rbac:delete", nil},
		{"get-user-roles", http.MethodGet, "/api/v1/rbac/users/" + userID + "/roles", "rbac:read", nil},
		{"assign-roles", http.MethodPut, "/api/v1/rbac/users/" + userID + "/roles", "rbac:assign", map[string]any{"role_ids": []string{roleID}}},
	}
}

var rbacPermissions = []op.Permission{"rbac:read", "rbac:create", "rbac:update", "rbac:delete", "rbac:assign"}

func TestRBAC_OperationAccess(t *testing.T) {
	pool := testkit.MigratedDB(t)
	var api huma.API
	h, err := app.New(app.Options{Logger: testkit.DiscardLogger(), Web: site, DB: pool,
		Register: func(a huma.API, d deps.Deps) error { api = a; return features.Register(a, d) }})
	require.NoError(t, err)

	var roleID, userID uuid.UUID
	require.NoError(t, pool.QueryRow(t.Context(), `INSERT INTO roles (name) VALUES ('Alvo') RETURNING id`).Scan(&roleID))
	require.NoError(t, pool.QueryRow(t.Context(),
		`INSERT INTO users (email, name, password_hash) VALUES ('alvo@x.com', 'Alvo', 'x') RETURNING id`).Scan(&userID))

	declared := map[string]op.Permission{}
	for _, item := range api.OpenAPI().Paths {
		for _, o := range []*huma.Operation{item.Get, item.Post, item.Put, item.Patch, item.Delete} {
			if o != nil && strings.HasPrefix(o.Path, "/api/v1/rbac/") {
				declared[o.OperationID], _ = o.Metadata[op.MetaPermission].(op.Permission)
			}
		}
	}
	operations := rbacOperations(roleID.String(), userID.String())
	require.Len(t, declared, len(operations))

	count := func(sql string) int {
		var n int
		require.NoError(t, pool.QueryRow(t.Context(), sql).Scan(&n))
		return n
	}
	namedRoles := `SELECT count(*) FROM roles WHERE name NOT LIKE 'role-%'`
	targetRoles := `SELECT count(*) FROM user_roles WHERE role_id = (SELECT id FROM roles WHERE name = 'Alvo')`
	roles := count(namedRoles)

	for _, o := range operations {
		t.Run(o.id, func(t *testing.T) {
			require.Equal(t, o.permission, declared[o.id])

			rec := testkit.Do(t, h, testkit.Request{Method: o.method, Path: o.path, Body: o.body})
			require.Equal(t, http.StatusUnauthorized, rec.Code)

			var others []op.Permission
			for _, p := range rbacPermissions {
				if p != o.permission {
					others = append(others, p)
				}
			}
			caller := testkit.SignIn(t, pool, others...)
			rec = testkit.Do(t, h, testkit.Request{Method: o.method, Path: o.path, Body: o.body, Cookie: caller.Cookie})
			require.Equal(t, http.StatusForbidden, rec.Code, rec.Body.String())
			require.Zero(t, count(`SELECT count(*) FROM audit_events`))
			require.Equal(t, roles, count(namedRoles))
			require.Equal(t, 1, count(`SELECT count(*) FROM roles WHERE name = 'Alvo'`))
			require.Zero(t, count(targetRoles))
		})
	}
}

func TestOpenAPI_RBACStatuses(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "..", "openapi.json"))
	require.NoError(t, err)
	var doc struct {
		Paths map[string]map[string]struct {
			Responses map[string]any `json:"responses"`
		} `json:"paths"`
	}
	require.NoError(t, json.Unmarshal(raw, &doc))

	want := map[string][]string{
		"get /api/v1/rbac/permissions":      {"200", "401", "403"},
		"get /api/v1/rbac/roles":            {"200", "401", "403"},
		"post /api/v1/rbac/roles":           {"201", "401", "403", "409", "422"},
		"get /api/v1/rbac/roles/{id}":       {"200", "401", "403", "404", "422"},
		"patch /api/v1/rbac/roles/{id}":     {"200", "401", "403", "404", "409", "422"},
		"delete /api/v1/rbac/roles/{id}":    {"204", "401", "403", "404", "409", "422"},
		"get /api/v1/rbac/users/{id}/roles": {"200", "401", "403", "404", "422"},
		"put /api/v1/rbac/users/{id}/roles": {"204", "401", "403", "404", "409", "422"},
	}
	for key, statuses := range want {
		method, path, _ := strings.Cut(key, " ")
		operation, ok := doc.Paths[path][method]
		require.True(t, ok, key)
		documented := slices.Sorted(maps.Keys(operation.Responses))
		require.Equal(t, slices.Sorted(slices.Values(append(statuses, "500"))), documented, key)
	}
}
