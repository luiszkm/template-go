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

// C19
func TestRegister_RejectsMissingPermission(t *testing.T) {
	err := op.Register(newAPI(t), op.Spec{ID: "list-things", Method: http.MethodGet, Path: "/things"}, handler)
	require.Error(t, err)
	require.Contains(t, err.Error(), "list-things")
}

// C20
func TestRegister_RejectsMutationWithoutAudit(t *testing.T) {
	for _, m := range []string{http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete} {
		t.Run(m, func(t *testing.T) {
			err := op.Register(newAPI(t), op.Spec{ID: "mutate-" + m, Method: m, Path: "/things", Permission: "things:write"}, handler)
			require.Error(t, err)
			require.Contains(t, err.Error(), "mutate-"+m)
		})
	}
}

// C21
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

// C59
func TestRegister_RejectsMissingID(t *testing.T) {
	err := op.Register(newAPI(t), op.Spec{Method: http.MethodGet, Path: "/things/{id}", Public: true}, handler)
	require.Error(t, err)
	require.Contains(t, err.Error(), http.MethodGet)
	require.Contains(t, err.Error(), "/things/{id}")
}
