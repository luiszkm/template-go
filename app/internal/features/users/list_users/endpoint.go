package listusers

import (
	"context"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"

	"github.com/luiszkm/template-go/internal/features/users/list_users/db"
	"github.com/luiszkm/template-go/internal/platform/deps"
	"github.com/luiszkm/template-go/internal/platform/op"
)

const permission op.Permission = "users:read"

type Input struct {
	Limit  int32 `query:"limit" default:"50" minimum:"1" maximum:"100"`
	Offset int32 `query:"offset" default:"0" minimum:"0"`
}

type ListedUser struct {
	ID            uuid.UUID  `json:"id"`
	Email         string     `json:"email"`
	Name          string     `json:"name"`
	Active        bool       `json:"active"`
	CreatedAt     time.Time  `json:"created_at"`
	DeactivatedAt *time.Time `json:"deactivated_at"`
}

type UserPage struct {
	Items []ListedUser `json:"items"`
	Total int64        `json:"total"`
}

type Output struct {
	Body UserPage
}

func Register(api huma.API, d deps.Deps) error {
	return op.Register(api, op.Spec{
		ID:         "list-users",
		Method:     http.MethodGet,
		Path:       "/api/v1/users",
		Summary:    "List users ordered by email",
		Tags:       []string{"users"},
		Errors:     []int{http.StatusUnprocessableEntity},
		Permission: permission,
	}, func(ctx context.Context, in *Input) (*Output, error) {
		q := db.New(d.DB)
		rows, err := q.ListUsers(ctx, db.ListUsersParams{Limit: in.Limit, Offset: in.Offset})
		if err != nil {
			return nil, err
		}
		total, err := q.CountUsers(ctx)
		if err != nil {
			return nil, err
		}
		page := UserPage{Items: make([]ListedUser, 0, len(rows)), Total: total}
		for _, r := range rows {
			page.Items = append(page.Items, ListedUser{ID: r.ID, Email: r.Email, Name: r.Name, Active: r.DeactivatedAt == nil,
				CreatedAt: r.CreatedAt, DeactivatedAt: r.DeactivatedAt})
		}
		return &Output{Body: page}, nil
	})
}
