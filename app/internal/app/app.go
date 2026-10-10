package app

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/luiszkm/template-go/internal/features"
	"github.com/luiszkm/template-go/internal/platform/auth"
	"github.com/luiszkm/template-go/internal/platform/deps"
	"github.com/luiszkm/template-go/internal/platform/health"
	"github.com/luiszkm/template-go/internal/platform/httpx"
	"github.com/luiszkm/template-go/internal/platform/webui"
)

type Options struct {
	Logger       *slog.Logger
	DB           *pgxpool.Pool
	ReadyTimeout time.Duration
	Web          fs.FS
	Register     func(huma.API, deps.Deps) error

	SessionTTL     time.Duration
	CookieSecure   bool
	TrustedProxies []netip.Prefix
}

func New(o Options) (http.Handler, error) {
	if o.Logger == nil {
		o.Logger = slog.New(slog.DiscardHandler)
	}
	if o.ReadyTimeout == 0 {
		o.ReadyTimeout = 2 * time.Second
	}
	if o.Web == nil {
		o.Web = webui.Dist()
	}
	if o.Register == nil {
		o.Register = features.Register
	}
	if o.SessionTTL == 0 {
		o.SessionTTL = 12 * time.Hour
	}

	mux := http.NewServeMux()
	api := httpx.NewAPI(mux)

	var db health.DB
	var sessions auth.Querier
	if o.DB != nil {
		db = o.DB
		sessions = o.DB
	}
	auth.Install(api, sessions, o.SessionTTL)
	if err := health.Register(api, db, o.ReadyTimeout); err != nil {
		return nil, err
	}
	d := deps.Deps{DB: o.DB, Logger: o.Logger, SessionTTL: o.SessionTTL, CookieSecure: o.CookieSecure}
	if err := o.Register(api, d); err != nil {
		return nil, fmt.Errorf("app: register features: %w", err)
	}

	mux.Handle("/api/", httpx.NotFound())
	mux.Handle("/", webui.Handler(o.Web))
	return httpx.Wrap(mux, httpx.Options{Logger: o.Logger, TrustedProxies: o.TrustedProxies, HSTS: o.CookieSecure}), nil
}

func OpenAPI(h http.Handler) ([]byte, error) {
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/openapi.json", nil))
	if rec.Code != http.StatusOK {
		return nil, fmt.Errorf("app: openapi export: status %d", rec.Code)
	}
	return rec.Body.Bytes(), nil
}

func Run(ctx context.Context, srv *http.Server, ln net.Listener, timeout time.Duration, log *slog.Logger) error {
	errc := make(chan error, 1)
	go func() { errc <- srv.Serve(ln) }()

	select {
	case err := <-errc:
		return err
	case <-ctx.Done():
	}

	sctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	err := srv.Shutdown(sctx)
	if errors.Is(err, context.DeadlineExceeded) {
		log.Warn("shutdown timeout reached; closing remaining connections", slog.Duration("timeout", timeout))
		_ = srv.Close()
		err = nil
	}
	if serveErr := <-errc; serveErr != nil && !errors.Is(serveErr, http.ErrServerClosed) {
		return serveErr
	}
	return err
}
