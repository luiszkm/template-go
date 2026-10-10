package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/luiszkm/template-go/internal/platform/op"
)

const CookieName = "session"

const Wildcard op.Permission = "*"

type Principal struct {
	UserID      uuid.UUID
	SessionHash []byte
	Permissions map[op.Permission]struct{}
}

func (p Principal) Can(perm op.Permission) bool {
	if _, ok := p.Permissions[Wildcard]; ok {
		return true
	}
	_, ok := p.Permissions[perm]
	return ok
}

type principalKey struct{}

func PrincipalFrom(ctx context.Context) (Principal, bool) {
	p, ok := ctx.Value(principalKey{}).(Principal)
	return p, ok
}

func WithPrincipal(ctx context.Context, p Principal) context.Context {
	return context.WithValue(ctx, principalKey{}, p)
}

type Querier interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

func NewToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("auth: token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func HashToken(token string) []byte {
	sum := sha256.Sum256([]byte(token))
	return sum[:]
}

func IssueSession(ctx context.Context, q Querier, userID uuid.UUID) (string, error) {
	token, err := NewToken()
	if err != nil {
		return "", err
	}
	if _, err := q.Exec(ctx, `INSERT INTO sessions (token_hash, user_id) VALUES ($1, $2)`, HashToken(token), userID); err != nil {
		return "", fmt.Errorf("auth: issue session: %w", err)
	}
	return token, nil
}

func SessionCookie(token string, ttl time.Duration, secure bool) http.Cookie {
	return http.Cookie{ //nolint:gosec
		Name: CookieName, Value: token, Path: "/", MaxAge: int(ttl / time.Second),
		HttpOnly: true, SameSite: http.SameSiteLaxMode, Secure: secure,
	}
}

func ClearCookie(secure bool) http.Cookie {
	return http.Cookie{ //nolint:gosec
		Name: CookieName, Value: "", Path: "/", MaxAge: -1,
		HttpOnly: true, SameSite: http.SameSiteLaxMode, Secure: secure,
	}
}

var ErrNoSession = errors.New("auth: no valid session")

func Lookup(ctx context.Context, q Querier, token string, ttl time.Duration) (Principal, error) {
	hash := HashToken(token)
	var userID uuid.UUID
	var perms []string
	err := q.QueryRow(ctx, `
		SELECT s.user_id, coalesce(array_agg(DISTINCT rp.permission) FILTER (WHERE rp.permission IS NOT NULL), '{}')
		FROM sessions s
		JOIN users u ON u.id = s.user_id
		LEFT JOIN user_roles ur ON ur.user_id = s.user_id
		LEFT JOIN role_permissions rp ON rp.role_id = ur.role_id
		WHERE s.token_hash = $1 AND u.deactivated_at IS NULL AND s.created_at > now() - make_interval(secs => $2)
		GROUP BY s.user_id`,
		hash, ttl.Seconds()).Scan(&userID, &perms)
	if errors.Is(err, pgx.ErrNoRows) {
		return Principal{}, ErrNoSession
	}
	if err != nil {
		return Principal{}, fmt.Errorf("auth: lookup session: %w", err)
	}
	p := Principal{UserID: userID, SessionHash: hash, Permissions: make(map[op.Permission]struct{}, len(perms))}
	for _, perm := range perms {
		p.Permissions[op.Permission(perm)] = struct{}{}
	}
	return p, nil
}

func SweepSessions(ctx context.Context, q Querier, ttl, every time.Duration, log *slog.Logger) {
	ticker := time.NewTicker(every)
	defer ticker.Stop()
	for {
		if _, err := q.Exec(ctx, `DELETE FROM sessions WHERE created_at <= now() - make_interval(secs => $1)`, ttl.Seconds()); err != nil && ctx.Err() == nil {
			log.LogAttrs(ctx, slog.LevelWarn, "session sweep failed", slog.String("error", err.Error()))
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func Install(api huma.API, q Querier, ttl time.Duration) {
	api.UseMiddleware(func(ctx huma.Context, next func(huma.Context)) {
		meta := ctx.Operation().Metadata
		if public, _ := meta[op.MetaPublic].(bool); public {
			next(ctx)
			return
		}
		if q == nil {
			_ = huma.WriteErr(api, ctx, http.StatusServiceUnavailable, "database not configured")
			return
		}
		cookie, err := huma.ReadCookie(ctx, CookieName)
		if err != nil || cookie.Value == "" {
			_ = huma.WriteErr(api, ctx, http.StatusUnauthorized, "authentication required")
			return
		}
		p, err := Lookup(ctx.Context(), q, cookie.Value, ttl)
		if errors.Is(err, ErrNoSession) {
			_ = huma.WriteErr(api, ctx, http.StatusUnauthorized, "authentication required")
			return
		}
		if err != nil {
			_ = huma.WriteErr(api, ctx, http.StatusInternalServerError, "internal server error", err)
			return
		}
		if perm, _ := meta[op.MetaPermission].(op.Permission); perm != "" && !p.Can(perm) {
			_ = huma.WriteErr(api, ctx, http.StatusForbidden, "missing permission "+string(perm))
			return
		}
		next(huma.WithContext(ctx, WithPrincipal(ctx.Context(), p)))
	})
}
