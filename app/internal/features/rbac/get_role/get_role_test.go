package getrole_test

import (
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	getrole "github.com/luiszkm/template-go/internal/features/rbac/get_role"
	"github.com/luiszkm/template-go/internal/features/rbac/rbactest"
	"github.com/luiszkm/template-go/internal/features/rbac/role"
	"github.com/luiszkm/template-go/internal/platform/testkit"
)

func TestGetRole_ReturnsRole(t *testing.T) {
	pool := testkit.MigratedDB(t)
	h := rbactest.Serve(t, pool, getrole.Register)
	caller := testkit.SignIn(t, pool, "rbac:read")
	beta := rbactest.CreateRole(t, pool, "Beta", "users:read", "users:create")
	rbactest.Assign(t, pool, rbactest.InsertUser(t, pool), beta)

	rec := testkit.Do(t, h, testkit.Request{Method: http.MethodGet, Path: "/api/v1/rbac/roles/" + beta.String(), Cookie: caller.Cookie})
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	require.Equal(t, role.Role{ID: beta, Name: "Beta", Permissions: []string{"users:create", "users:read"}, UserCount: 1},
		testkit.JSON[role.Role](t, rec))
}

func TestGetRole_404And422(t *testing.T) {
	pool := testkit.MigratedDB(t)
	h := rbactest.Serve(t, pool, getrole.Register)
	caller := testkit.SignIn(t, pool, "rbac:read")
	get := func(id string) int {
		return testkit.Do(t, h, testkit.Request{Method: http.MethodGet, Path: "/api/v1/rbac/roles/" + id, Cookie: caller.Cookie}).Code
	}
	require.Equal(t, http.StatusNotFound, get(uuid.NewString()))
	require.Equal(t, http.StatusUnprocessableEntity, get("abc"))
}
