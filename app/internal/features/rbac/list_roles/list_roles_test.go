package listroles_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	listroles "github.com/luiszkm/template-go/internal/features/rbac/list_roles"
	"github.com/luiszkm/template-go/internal/features/rbac/rbactest"
	"github.com/luiszkm/template-go/internal/features/rbac/role"
	"github.com/luiszkm/template-go/internal/platform/testkit"
)

func TestListRoles_OrderedWithCounts(t *testing.T) {
	pool := testkit.MigratedDB(t)
	h := rbactest.Serve(t, pool, listroles.Register)
	caller := testkit.SignIn(t, pool, "rbac:read")
	rbactest.Assign(t, pool, rbactest.InsertUser(t, pool), rbactest.AdminRoleID(t, pool))
	rbactest.CreateRole(t, pool, "Zeta")
	beta := rbactest.CreateRole(t, pool, "Beta", "users:read", "users:create")
	rbactest.Assign(t, pool, rbactest.InsertUser(t, pool), beta)
	off := rbactest.InsertUser(t, pool)
	rbactest.Assign(t, pool, off, beta)
	rbactest.Deactivate(t, pool, off)

	rec := testkit.Do(t, h, testkit.Request{Method: http.MethodGet, Path: "/api/v1/rbac/roles", Cookie: caller.Cookie})
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	items := testkit.JSON[struct {
		Items []role.Role `json:"items"`
	}](t, rec).Items

	byName := map[string]role.Role{}
	var order []string
	for _, r := range items {
		byName[r.Name] = r
		if r.Name == "admin" || r.Name == "Beta" || r.Name == "Zeta" {
			order = append(order, r.Name)
		}
	}
	require.Equal(t, []string{"admin", "Beta", "Zeta"}, order)
	require.Equal(t, []string{"*"}, byName["admin"].Permissions)
	require.Equal(t, 1, byName["admin"].UserCount)
	require.Equal(t, []string{"users:create", "users:read"}, byName["Beta"].Permissions)
	require.Equal(t, 2, byName["Beta"].UserCount)
	require.Equal(t, []string{}, byName["Zeta"].Permissions)
	require.Equal(t, 0, byName["Zeta"].UserCount)
}
