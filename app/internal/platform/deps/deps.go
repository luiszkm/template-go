// Package deps carries the shared dependencies handed to every feature at registration.
package deps

import (
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Deps is what a feature's Register function receives. Add a field only when a feature needs it.
type Deps struct {
	DB     *pgxpool.Pool
	Logger *slog.Logger

	SessionTTL   time.Duration
	CookieSecure bool
}
