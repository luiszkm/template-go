package updateuser

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/luiszkm/template-go/internal/features/users/email"
	"github.com/luiszkm/template-go/internal/features/users/update_user/db"
	"github.com/luiszkm/template-go/internal/platform/audit"
	platformdb "github.com/luiszkm/template-go/internal/platform/db"
	"github.com/luiszkm/template-go/internal/platform/deps"
	"github.com/luiszkm/template-go/internal/platform/op"
)

const permission op.Permission = "users:update"

const uniqueViolation = "23505"

type UserChanges struct {
	Email *string `json:"email,omitempty" maxLength:"254"`
	Name  *string `json:"name,omitempty" minLength:"1" maxLength:"100"`
}

func (b *UserChanges) Resolve(_ huma.Context, prefix *huma.PathBuffer) []error {
	if b.Email == nil && b.Name == nil {
		return []error{&huma.ErrorDetail{Location: prefix.String(), Message: "send name, email or both"}}
	}
	if b.Email != nil {
		normalized := email.Normalize(*b.Email)
		b.Email = &normalized
		if !email.Valid(normalized) {
			return []error{&huma.ErrorDetail{Location: prefix.With("email"), Message: "expected an email address", Value: normalized}}
		}
	}
	return nil
}

type Input struct {
	ID   uuid.UUID `path:"id"`
	Body UserChanges
}

type UpdatedUser struct {
	ID            uuid.UUID  `json:"id"`
	Email         string     `json:"email"`
	Name          string     `json:"name"`
	Active        bool       `json:"active"`
	CreatedAt     time.Time  `json:"created_at"`
	DeactivatedAt *time.Time `json:"deactivated_at"`
}

type Output struct {
	Body UpdatedUser
}

var (
	errNotFound   = errors.New("not found")
	errEmailTaken = errors.New("email taken")
)

func Register(api huma.API, d deps.Deps) error {
	return op.Register(api, op.Spec{
		ID:          "update-user",
		Method:      http.MethodPatch,
		Path:        "/api/v1/users/{id}",
		Summary:     "Change the name or email of a user",
		Tags:        []string{"users"},
		Errors:      []int{http.StatusNotFound, http.StatusConflict, http.StatusUnprocessableEntity},
		Permission:  permission,
		AuditAction: "user.updated",
	}, func(ctx context.Context, in *Input) (*Output, error) {
		var updated UpdatedUser
		err := platformdb.WithTx(ctx, d.DB, func(tx pgx.Tx) error {
			q := db.New(tx)
			current, err := q.LockUser(ctx, in.ID)
			if errors.Is(err, pgx.ErrNoRows) {
				return errNotFound
			}
			if err != nil {
				return err
			}
			before := UpdatedUser{ID: current.ID, Email: current.Email, Name: current.Name, Active: current.DeactivatedAt == nil,
				CreatedAt: current.CreatedAt, DeactivatedAt: current.DeactivatedAt}
			params := db.UpdateUserParams{ID: current.ID, Email: current.Email, Name: current.Name}
			if in.Body.Email != nil {
				params.Email = *in.Body.Email
			}
			if in.Body.Name != nil {
				params.Name = *in.Body.Name
			}
			row, err := q.UpdateUser(ctx, params)
			if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok && pgErr.Code == uniqueViolation {
				return errEmailTaken
			}
			if err != nil {
				return err
			}
			updated = UpdatedUser{ID: row.ID, Email: row.Email, Name: row.Name, Active: row.DeactivatedAt == nil,
				CreatedAt: row.CreatedAt, DeactivatedAt: row.DeactivatedAt}
			return audit.Record(ctx, tx, audit.Event{Action: "user.updated", ResourceType: "user", ResourceID: row.ID.String(),
				Before: before, After: updated})
		})
		switch {
		case errors.Is(err, errNotFound):
			return nil, huma.Error404NotFound("user not found")
		case errors.Is(err, errEmailTaken):
			return nil, huma.Error409Conflict("email already in use")
		case err != nil:
			return nil, err
		}
		return &Output{Body: updated}, nil
	})
}
