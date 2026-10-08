package login

import (
	"context"
	"encoding/hex"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/danielgtaylor/huma/v2"
	"github.com/jackc/pgx/v5"

	"github.com/luiszkm/template-go/internal/features/users/email"
	"github.com/luiszkm/template-go/internal/features/users/login/db"
	"github.com/luiszkm/template-go/internal/features/users/password"
	"github.com/luiszkm/template-go/internal/platform/audit"
	"github.com/luiszkm/template-go/internal/platform/auth"
	platformdb "github.com/luiszkm/template-go/internal/platform/db"
	"github.com/luiszkm/template-go/internal/platform/deps"
	"github.com/luiszkm/template-go/internal/platform/httpx"
	"github.com/luiszkm/template-go/internal/platform/op"
)

const (
	emailLimit = 5
	ipLimit    = 20
)

const invalidCredentials = "invalid email or password"

type Credentials struct {
	Email    string `json:"email" maxLength:"254"`
	Password string `json:"password" minLength:"1" maxLength:"128"`
}

func (b *Credentials) Resolve(_ huma.Context, prefix *huma.PathBuffer) []error {
	b.Email = email.Normalize(b.Email)
	if !email.Valid(b.Email) {
		return []error{&huma.ErrorDetail{Location: prefix.With("email"), Message: "expected an email address", Value: b.Email}}
	}
	return nil
}

type Input struct {
	Session string `cookie:"session"`
	Body    Credentials
}

type Output struct {
	SetCookie http.Cookie `header:"Set-Cookie"`
}

func Register(api huma.API, d deps.Deps) error {
	return RegisterWithVerifier(api, d, password.Verify)
}

func RegisterWithVerifier(api huma.API, d deps.Deps, verify password.Verifier) error {
	log := d.Logger
	if log == nil {
		log = slog.New(slog.DiscardHandler)
	}
	return op.Register(api, op.Spec{
		ID:            "login",
		Method:        http.MethodPost,
		Path:          "/api/v1/users/session",
		Summary:       "Sign in: exchange email and password for a session cookie",
		Tags:          []string{"users"},
		Errors:        []int{http.StatusUnauthorized, http.StatusUnprocessableEntity, http.StatusTooManyRequests},
		DefaultStatus: http.StatusNoContent,
		Public:        true,
		AuditAction:   "session.created",
	}, func(ctx context.Context, in *Input) (*Output, error) {
		q := db.New(d.DB)
		ip := clientKey(ctx)

		for _, lim := range []struct {
			kind, key string
			n         int32
		}{{"email", in.Body.Email, emailLimit}, {"ip", ip, ipLimit}} {
			secs, err := q.RetryAfter(ctx, db.RetryAfterParams{Kind: lim.kind, Key: lim.key, LimitMinusOne: lim.n - 1})
			if err == nil {
				return nil, huma.ErrorWithHeaders(huma.Error429TooManyRequests("too many failed logins; try again later"),
					http.Header{"Retry-After": {strconv.Itoa(max(1, int(secs)))}})
			}
			if !errors.Is(err, pgx.ErrNoRows) {
				return nil, err
			}
		}

		user, err := q.UserByEmail(ctx, in.Body.Email)
		found := err == nil
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return nil, err
		}
		hash := password.DummyHash()
		if found {
			hash = user.PasswordHash
		}
		match, err := verify(in.Body.Password, hash)
		if err != nil {
			return nil, err
		}
		if !found || !match || user.DeactivatedAt != nil {
			log.LogAttrs(ctx, slog.LevelWarn, "login failed",
				slog.String("request_id", httpx.RequestIDFrom(ctx)), slog.String("ip", ip))
			if err := recordFailure(ctx, q, in.Body.Email, ip); err != nil {
				return nil, err
			}
			return nil, huma.Error401Unauthorized(invalidCredentials)
		}

		var token string
		err = platformdb.WithTx(ctx, d.DB, func(tx pgx.Tx) error {
			qtx := db.New(tx)
			if password.NeedsRehash(user.PasswordHash) {
				rehashed, err := password.Hash(in.Body.Password)
				if err != nil {
					return err
				}
				if err := qtx.UpdatePasswordHash(ctx, db.UpdatePasswordHashParams{ID: user.ID, PasswordHash: rehashed}); err != nil {
					return err
				}
			}
			if err := qtx.ClearEmailAttempts(ctx, in.Body.Email); err != nil {
				return err
			}
			if in.Session != "" {
				if err := qtx.DeleteSession(ctx, auth.HashToken(in.Session)); err != nil {
					return err
				}
			}
			if token, err = auth.IssueSession(ctx, tx, user.ID); err != nil {
				return err
			}
			actx := auth.WithPrincipal(ctx, auth.Principal{UserID: user.ID})
			return audit.Record(actx, tx, audit.Event{
				Action: "session.created", ResourceType: "session", ResourceID: hex.EncodeToString(auth.HashToken(token)),
				After: map[string]string{"user_id": user.ID.String()},
			})
		})
		if err != nil {
			return nil, err
		}
		return &Output{SetCookie: auth.SessionCookie(token, d.SessionTTL, d.CookieSecure)}, nil
	})
}

func recordFailure(ctx context.Context, q *db.Queries, emailKey, ip string) error {
	if err := q.PruneAttempts(ctx); err != nil {
		return err
	}
	if err := q.RecordAttempt(ctx, db.RecordAttemptParams{Kind: "email", Key: emailKey}); err != nil {
		return err
	}
	return q.RecordAttempt(ctx, db.RecordAttemptParams{Kind: "ip", Key: ip})
}

func clientKey(ctx context.Context) string {
	if addr := httpx.ClientIP(ctx); addr.IsValid() {
		return addr.String()
	}
	return "unknown"
}
