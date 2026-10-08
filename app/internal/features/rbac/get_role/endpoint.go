package getrole

import (
	"context"
	"errors"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/luiszkm/template-go/internal/features/rbac/get_role/db"
	"github.com/luiszkm/template-go/internal/features/rbac/role"
	"github.com/luiszkm/template-go/internal/platform/deps"
	"github.com/luiszkm/template-go/internal/platform/op"
)

const permission op.Permission = "rbac:read"

type Input struct {
	ID uuid.UUID `path:"id"`
}

type Output struct {
	Body role.Role
}

func Register(api huma.API, d deps.Deps) error {
	return op.Register(api, op.Spec{
		ID:         "get-role",
		Method:     http.MethodGet,
		Path:       "/api/v1/rbac/roles/{id}",
		Summary:    "One role with its permissions and how many users hold it",
		Tags:       []string{"rbac"},
		Errors:     []int{http.StatusNotFound, http.StatusUnprocessableEntity},
		Permission: permission,
	}, func(ctx context.Context, in *Input) (*Output, error) {
		r, err := db.New(d.DB).GetRole(ctx, in.ID)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, huma.Error404NotFound("role not found")
		}
		if err != nil {
			return nil, err
		}
		return &Output{Body: role.Role{ID: r.ID, Name: r.Name, Permissions: r.Permissions, UserCount: int(r.UserCount)}}, nil
	})
}
