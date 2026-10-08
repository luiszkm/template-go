package deleterole_test

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"

	deleterole "github.com/luiszkm/template-go/internal/features/rbac/delete_role"
	"github.com/luiszkm/template-go/internal/features/rbac/rbactest"
	"github.com/luiszkm/template-go/internal/platform/testkit"
)

type fixture struct {
	pool   *pgxpool.Pool
	h      http.Handler
	caller testkit.User
}

func setup(t *testing.T) fixture {
	t.Helper()
	pool := testkit.MigratedDB(t)
	return fixture{pool: pool, h: rbactest.Serve(t, pool, deleterole.Register), caller: testkit.SignIn(t, pool, "rbac:delete")}
}

func (f fixture) delete(t *testing.T, id string) int {
	t.Helper()
	return testkit.Do(t, f.h, testkit.Request{Method: http.MethodDelete, Path: "/api/v1/rbac/roles/" + id, Cookie: f.caller.Cookie}).Code
}

func TestDeleteRole_Deletes(t *testing.T) {
	f := setup(t)
	id := rbactest.CreateRole(t, f.pool, "Leitor", "users:read", "users:create")

	require.Equal(t, http.StatusNoContent, f.delete(t, id.String()))
	require.Zero(t, rbactest.Count(t, f.pool, `SELECT count(*) FROM roles WHERE id = $1`, id))
	require.Zero(t, rbactest.Count(t, f.pool, `SELECT count(*) FROM role_permissions WHERE role_id = $1`, id))
	require.Equal(t, 1, rbactest.Count(t, f.pool, `SELECT count(*) FROM audit_events`))
	var before []byte
	require.NoError(t, f.pool.QueryRow(t.Context(),
		`SELECT before FROM audit_events WHERE action = 'role.deleted' AND resource_type = 'role' AND resource_id = $1 AND actor_id = $2`,
		id.String(), f.caller.ID).Scan(&before))
	var snapshot struct {
		Name        string   `json:"name"`
		Permissions []string `json:"permissions"`
	}
	require.NoError(t, json.Unmarshal(before, &snapshot))
	require.Equal(t, "Leitor", snapshot.Name)
	require.Equal(t, []string{"users:create", "users:read"}, snapshot.Permissions)
}

func TestDeleteRole_AdminIs409(t *testing.T) {
	f := setup(t)
	admin := rbactest.AdminRoleID(t, f.pool)
	require.Equal(t, http.StatusConflict, f.delete(t, admin.String()))
	require.Equal(t, 1, rbactest.Count(t, f.pool, `SELECT count(*) FROM role_permissions WHERE role_id = $1 AND permission = '*'`, admin))
	require.Zero(t, rbactest.Count(t, f.pool, `SELECT count(*) FROM audit_events`))
}

func TestDeleteRole_InUseIs409(t *testing.T) {
	f := setup(t)
	id := rbactest.CreateRole(t, f.pool, "Leitor", "users:read")
	user := rbactest.InsertUser(t, f.pool)
	rbactest.Assign(t, f.pool, user, id)

	require.Equal(t, http.StatusConflict, f.delete(t, id.String()))
	require.Equal(t, 1, rbactest.Count(t, f.pool, `SELECT count(*) FROM roles WHERE id = $1`, id))
	require.Equal(t, 1, rbactest.Count(t, f.pool, `SELECT count(*) FROM role_permissions WHERE role_id = $1`, id))
	require.Equal(t, 1, rbactest.Count(t, f.pool, `SELECT count(*) FROM user_roles WHERE role_id = $1 AND user_id = $2`, id, user))
	require.Zero(t, rbactest.Count(t, f.pool, `SELECT count(*) FROM audit_events`))
}

func TestDeleteRole_WaitsForConcurrentAssignment(t *testing.T) {
	f := setup(t)
	id := rbactest.CreateRole(t, f.pool, "Leitor")
	user := rbactest.InsertUser(t, f.pool)

	tx, err := f.pool.Begin(t.Context())
	require.NoError(t, err)
	_, err = tx.Exec(t.Context(), `INSERT INTO user_roles (user_id, role_id) VALUES ($1, $2)`, user, id)
	require.NoError(t, err)

	done := make(chan int, 1)
	go func() { done <- f.delete(t, id.String()) }()
	select {
	case code := <-done:
		t.Fatalf("delete answered %d while the assignment was uncommitted", code)
	case <-time.After(500 * time.Millisecond):
	}
	require.NoError(t, tx.Commit(t.Context()))

	require.Equal(t, http.StatusConflict, <-done)
	require.Equal(t, 1, rbactest.Count(t, f.pool, `SELECT count(*) FROM user_roles WHERE role_id = $1 AND user_id = $2`, id, user))
}

func TestDeleteRole_404And422(t *testing.T) {
	f := setup(t)
	require.Equal(t, http.StatusNotFound, f.delete(t, uuid.NewString()))
	require.Equal(t, http.StatusUnprocessableEntity, f.delete(t, "abc"))
}
