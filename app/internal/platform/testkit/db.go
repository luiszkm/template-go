package testkit

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/luiszkm/template-go/migrations"
)

var (
	sharedOnce sync.Once
	sharedURL  string
	sharedErr  error
)

func MigratedDB(t *testing.T) *pgxpool.Pool {
	t.Helper()
	sharedOnce.Do(startShared)
	if sharedErr != nil {
		t.Fatalf("testkit: %v", sharedErr)
	}
	name := "t_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	admin, err := sql.Open("pgx", sharedURL)
	if err != nil {
		t.Fatalf("testkit: %v", err)
	}
	defer func() { _ = admin.Close() }()
	if _, err := admin.ExecContext(t.Context(), "CREATE DATABASE "+name+" TEMPLATE app_template"); err != nil {
		t.Fatalf("testkit: create database: %v", err)
	}
	pool, err := pgxpool.New(context.Background(), strings.Replace(sharedURL, "/app?", "/"+name+"?", 1))
	if err != nil {
		t.Fatalf("testkit: pool: %v", err)
	}
	t.Cleanup(pool.Close)
	return pool
}

func URLOf(pool *pgxpool.Pool) string {
	return pool.Config().ConnString()
}

const dockerDesktopPipe = "npipe:////./pipe/docker_engine"

func pinDockerHost() error {
	if runtime.GOOS != "windows" || os.Getenv("DOCKER_HOST") != "" {
		return nil
	}
	return os.Setenv("DOCKER_HOST", dockerDesktopPipe)
}

func startShared() {
	ctx := context.Background()
	if sharedErr = pinDockerHost(); sharedErr != nil {
		return
	}
	var c *postgres.PostgresContainer
	var err error
	for attempt := range 6 {
		c, err = postgres.Run(ctx, "postgres:17-alpine",
			postgres.WithDatabase("app"), postgres.WithUsername("app"), postgres.WithPassword("app"),
			testcontainers.WithWaitStrategy(wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).WithStartupTimeout(120*time.Second)))
		if err == nil {
			break
		}
		time.Sleep(min(time.Duration(1<<attempt)*2*time.Second, 20*time.Second))
	}
	if err != nil {
		sharedErr = fmt.Errorf("start postgres: %w", err)
		return
	}
	if sharedURL, err = c.ConnectionString(ctx, "sslmode=disable"); err != nil {
		sharedErr = err
		return
	}
	admin, err := sql.Open("pgx", sharedURL)
	if err != nil {
		sharedErr = err
		return
	}
	defer func() { _ = admin.Close() }()
	if _, err := admin.ExecContext(ctx, "CREATE DATABASE app_template"); err != nil {
		sharedErr = err
		return
	}
	tmpl, err := sql.Open("pgx", strings.Replace(sharedURL, "/app?", "/app_template?", 1))
	if err != nil {
		sharedErr = err
		return
	}
	defer func() { _ = tmpl.Close() }()
	provider, err := goose.NewProvider(goose.DialectPostgres, tmpl, migrations.FS)
	if err != nil {
		sharedErr = err
		return
	}
	if _, err := provider.Up(ctx); err != nil {
		sharedErr = fmt.Errorf("migrate template: %w", err)
		return
	}
	sharedErr = tmpl.Close()
}
