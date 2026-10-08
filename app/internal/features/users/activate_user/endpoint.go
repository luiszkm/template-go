package activateuser

import (
	"context"
	"errors"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/luiszkm/template-go/internal/features/users/activate_user/db"
	"github.com/luiszkm/template-go/internal/platform/audit"
	platformdb "github.com/luiszkm/template-go/internal/platform/db"
	"github.com/luiszkm/template-go/internal/platform/deps"
	"github.com/luiszkm/template-go/internal/platform/op"
)

const permission op.Permission = "users:activate"

type Input struct {
	ID uuid.UUID `path:"id"`
}

type state struct {
	Email  string `json:"email"`
	Active bool   `json:"active"`
}

var errNotFound = errors.New("not found")

func Register(api huma.API, d deps.Deps) error {
	return op.Register(api, op.Spec{
		ID:            "activate-user",
		Method:        http.MethodPost,
		Path:          "/api/v1/users/{id}/activate",
		Summary:       "Reactivate a deactivated user",
		Tags:          []string{"users"},
		Errors:        []int{http.StatusNotFound, http.StatusUnprocessableEntity},
		DefaultStatus: http.StatusNoContent,
		Permission:    permission,
		AuditAction:   "user.activated",
	}, func(ctx context.Context, in *Input) (*struct{}, error) {
		err := platformdb.WithTx(ctx, d.DB, func(tx pgx.Tx) error {
			q := db.New(tx)
			user, err := q.LockUser(ctx, in.ID)
			if errors.Is(err, pgx.ErrNoRows) {
				return errNotFound
			}
			if err != nil || user.DeactivatedAt == nil {
				return err
			}
			if err := q.Activate(ctx, user.ID); err != nil {
				return err
			}
			return audit.Record(ctx, tx, audit.Event{Action: "user.activated", ResourceType: "user", ResourceID: user.ID.String(),
				Before: state{Email: user.Email, Active: false}, After: state{Email: user.Email, Active: true}})
		})
		if errors.Is(err, errNotFound) {
			return nil, huma.Error404NotFound("user not found")
		}
		return nil, err
	})
}
