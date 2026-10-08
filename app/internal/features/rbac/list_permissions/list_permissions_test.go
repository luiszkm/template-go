package listpermissions_test

import (
	"net/http"
	"slices"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/luiszkm/template-go/internal/features/rbac"
	"github.com/luiszkm/template-go/internal/features/rbac/rbactest"
	"github.com/luiszkm/template-go/internal/platform/testkit"
)

func TestListPermissions_ReturnsCatalogue(t *testing.T) {
	pool := testkit.MigratedDB(t)
	h := rbactest.Serve(t, pool, rbac.Register, rbactest.Probe)
	caller := testkit.SignIn(t, pool, "rbac:read")

	rec := testkit.Do(t, h, testkit.Request{Method: http.MethodGet, Path: "/api/v1/rbac/permissions", Cookie: caller.Cookie})
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	items := testkit.JSON[struct {
		Items []string `json:"items"`
	}](t, rec).Items
	require.True(t, slices.IsSorted(items), items)
	require.Contains(t, items, "rbac:read")
	require.Contains(t, items, "rbac:assign")
	require.Contains(t, items, "zeta:do")
	require.NotContains(t, items, "*")
}
