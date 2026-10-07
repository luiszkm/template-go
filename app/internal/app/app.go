// Package app is the composition root: the only place, with cmd/, that imports features.
// cmd/api and the tests build the server through New, so both run the same assembly.
package app

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/luiszkm/template-go/internal/features"
	"github.com/luiszkm/template-go/internal/platform/deps"
	"github.com/luiszkm/template-go/internal/platform/health"
	"github.com/luiszkm/template-go/internal/platform/httpx"
	"github.com/luiszkm/template-go/internal/platform/webui"
)

// Options configures New. Zero values fall back to production defaults.
type Options struct {
	Logger       *slog.Logger
	DB           *pgxpool.Pool // nil: /readyz answers 503 (used by the OpenAPI export)
	ReadyTimeout time.Duration
	Web          fs.FS                           // nil: the embedded SPA build
	Register     func(huma.API, deps.Deps) error // nil: features.Register
}

// New assembles the whole HTTP server. It fails if any operation violates the op contract.
func New(o Options) (http.Handler, error) {
	if o.Logger == nil {
		o.Logger = slog.New(slog.NewJSONHandler(io.Discard, nil))
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

	mux := http.NewServeMux()
	api := httpx.NewAPI(mux)

	var db health.DB
	if o.DB != nil {
		db = o.DB
	}
	if err := health.Register(api, db, o.ReadyTimeout); err != nil {
		return nil, err
	}
	if err := o.Register(api, deps.Deps{DB: o.DB}); err != nil {
		return nil, fmt.Errorf("app: register features: %w", err)
	}

	mux.Handle("/api/", httpx.NotFound())
	mux.Handle("/", webui.Handler(o.Web))
	// Spans go to the global OpenTelemetry provider (no-op until one is installed).
	return otelhttp.NewHandler(httpx.Chain(mux, o.Logger), "http"), nil
}

// OpenAPI returns the contract exactly as GET /api/openapi.json serves it.
func OpenAPI(h http.Handler) ([]byte, error) {
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/openapi.json", nil))
	if rec.Code != http.StatusOK {
		return nil, fmt.Errorf("app: openapi export: status %d", rec.Code)
	}
	return rec.Body.Bytes(), nil
}

// Run serves on ln until ctx is cancelled, then stops accepting connections and waits for
// in-flight requests up to timeout. It returns nil on a clean or timed-out shutdown.
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
