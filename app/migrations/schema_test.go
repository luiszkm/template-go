package migrations_test

import (
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"

	"github.com/luiszkm/template-go/internal/platform/testkit"
)

func exec(t *testing.T, pool *pgxpool.Pool, sql string, args ...any) error {
	t.Helper()
	_, err := pool.Exec(t.Context(), sql, args...)
	return err
}

func TestSchema_UsersConstraints(t *testing.T) {
	pool := testkit.MigratedDB(t)

	require.Error(t, exec(t, pool, `INSERT INTO users (email, name, password_hash) VALUES ('A@x.com', 'A', 'h')`))
	require.NoError(t, exec(t, pool, `INSERT INTO users (email, name, password_hash) VALUES ('a@x.com', 'A', 'h')`))
	require.Error(t, exec(t, pool, `INSERT INTO users (email, name, password_hash) VALUES ('a@x.com', 'B', 'h')`))

	var id uuid.UUID
	require.NoError(t, pool.QueryRow(t.Context(), `SELECT id FROM users WHERE email = 'a@x.com'`).Scan(&id))
	require.NotEqual(t, uuid.Nil, id)

	require.NoError(t, exec(t, pool, `INSERT INTO sessions (token_hash, user_id) VALUES ('\x01', $1)`, id))
	var role uuid.UUID
	require.NoError(t, pool.QueryRow(t.Context(), `INSERT INTO roles (name) VALUES ('r') RETURNING id`).Scan(&role))
	require.NoError(t, exec(t, pool, `INSERT INTO user_roles (user_id, role_id) VALUES ($1, $2)`, id, role))
	require.Error(t, exec(t, pool, `INSERT INTO user_roles (user_id, role_id) VALUES ($1, $2)`, id, role))
	require.NoError(t, exec(t, pool, `INSERT INTO role_permissions (role_id, permission) VALUES ($1, 'x:y')`, role))
	require.Error(t, exec(t, pool, `INSERT INTO role_permissions (role_id, permission) VALUES ($1, 'x:y')`, role))
	require.Error(t, exec(t, pool, `INSERT INTO login_attempts (kind, key) VALUES ('other', 'k')`))

	require.NoError(t, exec(t, pool, `DELETE FROM users WHERE id = $1`, id))
	var sessions int
	require.NoError(t, pool.QueryRow(t.Context(), `SELECT count(*) FROM sessions`).Scan(&sessions))
	require.Zero(t, sessions)
}

func TestSchema_AdminRoleSeeded(t *testing.T) {
	pool := testkit.MigratedDB(t)
	rows, err := pool.Query(t.Context(),
		`SELECT rp.permission FROM roles r JOIN role_permissions rp ON rp.role_id = r.id WHERE r.name = 'admin'`)
	require.NoError(t, err)
	var perms []string
	for rows.Next() {
		var p string
		require.NoError(t, rows.Scan(&p))
		perms = append(perms, p)
	}
	require.NoError(t, rows.Err())
	require.Equal(t, []string{"*"}, perms)
}

func TestSchema_RoleNameUniqueIgnoringCase(t *testing.T) {
	pool := testkit.MigratedDB(t)
	err := exec(t, pool, `INSERT INTO roles (name) VALUES ('ADMIN')`)
	require.ErrorContains(t, err, "23505")
	require.NoError(t, exec(t, pool, `INSERT INTO roles (name) VALUES ('Financeiro')`))
	var name string
	require.NoError(t, pool.QueryRow(t.Context(), `SELECT name FROM roles WHERE lower(name) = 'financeiro'`).Scan(&name))
	require.Equal(t, "Financeiro", name)
}

func TestSchema_AuditIndexes(t *testing.T) {
	pool := testkit.MigratedDB(t)
	rows, err := pool.Query(t.Context(), `SELECT indexname, indexdef FROM pg_indexes WHERE tablename = 'audit_events'`)
	require.NoError(t, err)
	defs := map[string]string{}
	for rows.Next() {
		var name, def string
		require.NoError(t, rows.Scan(&name, &def))
		defs[name] = def
	}
	require.NoError(t, rows.Err())
	want := map[string]string{
		"audit_events_actor_idx":    "(actor_id, id DESC)",
		"audit_events_resource_idx": "(resource_type, resource_id, id DESC)",
		"audit_events_action_idx":   "(action, id DESC)",
		"audit_events_occurred_idx": "(occurred_at)",
	}
	for name, columns := range want {
		require.Contains(t, defs, name)
		require.True(t, strings.HasSuffix(defs[name], "USING btree "+columns), defs[name])
	}
}
