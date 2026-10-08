package createuser

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/luiszkm/template-go/internal/features/users/create_user/db"
	"github.com/luiszkm/template-go/internal/features/users/email"
	"github.com/luiszkm/template-go/internal/features/users/password"
	"github.com/luiszkm/template-go/internal/platform/audit"
	platformdb "github.com/luiszkm/template-go/internal/platform/db"
	"github.com/luiszkm/template-go/internal/platform/deps"
	"github.com/luiszkm/template-go/internal/platform/op"
)

const permission op.Permission = "users:create"

type NewUser struct {
	Email    string `json:"email" maxLength:"254"`
	Name     string `json:"name" minLength:"1" maxLength:"100"`
	Password string `json:"password" minLength:"12" maxLength:"128"`
}

func (b *NewUser) Resolve(_ huma.Context, prefix *huma.PathBuffer) []error {
	b.Email = email.Normalize(b.Email)
	if !email.Valid(b.Email) {
		return []error{&huma.ErrorDetail{Location: prefix.With("email"), Message: "expected an email address", Value: b.Email}}
	}
	return nil
}

type Input struct {
	Body NewUser
}

type CreatedUser struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"created_at"`
}

type Output struct {
	Body CreatedUser
}

var errEmailTaken = errors.New("email taken")

func Register(api huma.API, d deps.Deps) error {
	return op.Register(api, op.Spec{
		ID:            "create-user",
		Method:        http.MethodPost,
		Path:          "/api/v1/users",
		Summary:       "Create a user",
		Tags:          []string{"users"},
		Errors:        []int{http.StatusConflict, http.StatusUnprocessableEntity},
		DefaultStatus: http.StatusCreated,
		Permission:    permission,
		AuditAction:   "user.created",
	}, func(ctx context.Context, in *Input) (*Output, error) {
		hash, err := password.Hash(in.Body.Password)
		if err != nil {
			return nil, err
		}
		var created CreatedUser
		err = platformdb.WithTx(ctx, d.DB, func(tx pgx.Tx) error {
			row, err := db.New(tx).InsertUser(ctx, db.InsertUserParams{Email: in.Body.Email, Name: in.Body.Name, PasswordHash: hash})
			if errors.Is(err, pgx.ErrNoRows) {
				return errEmailTaken
			}
			if err != nil {
				return err
			}
			created = CreatedUser{ID: row.ID, Email: row.Email, Name: row.Name, Active: true, CreatedAt: row.CreatedAt}
			return audit.Record(ctx, tx, audit.Event{Action: "user.created", ResourceType: "user", ResourceID: row.ID.String(), After: created})
		})
		if errors.Is(err, errEmailTaken) {
			return nil, huma.Error409Conflict("email already in use")
		}
		if err != nil {
			return nil, err
		}
		return &Output{Body: created}, nil
	})
}
