package getuser

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/luiszkm/template-go/internal/features/users/get_user/db"
	"github.com/luiszkm/template-go/internal/platform/deps"
	"github.com/luiszkm/template-go/internal/platform/op"
)

const permission op.Permission = "users:read"

type Input struct {
	ID uuid.UUID `path:"id"`
}

type UserDetail struct {
	ID            uuid.UUID  `json:"id"`
	Email         string     `json:"email"`
	Name          string     `json:"name"`
	Active        bool       `json:"active"`
	CreatedAt     time.Time  `json:"created_at"`
	DeactivatedAt *time.Time `json:"deactivated_at"`
}

type Output struct {
	Body UserDetail
}

func Register(api huma.API, d deps.Deps) error {
	return op.Register(api, op.Spec{
		ID:         "get-user",
		Method:     http.MethodGet,
		Path:       "/api/v1/users/{id}",
		Summary:    "Get a user",
		Tags:       []string{"users"},
		Errors:     []int{http.StatusNotFound, http.StatusUnprocessableEntity},
		Permission: permission,
	}, func(ctx context.Context, in *Input) (*Output, error) {
		r, err := db.New(d.DB).UserByID(ctx, in.ID)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, huma.Error404NotFound("user not found")
		}
		if err != nil {
			return nil, err
		}
		return &Output{Body: UserDetail{ID: r.ID, Email: r.Email, Name: r.Name, Active: r.DeactivatedAt == nil,
			CreatedAt: r.CreatedAt, DeactivatedAt: r.DeactivatedAt}}, nil
	})
}
