package getuserroles_test

import (
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	getuserroles "github.com/luiszkm/template-go/internal/features/rbac/get_user_roles"
	"github.com/luiszkm/template-go/internal/features/rbac/rbactest"
	"github.com/luiszkm/template-go/internal/platform/testkit"
)

type held struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

func TestGetUserRoles_ReturnsRoles(t *testing.T) {
	pool := testkit.MigratedDB(t)
	h := rbactest.Serve(t, pool, getuserroles.Register)
	caller := testkit.SignIn(t, pool, "rbac:read")
	zeta := rbactest.CreateRole(t, pool, "Zeta")
	beta := rbactest.CreateRole(t, pool, "Beta")
	user := rbactest.InsertUser(t, pool)
	rbactest.Assign(t, pool, user, zeta, beta)
	bare := rbactest.InsertUser(t, pool)
	get := func(id uuid.UUID) []held {
		rec := testkit.Do(t, h, testkit.Request{Method: http.MethodGet, Path: "/api/v1/rbac/users/" + id.String() + "/roles", Cookie: caller.Cookie})
		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
		return testkit.JSON[struct {
			Items []held `json:"items"`
		}](t, rec).Items
	}

	require.Equal(t, []held{{beta, "Beta"}, {zeta, "Zeta"}}, get(user))
	require.Equal(t, []held{}, get(bare))
}

func TestGetUserRoles_404And422(t *testing.T) {
	pool := testkit.MigratedDB(t)
	h := rbactest.Serve(t, pool, getuserroles.Register)
	caller := testkit.SignIn(t, pool, "rbac:read")
	get := func(id string) int {
		return testkit.Do(t, h, testkit.Request{Method: http.MethodGet, Path: "/api/v1/rbac/users/" + id + "/roles", Cookie: caller.Cookie}).Code
	}
	require.Equal(t, http.StatusNotFound, get(uuid.NewString()))
	require.Equal(t, http.StatusUnprocessableEntity, get("abc"))
}
