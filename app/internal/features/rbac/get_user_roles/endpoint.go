package getuserroles

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"

	"github.com/luiszkm/template-go/internal/features/rbac/get_user_roles/db"
	"github.com/luiszkm/template-go/internal/platform/deps"
	"github.com/luiszkm/template-go/internal/platform/op"
)

const permission op.Permission = "rbac:read"

type Input struct {
	ID uuid.UUID `path:"id"`
}

type HeldRole struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

type UserRoleList struct {
	Items []HeldRole `json:"items"`
}

type Output struct {
	Body UserRoleList
}

func Register(api huma.API, d deps.Deps) error {
	return op.Register(api, op.Spec{
		ID:         "get-user-roles",
		Method:     http.MethodGet,
		Path:       "/api/v1/rbac/users/{id}/roles",
		Summary:    "The roles a user holds, ordered by name",
		Tags:       []string{"rbac"},
		Errors:     []int{http.StatusNotFound, http.StatusUnprocessableEntity},
		Permission: permission,
	}, func(ctx context.Context, in *Input) (*Output, error) {
		q := db.New(d.DB)
		exists, err := q.UserExists(ctx, in.ID)
		if err != nil {
			return nil, err
		}
		if !exists {
			return nil, huma.Error404NotFound("user not found")
		}
		rows, err := q.UserRoles(ctx, in.ID)
		if err != nil {
			return nil, err
		}
		items := make([]HeldRole, 0, len(rows))
		for _, r := range rows {
			items = append(items, HeldRole{ID: r.ID, Name: r.Name})
		}
		return &Output{Body: UserRoleList{Items: items}}, nil
	})
}
