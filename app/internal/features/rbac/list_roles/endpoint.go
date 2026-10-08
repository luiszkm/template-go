package listroles

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	"github.com/luiszkm/template-go/internal/features/rbac/list_roles/db"
	"github.com/luiszkm/template-go/internal/features/rbac/role"
	"github.com/luiszkm/template-go/internal/platform/deps"
	"github.com/luiszkm/template-go/internal/platform/op"
)

const permission op.Permission = "rbac:read"

type Input struct{}

type RoleList struct {
	Items []role.Role `json:"items"`
}

type Output struct {
	Body RoleList
}

func Register(api huma.API, d deps.Deps) error {
	return op.Register(api, op.Spec{
		ID:         "list-roles",
		Method:     http.MethodGet,
		Path:       "/api/v1/rbac/roles",
		Summary:    "Every role, ordered by name, with its permissions and how many users hold it",
		Tags:       []string{"rbac"},
		Permission: permission,
	}, func(ctx context.Context, _ *Input) (*Output, error) {
		rows, err := db.New(d.DB).ListRoles(ctx)
		if err != nil {
			return nil, err
		}
		items := make([]role.Role, 0, len(rows))
		for _, r := range rows {
			items = append(items, role.Role{ID: r.ID, Name: r.Name, Permissions: r.Permissions, UserCount: int(r.UserCount)})
		}
		return &Output{Body: RoleList{Items: items}}, nil
	})
}
