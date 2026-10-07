// Command api runs the HTTP server and its maintenance subcommands.
//
//	api serve        start the HTTP server (never migrates)
//	api migrate up   apply pending migrations and exit
//	api openapi      print the OpenAPI document served at /api/openapi.json
package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/caarlos0/env/v11"
	_ "github.com/jackc/pgx/v5/stdlib" // database/sql driver "pgx" for goose
	"github.com/pressly/goose/v3"

	"github.com/luiszkm/template-go/internal/app"
	"github.com/luiszkm/template-go/internal/platform/config"
	"github.com/luiszkm/template-go/internal/platform/db"
	"github.com/luiszkm/template-go/migrations"
)

const usage = "usage: api serve | api migrate up | api openapi"

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	os.Exit(run(ctx, os.Args[1:], env.ToMap(os.Environ()), os.Stdout, os.Stderr))
}

func run(ctx context.Context, args []string, environ map[string]string, stdout, stderr io.Writer) int {
	cmd := strings.Join(args, " ")
	if cmd == "openapi" {
		return exportOpenAPI(stdout, stderr)
	}
	if cmd != "serve" && cmd != "migrate up" {
		fmt.Fprintln(stderr, usage)
		return 2
	}

	cfg, err := config.Load(environ)
	if err != nil {
		fmt.Fprintf(stderr, "config: %v\n", err)
		return 1
	}
	log := newLogger(stdout, cfg.LogLevel)

	if cmd == "migrate up" {
		err = migrateUp(ctx, cfg.DatabaseURL)
	} else {
		err = serve(ctx, cfg, log)
	}
	if err != nil {
		log.Error(cmd+" failed", slog.String("error", err.Error()))
		fmt.Fprintf(stderr, "%s: %v\n", cmd, err)
		return 1
	}
	return 0
}

func newLogger(w io.Writer, level string) *slog.Logger {
	var l slog.Level
	if err := l.UnmarshalText([]byte(level)); err != nil {
		l = slog.LevelInfo
	}
	return slog.New(slog.NewJSONHandler(w, &slog.HandlerOptions{Level: l}))
}

func serve(ctx context.Context, cfg config.Config, log *slog.Logger) error {
	pool, err := db.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	h, err := app.New(app.Options{Logger: log, DB: pool, ReadyTimeout: cfg.ReadyTimeout})
	if err != nil {
		return err
	}
	ln, err := net.Listen("tcp", cfg.HTTPAddr)
	if err != nil {
		return err
	}
	srv := &http.Server{Handler: h, ReadHeaderTimeout: 10 * time.Second}
	log.Info("listening", slog.String("addr", ln.Addr().String()))
	return app.Run(ctx, srv, ln, cfg.ShutdownTimeout, log)
}

func migrateUp(ctx context.Context, url string) error {
	sqlDB, err := sql.Open("pgx", url)
	if err != nil {
		return err
	}
	defer func() { _ = sqlDB.Close() }()
	provider, err := goose.NewProvider(goose.DialectPostgres, sqlDB, migrations.FS)
	if err != nil {
		return err
	}
	_, err = provider.Up(ctx)
	return err
}

func exportOpenAPI(stdout, stderr io.Writer) int {
	h, err := app.New(app.Options{})
	if err == nil {
		var doc []byte
		if doc, err = app.OpenAPI(h); err == nil {
			_, err = stdout.Write(doc)
		}
	}
	if err != nil {
		fmt.Fprintln(stderr, errors.Join(errors.New("openapi"), err))
		return 1
	}
	return 0
}
