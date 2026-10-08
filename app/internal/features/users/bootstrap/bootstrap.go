package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/luiszkm/template-go/internal/features/users/bootstrap/db"
	"github.com/luiszkm/template-go/internal/features/users/email"
	"github.com/luiszkm/template-go/internal/features/users/password"
	"github.com/luiszkm/template-go/internal/platform/audit"
	platformdb "github.com/luiszkm/template-go/internal/platform/db"
)

const maxNameLen = 100

var (
	ErrInvalid = errors.New("invalid admin")
	ErrExists  = errors.New("already exists")
)

type createdAdmin struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	Roles     []string  `json:"roles"`
	CreatedAt time.Time `json:"created_at"`
}

func Validate(rawEmail, name, pw string) error {
	switch {
	case !email.Valid(email.Normalize(rawEmail)):
		return fmt.Errorf("%w: email %q is not an address", ErrInvalid, rawEmail)
	case name == "" || utf8.RuneCountInString(name) > maxNameLen:
		return fmt.Errorf("%w: name must have 1 to %d characters", ErrInvalid, maxNameLen)
	case utf8.RuneCountInString(pw) < password.MinLen || utf8.RuneCountInString(pw) > password.MaxLen:
		return fmt.Errorf("%w: password must have %d to %d characters", ErrInvalid, password.MinLen, password.MaxLen)
	}
	return nil
}

func CreateAdmin(ctx context.Context, pool *pgxpool.Pool, rawEmail, name, pw string) (uuid.UUID, error) {
	if err := Validate(rawEmail, name, pw); err != nil {
		return uuid.Nil, err
	}
	normalized := email.Normalize(rawEmail)
	hash, err := password.Hash(pw)
	if err != nil {
		return uuid.Nil, err
	}
	var id uuid.UUID
	err = platformdb.WithTx(ctx, pool, func(tx pgx.Tx) error {
		q := db.New(tx)
		row, err := q.InsertUser(ctx, db.InsertUserParams{Email: normalized, Name: name, PasswordHash: hash})
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("user %s %w", normalized, ErrExists)
		}
		if err != nil {
			return err
		}
		if err := q.GrantAdmin(ctx, row.ID); err != nil {
			return err
		}
		id = row.ID
		return audit.Record(ctx, tx, audit.Event{Action: "user.created", ResourceType: "user", ResourceID: row.ID.String(),
			After: createdAdmin{ID: row.ID, Email: row.Email, Name: row.Name, Roles: []string{"admin"}, CreatedAt: row.CreatedAt}})
	})
	return id, err
}
