package main

import (
	"bufio"
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/caarlos0/env/v11"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	"github.com/luiszkm/template-go/internal/app"
	"github.com/luiszkm/template-go/internal/platform/auth"
	"github.com/luiszkm/template-go/internal/platform/config"
	"github.com/luiszkm/template-go/internal/platform/db"
	"github.com/luiszkm/template-go/migrations"
)

const usage = "usage: api serve | api migrate up | api openapi | api users create-admin --email <e> --name <n> < password"

const createAdminUsage = "usage: api users create-admin --email <e> --name <n> (password on stdin, 12-128 characters)"

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	os.Exit(run(ctx, os.Args[1:], env.ToMap(os.Environ()), os.Stdin, os.Stdout, os.Stderr))
}

func run(ctx context.Context, args []string, environ map[string]string, stdin io.Reader, stdout, stderr io.Writer) int {
	cmd := strings.Join(args, " ")
	if cmd == "openapi" {
		return exportOpenAPI(stdout, stderr)
	}
	if len(args) >= 2 && args[0] == "users" && args[1] == "create-admin" {
		return createAdmin(ctx, args[2:], environ, stdin, stdout, stderr)
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

	h, err := app.New(app.Options{Logger: log, DB: pool, ReadyTimeout: cfg.ReadyTimeout,
		SessionTTL: cfg.SessionTTL, CookieSecure: cfg.CookieSecure, TrustedProxies: cfg.TrustedProxies})
	if err != nil {
		return err
	}
	var lc net.ListenConfig
	ln, err := lc.Listen(ctx, "tcp", cfg.HTTPAddr)
	if err != nil {
		return err
	}
	srv := newServer(cfg, h)

	sweepCtx, stopSweep := context.WithCancel(ctx)
	var sweeping sync.WaitGroup
	sweeping.Go(func() { auth.SweepSessions(sweepCtx, pool, cfg.SessionTTL, cfg.SessionSweepInterval, log) })
	defer sweeping.Wait()
	defer stopSweep()

	log.Info("listening", slog.String("addr", ln.Addr().String()))
	return app.Run(ctx, srv, ln, cfg.ShutdownTimeout, log)
}

func newServer(cfg config.Config, h http.Handler) *http.Server {
	return &http.Server{
		Handler:           h,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       cfg.HTTPReadTimeout,
		WriteTimeout:      cfg.HTTPWriteTimeout,
		IdleTimeout:       cfg.HTTPIdleTimeout,
	}
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

func createAdmin(ctx context.Context, args []string, environ map[string]string, stdin io.Reader, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("create-admin", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	email := flags.String("email", "", "")
	name := flags.String("name", "", "")
	if err := flags.Parse(args); err != nil || *email == "" || *name == "" {
		fmt.Fprintln(stderr, createAdminUsage)
		return 2
	}
	password, err := bufio.NewReader(stdin).ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		fmt.Fprintln(stderr, createAdminUsage)
		return 2
	}
	password = strings.TrimRight(password, "\r\n")
	if err := app.ValidateAdmin(*email, *name, password); err != nil {
		fmt.Fprintf(stderr, "%s\n%v\n", createAdminUsage, err)
		return 2
	}

	cfg, err := config.Load(environ)
	if err != nil {
		fmt.Fprintf(stderr, "config: %v\n", err)
		return 1
	}
	pool, err := db.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		fmt.Fprintf(stderr, "create-admin: %v\n", err)
		return 1
	}
	defer pool.Close()
	id, err := app.CreateAdmin(ctx, pool, *email, *name, password)
	if err != nil {
		fmt.Fprintf(stderr, "create-admin: %v\n", err)
		return 1
	}
	fmt.Fprintln(stdout, id)
	return 0
}
