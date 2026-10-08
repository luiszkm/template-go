package deps

import (
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Deps struct {
	DB     *pgxpool.Pool
	Logger *slog.Logger

	SessionTTL   time.Duration
	CookieSecure bool
}
