package audittest

import (
	"net/http"
	"testing"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/luiszkm/template-go/internal/platform/deps"
	"github.com/luiszkm/template-go/internal/platform/testkit"
)

type Register func(huma.API, deps.Deps) error

func Serve(t *testing.T, pool *pgxpool.Pool, slices ...Register) http.Handler {
	t.Helper()
	api, h, d := testkit.NewAPI(t, pool, testkit.APIOptions{})
	for _, register := range slices {
		if err := register(api, d); err != nil {
			t.Fatalf("audittest: register: %v", err)
		}
	}
	return h
}

type Event struct {
	Action       string
	Actor        *uuid.UUID
	ResourceType string
	ResourceID   string
	Before       string
	After        string
	At           time.Time
	NoIP         bool
}

func Insert(t *testing.T, pool *pgxpool.Pool, e Event) int64 {
	t.Helper()
	if e.Action == "" {
		e.Action = "thing.done"
	}
	if e.ResourceType == "" {
		e.ResourceType = "thing"
	}
	if e.ResourceID == "" {
		e.ResourceID = "1"
	}
	if e.At.IsZero() {
		e.At = time.Now()
	}
	var before, after any
	if e.Before != "" {
		before = e.Before
	}
	if e.After != "" {
		after = e.After
	}
	var ip any = "10.0.0.7"
	if e.NoIP {
		ip = nil
	}
	var id int64
	err := pool.QueryRow(t.Context(), `
		INSERT INTO audit_events (occurred_at, actor_id, action, resource_type, resource_id, before, after, ip, request_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 'req-1') RETURNING id`,
		e.At, e.Actor, e.Action, e.ResourceType, e.ResourceID, before, after, ip).Scan(&id)
	if err != nil {
		t.Fatalf("audittest: insert event: %v", err)
	}
	return id
}

func InsertUser(t *testing.T, pool *pgxpool.Pool, email string) uuid.UUID {
	t.Helper()
	var id uuid.UUID
	err := pool.QueryRow(t.Context(), `INSERT INTO users (email, name, password_hash) VALUES ($1, 'Test', 'x') RETURNING id`, email).Scan(&id)
	if err != nil {
		t.Fatalf("audittest: insert user: %v", err)
	}
	return id
}
