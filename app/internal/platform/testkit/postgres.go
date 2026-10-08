// Package testkit holds helpers shared by tests. Never import it from production code.
package testkit

import (
	"context"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// Postgres is a throwaway Postgres 17 container.
type Postgres struct {
	URL       string
	Container *postgres.PostgresContainer
}

// StartPostgres starts an empty database and terminates it when the test ends. Requires Docker.
func StartPostgres(t *testing.T) *Postgres {
	t.Helper()
	if err := pinDockerHost(); err != nil {
		t.Fatalf("testkit: %v", err)
	}
	ctx := context.Background()
	c, err := postgres.Run(ctx, "postgres:17-alpine",
		postgres.WithDatabase("app"),
		postgres.WithUsername("app"),
		postgres.WithPassword("app"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).WithStartupTimeout(60*time.Second)),
	)
	if err != nil {
		t.Fatalf("testkit: start postgres: %v", err)
	}
	t.Cleanup(func() { _ = testcontainers.TerminateContainer(c) })
	url, err := c.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("testkit: connection string: %v", err)
	}
	return &Postgres{URL: url, Container: c}
}
