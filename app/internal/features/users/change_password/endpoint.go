package changepassword

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/jackc/pgx/v5"

	"github.com/luiszkm/template-go/internal/features/users/change_password/db"
	"github.com/luiszkm/template-go/internal/features/users/password"
	"github.com/luiszkm/template-go/internal/platform/audit"
	"github.com/luiszkm/template-go/internal/platform/auth"
	platformdb "github.com/luiszkm/template-go/internal/platform/db"
	"github.com/luiszkm/template-go/internal/platform/deps"
	"github.com/luiszkm/template-go/internal/platform/op"
)

type PasswordChange struct {
	CurrentPassword string `json:"current_password" minLength:"1" maxLength:"128"`
	NewPassword     string `json:"new_password" minLength:"12" maxLength:"128"`
}

type Input struct {
	Body PasswordChange
}

type revocation struct {
	OtherSessionsEnded int64 `json:"other_sessions_ended"`
}

func Register(api huma.API, d deps.Deps) error {
	return op.Register(api, op.Spec{
		ID:            "change-own-password",
		Method:        http.MethodPut,
		Path:          "/api/v1/users/me/password",
		Summary:       "Change the password of the signed-in user and end their other sessions",
		Tags:          []string{"users"},
		Errors:        []int{http.StatusUnprocessableEntity},
		DefaultStatus: http.StatusNoContent,
		Authenticated: true,
		AuditAction:   "user.password_changed",
	}, func(ctx context.Context, in *Input) (*struct{}, error) {
		principal, _ := auth.PrincipalFrom(ctx)
		current, err := db.New(d.DB).PasswordHash(ctx, principal.UserID)
		if err != nil {
			return nil, err
		}
		match, err := password.Verify(in.Body.CurrentPassword, current)
		if err != nil {
			return nil, err
		}
		if !match {
			return nil, huma.Error422UnprocessableEntity("validation failed",
				&huma.ErrorDetail{Location: "body.current_password", Message: "incorrect password"})
		}
		hash, err := password.Hash(in.Body.NewPassword)
		if err != nil {
			return nil, err
		}
		return nil, platformdb.WithTx(ctx, d.DB, func(tx pgx.Tx) error {
			q := db.New(tx)
			if err := q.UpdatePasswordHash(ctx, db.UpdatePasswordHashParams{ID: principal.UserID, PasswordHash: hash}); err != nil {
				return err
			}
			ended, err := q.DeleteOtherSessions(ctx, db.DeleteOtherSessionsParams{UserID: principal.UserID, TokenHash: principal.SessionHash})
			if err != nil {
				return err
			}
			return audit.Record(ctx, tx, audit.Event{Action: "user.password_changed", ResourceType: "user",
				ResourceID: principal.UserID.String(), After: revocation{OtherSessionsEnded: ended}})
		})
	})
}
