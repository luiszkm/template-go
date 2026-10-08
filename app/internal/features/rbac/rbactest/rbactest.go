package rbactest

import (
	"context"
	"net/http"
	"testing"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/luiszkm/template-go/internal/platform/deps"
	"github.com/luiszkm/template-go/internal/platform/op"
	"github.com/luiszkm/template-go/internal/platform/testkit"
)

type Register func(huma.API, deps.Deps) error

const ProbePath = "/api/v1/probe"

const ProbePermission op.Permission = "zeta:do"

type probe struct{}

type probeOutput struct {
	Body struct {
		OK bool `json:"ok"`
	}
}

func answer(context.Context, *probe) (*probeOutput, error) {
	out := &probeOutput{}
	out.Body.OK = true
	return out, nil
}

func Probe(api huma.API, _ deps.Deps) error {
	return op.Register(api, op.Spec{ID: "probe", Method: http.MethodGet, Path: ProbePath, Permission: ProbePermission},
		answer)
}

const SessionProbePath = "/api/v1/session-probe"

func SessionProbe(api huma.API, _ deps.Deps) error {
	return op.Register(api, op.Spec{ID: "session-probe", Method: http.MethodGet, Path: SessionProbePath, Authenticated: true},
		answer)
}

func Declare(permissions ...op.Permission) Register {
	return func(api huma.API, _ deps.Deps) error {
		for _, p := range permissions {
			err := op.Register(api, op.Spec{ID: "declare-" + string(p), Method: http.MethodGet, Path: "/api/v1/declared/" + string(p), Permission: p},
				answer)
			if err != nil {
				return err
			}
		}
		return nil
	}
}

func Serve(t *testing.T, pool *pgxpool.Pool, slices ...Register) http.Handler {
	t.Helper()
	api, h, d := testkit.NewAPI(t, pool, testkit.APIOptions{})
	for _, register := range slices {
		if err := register(api, d); err != nil {
			t.Fatalf("rbactest: register: %v", err)
		}
	}
	return h
}

func CreateRole(t *testing.T, pool *pgxpool.Pool, name string, permissions ...string) uuid.UUID {
	t.Helper()
	var id uuid.UUID
	if err := pool.QueryRow(t.Context(), `INSERT INTO roles (name) VALUES ($1) RETURNING id`, name).Scan(&id); err != nil {
		t.Fatalf("rbactest: insert role: %v", err)
	}
	for _, p := range permissions {
		if _, err := pool.Exec(t.Context(), `INSERT INTO role_permissions (role_id, permission) VALUES ($1, $2)`, id, p); err != nil {
			t.Fatalf("rbactest: insert permission: %v", err)
		}
	}
	return id
}

func AdminRoleID(t *testing.T, pool *pgxpool.Pool) uuid.UUID {
	t.Helper()
	var id uuid.UUID
	if err := pool.QueryRow(t.Context(), `SELECT id FROM roles WHERE name = 'admin'`).Scan(&id); err != nil {
		t.Fatalf("rbactest: admin role: %v", err)
	}
	return id
}

func InsertUser(t *testing.T, pool *pgxpool.Pool) uuid.UUID {
	t.Helper()
	var id uuid.UUID
	err := pool.QueryRow(t.Context(), `INSERT INTO users (email, name, password_hash) VALUES ($1, 'Test', 'x') RETURNING id`,
		uuid.NewString()+"@test.local").Scan(&id)
	if err != nil {
		t.Fatalf("rbactest: insert user: %v", err)
	}
	return id
}

func Assign(t *testing.T, pool *pgxpool.Pool, userID uuid.UUID, roleIDs ...uuid.UUID) {
	t.Helper()
	for _, r := range roleIDs {
		if _, err := pool.Exec(t.Context(), `INSERT INTO user_roles (user_id, role_id) VALUES ($1, $2)`, userID, r); err != nil {
			t.Fatalf("rbactest: assign: %v", err)
		}
	}
}

func Deactivate(t *testing.T, pool *pgxpool.Pool, userID uuid.UUID) {
	t.Helper()
	if _, err := pool.Exec(t.Context(), `UPDATE users SET deactivated_at = now() WHERE id = $1`, userID); err != nil {
		t.Fatalf("rbactest: deactivate: %v", err)
	}
}

func Count(t *testing.T, pool *pgxpool.Pool, sql string, args ...any) int {
	t.Helper()
	var n int
	if err := pool.QueryRow(t.Context(), sql, args...).Scan(&n); err != nil {
		t.Fatalf("rbactest: count: %v", err)
	}
	return n
}
