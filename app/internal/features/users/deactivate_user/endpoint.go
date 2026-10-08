package deactivateuser

import (
	"context"
	"errors"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/luiszkm/template-go/internal/features/users/deactivate_user/db"
	"github.com/luiszkm/template-go/internal/platform/audit"
	"github.com/luiszkm/template-go/internal/platform/auth"
	platformdb "github.com/luiszkm/template-go/internal/platform/db"
	"github.com/luiszkm/template-go/internal/platform/deps"
	"github.com/luiszkm/template-go/internal/platform/op"
)

const permission op.Permission = "users:deactivate"

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
		ID:            "deactivate-user",
		Method:        http.MethodPost,
		Path:          "/api/v1/users/{id}/deactivate",
		Summary:       "Deactivate a user and end all of their sessions",
		Tags:          []string{"users"},
		Errors:        []int{http.StatusNotFound, http.StatusConflict, http.StatusUnprocessableEntity},
		DefaultStatus: http.StatusNoContent,
		Permission:    permission,
		AuditAction:   "user.deactivated",
	}, func(ctx context.Context, in *Input) (*struct{}, error) {
		if principal, _ := auth.PrincipalFrom(ctx); principal.UserID == in.ID {
			return nil, huma.Error409Conflict("a user cannot deactivate themselves")
		}
		err := platformdb.WithTx(ctx, d.DB, func(tx pgx.Tx) error {
			q := db.New(tx)
			user, err := q.LockUser(ctx, in.ID)
			if errors.Is(err, pgx.ErrNoRows) {
				return errNotFound
			}
			if err != nil || user.DeactivatedAt != nil {
				return err
			}
			if _, err := q.Deactivate(ctx, user.ID); err != nil {
				return err
			}
			if _, err := q.DeleteUserSessions(ctx, user.ID); err != nil {
				return err
			}
			return audit.Record(ctx, tx, audit.Event{Action: "user.deactivated", ResourceType: "user", ResourceID: user.ID.String(),
				Before: state{Email: user.Email, Active: true}, After: state{Email: user.Email, Active: false}})
		})
		if errors.Is(err, errNotFound) {
			return nil, huma.Error404NotFound("user not found")
		}
		return nil, err
	})
}
