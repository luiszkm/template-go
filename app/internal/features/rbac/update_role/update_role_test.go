package updaterole_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"

	"github.com/luiszkm/template-go/internal/features/rbac"
	"github.com/luiszkm/template-go/internal/features/rbac/rbactest"
	"github.com/luiszkm/template-go/internal/features/rbac/role"
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
	h := rbactest.Serve(t, pool, rbac.Register, rbactest.Declare("users:read", "users:create"), rbactest.Probe)
	return fixture{pool: pool, h: h, caller: testkit.SignIn(t, pool, "rbac:update")}
}

func (f fixture) patch(t *testing.T, id string, body any) *httptest.ResponseRecorder {
	t.Helper()
	return testkit.Do(t, f.h, testkit.Request{Method: http.MethodPatch, Path: "/api/v1/rbac/roles/" + id, Body: body, Cookie: f.caller.Cookie})
}

func permissionsOf(t *testing.T, pool *pgxpool.Pool, id uuid.UUID) []string {
	t.Helper()
	rows, err := pool.Query(t.Context(), `SELECT permission FROM role_permissions WHERE role_id = $1 ORDER BY permission`, id)
	require.NoError(t, err)
	perms := []string{}
	for rows.Next() {
		var p string
		require.NoError(t, rows.Scan(&p))
		perms = append(perms, p)
	}
	require.NoError(t, rows.Err())
	return perms
}

func nameOf(t *testing.T, pool *pgxpool.Pool, id uuid.UUID) string {
	t.Helper()
	var name string
	require.NoError(t, pool.QueryRow(t.Context(), `SELECT name FROM roles WHERE id = $1`, id).Scan(&name))
	return name
}

type snapshot struct {
	Name        string   `json:"name"`
	Permissions []string `json:"permissions"`
}

func lastAudit(t *testing.T, pool *pgxpool.Pool, id uuid.UUID) (snapshot, snapshot) {
	t.Helper()
	var before, after []byte
	require.NoError(t, pool.QueryRow(t.Context(),
		`SELECT before, after FROM audit_events WHERE action = 'role.updated' AND resource_id = $1 ORDER BY id DESC LIMIT 1`, id.String()).
		Scan(&before, &after))
	var b, a snapshot
	require.NoError(t, json.Unmarshal(before, &b))
	require.NoError(t, json.Unmarshal(after, &a))
	return b, a
}

func TestUpdateRole_UpdatesSentFields(t *testing.T) {
	f := setup(t)
	id := rbactest.CreateRole(t, f.pool, "Leitor", "users:read")

	rec := f.patch(t, id.String(), map[string]any{"name": "Consulta"})
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	got := testkit.JSON[role.Role](t, rec)
	require.Equal(t, "Consulta", got.Name)
	require.Equal(t, []string{"users:read"}, got.Permissions)
	before, after := lastAudit(t, f.pool, id)
	require.Equal(t, snapshot{"Leitor", []string{"users:read"}}, before)
	require.Equal(t, snapshot{"Consulta", []string{"users:read"}}, after)

	rec = f.patch(t, id.String(), map[string]any{"permissions": []string{"users:create"}})
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	got = testkit.JSON[role.Role](t, rec)
	require.Equal(t, "Consulta", got.Name)
	require.Equal(t, []string{"users:create"}, got.Permissions)
	require.Equal(t, []string{"users:create"}, permissionsOf(t, f.pool, id))
	before, after = lastAudit(t, f.pool, id)
	require.Equal(t, snapshot{"Consulta", []string{"users:read"}}, before)
	require.Equal(t, snapshot{"Consulta", []string{"users:create"}}, after)
	require.Equal(t, 2, rbactest.Count(t, f.pool, `SELECT count(*) FROM audit_events WHERE action = 'role.updated'`))
	require.Equal(t, 2, rbactest.Count(t, f.pool, `SELECT count(*) FROM audit_events`))
}

func TestUpdateRole_AppliesOnNextRequest(t *testing.T) {
	f := setup(t)
	id := rbactest.CreateRole(t, f.pool, "Zeta", string(rbactest.ProbePermission))
	holder := testkit.SignIn(t, f.pool)
	rbactest.Assign(t, f.pool, holder.ID, id)
	probe := func() int {
		return testkit.Do(t, f.h, testkit.Request{Method: http.MethodGet, Path: rbactest.ProbePath, Cookie: holder.Cookie}).Code
	}
	require.Equal(t, http.StatusOK, probe())
	sessions := rbactest.Count(t, f.pool, `SELECT count(*) FROM sessions WHERE user_id = $1`, holder.ID)

	require.Equal(t, http.StatusOK, f.patch(t, id.String(), map[string]any{"permissions": []string{}}).Code)
	require.Equal(t, http.StatusForbidden, probe())
	require.Equal(t, sessions, rbactest.Count(t, f.pool, `SELECT count(*) FROM sessions WHERE user_id = $1`, holder.ID))
}

func TestUpdateRole_DuplicateNameIs409(t *testing.T) {
	f := setup(t)
	rbactest.CreateRole(t, f.pool, "Leitor")
	other := rbactest.CreateRole(t, f.pool, "Outro")

	require.Equal(t, http.StatusConflict, f.patch(t, other.String(), map[string]any{"name": " leitor "}).Code)
	require.Equal(t, "Outro", nameOf(t, f.pool, other))
	require.Zero(t, rbactest.Count(t, f.pool, `SELECT count(*) FROM audit_events`))
}

func TestUpdateRole_AdminIs409(t *testing.T) {
	f := setup(t)
	admin := rbactest.AdminRoleID(t, f.pool)
	require.Equal(t, http.StatusConflict, f.patch(t, admin.String(), map[string]any{"name": "Chefe"}).Code)
	require.Equal(t, http.StatusConflict, f.patch(t, admin.String(), map[string]any{"permissions": []string{"users:read"}}).Code)
	require.Equal(t, "admin", nameOf(t, f.pool, admin))
	require.Equal(t, []string{"*"}, permissionsOf(t, f.pool, admin))
	require.Zero(t, rbactest.Count(t, f.pool, `SELECT count(*) FROM audit_events`))
}

func TestUpdateRole_Validation(t *testing.T) {
	f := setup(t)
	id := rbactest.CreateRole(t, f.pool, "Leitor", "users:read")
	rejected := []struct {
		body     map[string]any
		location string
	}{
		{map[string]any{"name": ""}, "body.name"},
		{map[string]any{"name": "   "}, "body.name"},
		{map[string]any{"name": strings.Repeat("a", 51)}, "body.name"},
		{map[string]any{"permissions": []string{"*"}}, "body.permissions"},
		{map[string]any{"permissions": []string{"nope:x"}}, "body.permissions"},
		{map[string]any{"permissions": []string{"users:read", "users:read"}}, "body.permissions"},
	}
	for _, c := range rejected {
		rec := f.patch(t, id.String(), c.body)
		require.Equal(t, http.StatusUnprocessableEntity, rec.Code, "%v", c.body)
		require.Contains(t, rec.Body.String(), `"location":"`+c.location+`"`, "%v", c.body)
	}
	require.Equal(t, "Leitor", nameOf(t, f.pool, id))
	require.Equal(t, []string{"users:read"}, permissionsOf(t, f.pool, id))
	require.Zero(t, rbactest.Count(t, f.pool, `SELECT count(*) FROM audit_events`))
}

func TestUpdateRole_404And422(t *testing.T) {
	f := setup(t)
	require.Equal(t, http.StatusNotFound, f.patch(t, uuid.NewString(), map[string]any{"name": "X"}).Code)
	require.Equal(t, http.StatusUnprocessableEntity, f.patch(t, "abc", map[string]any{"name": "X"}).Code)
}

func TestUpdateRole_UpdatesBothFields(t *testing.T) {
	f := setup(t)
	id := rbactest.CreateRole(t, f.pool, "Leitor", "users:read")

	rec := f.patch(t, id.String(), map[string]any{"name": "Gestor", "permissions": []string{"users:create", "users:read"}})
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	got := testkit.JSON[role.Role](t, rec)
	require.Equal(t, "Gestor", got.Name)
	require.Equal(t, []string{"users:create", "users:read"}, got.Permissions)
	require.Equal(t, "Gestor", nameOf(t, f.pool, id))
	require.Equal(t, []string{"users:create", "users:read"}, permissionsOf(t, f.pool, id))
	before, after := lastAudit(t, f.pool, id)
	require.Equal(t, snapshot{"Leitor", []string{"users:read"}}, before)
	require.Equal(t, snapshot{"Gestor", []string{"users:create", "users:read"}}, after)
	require.Equal(t, 1, rbactest.Count(t, f.pool, `SELECT count(*) FROM audit_events`))
}
