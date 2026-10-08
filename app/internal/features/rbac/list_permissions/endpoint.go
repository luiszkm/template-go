package listpermissions

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	"github.com/luiszkm/template-go/internal/features/rbac/role"
	"github.com/luiszkm/template-go/internal/platform/deps"
	"github.com/luiszkm/template-go/internal/platform/op"
)

const permission op.Permission = "rbac:read"

type Input struct{}

type PermissionList struct {
	Items []string `json:"items"`
}

type Output struct {
	Body PermissionList
}

func Register(api huma.API, _ deps.Deps) error {
	return op.Register(api, op.Spec{
		ID:         "list-permissions",
		Method:     http.MethodGet,
		Path:       "/api/v1/rbac/permissions",
		Summary:    "Every permission an operation declares, which a role may grant",
		Tags:       []string{"rbac"},
		Permission: permission,
	}, func(context.Context, *Input) (*Output, error) {
		return &Output{Body: PermissionList{Items: role.Catalogue(api)}}, nil
	})
}
