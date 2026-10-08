package op_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/humatest"
	"github.com/stretchr/testify/require"

	"github.com/luiszkm/template-go/internal/platform/op"
)

type empty struct{}

func handler(context.Context, *empty) (*empty, error) { return &empty{}, nil }

func newAPI(t *testing.T) huma.API {
	t.Helper()
	_, api := humatest.New(t, huma.DefaultConfig("t", "1"))
	return api
}

func TestRegister_RejectsMissingPermission(t *testing.T) {
	err := op.Register(newAPI(t), op.Spec{ID: "list-things", Method: http.MethodGet, Path: "/things"}, handler)
	require.Error(t, err)
	require.Contains(t, err.Error(), "list-things")
}

func TestRegister_RejectsMutationWithoutAudit(t *testing.T) {
	for _, m := range []string{http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete} {
		t.Run(m, func(t *testing.T) {
			err := op.Register(newAPI(t), op.Spec{ID: "mutate-" + m, Method: m, Path: "/things", Permission: "things:write"}, handler)
			require.Error(t, err)
			require.Contains(t, err.Error(), "mutate-"+m)
		})
	}
}

func TestRegister_AcceptsValidSpecs(t *testing.T) {
	api := newAPI(t)
	require.NoError(t, op.Register(api, op.Spec{ID: "public-get", Method: http.MethodGet, Path: "/public", Public: true}, handler))
	require.NoError(t, op.Register(api, op.Spec{ID: "perm-get", Method: http.MethodGet, Path: "/perm", Permission: "things:read"}, handler))
	require.NoError(t, op.Register(api, op.Spec{ID: "perm-post", Method: http.MethodPost, Path: "/perm", Permission: "things:create", AuditAction: "thing.created"}, handler))
	require.NotNil(t, api.OpenAPI().Paths["/public"].Get)
	require.NotNil(t, api.OpenAPI().Paths["/perm"].Get)
	require.NotNil(t, api.OpenAPI().Paths["/perm"].Post)
}

func TestRegister_RejectsPermissionAndPublicTogether(t *testing.T) {
	err := op.Register(newAPI(t), op.Spec{ID: "both", Method: http.MethodGet, Path: "/x", Public: true, Permission: "x:read"}, handler)
	require.ErrorContains(t, err, "both")
}

func TestRegister_RejectsMissingID(t *testing.T) {
	err := op.Register(newAPI(t), op.Spec{Method: http.MethodGet, Path: "/things/{id}", Public: true}, handler)
	require.Error(t, err)
	require.Contains(t, err.Error(), http.MethodGet)
	require.Contains(t, err.Error(), "/things/{id}")
}

func TestRegister_ExactlyOneAccessMarker(t *testing.T) {
	cases := []struct {
		name          string
		permission    op.Permission
		public, authn bool
		ok            bool
	}{
		{"permission only", "x:read", false, false, true},
		{"public only", "", true, false, true},
		{"authenticated only", "", false, true, true},
		{"none", "", false, false, false},
		{"permission and public", "x:read", true, false, false},
		{"permission and authenticated", "x:read", false, true, false},
		{"public and authenticated", "", true, true, false},
		{"all three", "x:read", true, true, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			id := "op-" + c.name
			err := op.Register(newAPI(t), op.Spec{ID: id, Method: http.MethodGet, Path: "/x",
				Permission: c.permission, Public: c.public, Authenticated: c.authn}, handler)
			if c.ok {
				require.NoError(t, err)
				return
			}
			require.Error(t, err)
			require.Contains(t, err.Error(), id)
		})
	}
}

func TestPermissions_ListsRegisteredSortedUnique(t *testing.T) {
	api := newAPI(t)
	require.NoError(t, op.Register(api, op.Spec{ID: "b", Method: http.MethodGet, Path: "/b", Permission: "b:read"}, handler))
	require.NoError(t, op.Register(api, op.Spec{ID: "a", Method: http.MethodGet, Path: "/a", Permission: "a:read"}, handler))
	require.NoError(t, op.Register(api, op.Spec{ID: "a2", Method: http.MethodGet, Path: "/a2", Permission: "a:read"}, handler))
	require.NoError(t, op.Register(api, op.Spec{ID: "p", Method: http.MethodGet, Path: "/p", Public: true}, handler))
	require.Equal(t, []op.Permission{"a:read", "b:read"}, op.Permissions(api))
}
