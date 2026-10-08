package logout

import (
	"context"
	"encoding/hex"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/jackc/pgx/v5"

	"github.com/luiszkm/template-go/internal/features/users/logout/db"
	"github.com/luiszkm/template-go/internal/platform/audit"
	"github.com/luiszkm/template-go/internal/platform/auth"
	platformdb "github.com/luiszkm/template-go/internal/platform/db"
	"github.com/luiszkm/template-go/internal/platform/deps"
	"github.com/luiszkm/template-go/internal/platform/op"
)

type Input struct{}

type Output struct {
	SetCookie http.Cookie `header:"Set-Cookie"`
}

func Register(api huma.API, d deps.Deps) error {
	return op.Register(api, op.Spec{
		ID:            "logout",
		Method:        http.MethodDelete,
		Path:          "/api/v1/users/session",
		Summary:       "Sign out: end the current session",
		Tags:          []string{"users"},
		DefaultStatus: http.StatusNoContent,
		Authenticated: true,
		AuditAction:   "session.deleted",
	}, func(ctx context.Context, _ *Input) (*Output, error) {
		principal, _ := auth.PrincipalFrom(ctx)
		err := platformdb.WithTx(ctx, d.DB, func(tx pgx.Tx) error {
			if _, err := db.New(tx).DeleteSession(ctx, principal.SessionHash); err != nil {
				return err
			}
			return audit.Record(ctx, tx, audit.Event{
				Action: "session.deleted", ResourceType: "session", ResourceID: hex.EncodeToString(principal.SessionHash),
				Before: map[string]string{"user_id": principal.UserID.String()},
			})
		})
		if err != nil {
			return nil, err
		}
		return &Output{SetCookie: auth.ClearCookie(d.CookieSecure)}, nil
	})
}
