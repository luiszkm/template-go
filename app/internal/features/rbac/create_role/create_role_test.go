package createrole_test

import (
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"

	"github.com/luiszkm/template-go/internal/features/rbac"
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
	return fixture{pool: pool, h: rbactest.Serve(t, pool, rbac.Register, rbactest.Declare("users:read")), caller: testkit.SignIn(t, pool, "rbac:create")}
}

func (f fixture) create(t *testing.T, body any) *httptest.ResponseRecorder {
	t.Helper()
	return testkit.Do(t, f.h, testkit.Request{Method: http.MethodPost, Path: "/api/v1/rbac/roles", Body: body, Cookie: f.caller.Cookie})
}

type problem struct {
	Errors []struct {
		Location string `json:"location"`
	} `json:"errors"`
}

func TestCreateRole_Creates(t *testing.T) {
	f := setup(t)
	rec := f.create(t, map[string]any{"name": "  Leitor ", "permissions": []string{"users:read"}})
	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())

	body := testkit.JSON[map[string]any](t, rec)
	raw, ok := body["id"].(string)
	require.True(t, ok)
	id, err := uuid.Parse(raw)
	require.NoError(t, err)
	require.Equal(t, "Leitor", body["name"])
	require.Equal(t, []any{"users:read"}, body["permissions"])
	require.InDelta(t, 0, body["user_count"], 0)
	require.Equal(t, 1, rbactest.Count(t, f.pool, `SELECT count(*) FROM audit_events`))
	require.Equal(t, 1, rbactest.Count(t, f.pool,
		`SELECT count(*) FROM audit_events WHERE action = 'role.created' AND resource_type = 'role' AND resource_id = $1 AND actor_id = $2`,
		id.String(), f.caller.ID))
}

func TestCreateRole_DuplicateNameIs409(t *testing.T) {
	f := setup(t)
	rbactest.CreateRole(t, f.pool, "Leitor")
	before := rbactest.Count(t, f.pool, `SELECT count(*) FROM roles`)

	rec := f.create(t, map[string]any{"name": "LEITOR", "permissions": []string{}})
	require.Equal(t, http.StatusConflict, rec.Code, rec.Body.String())
	require.Equal(t, before, rbactest.Count(t, f.pool, `SELECT count(*) FROM roles`))
	require.Zero(t, rbactest.Count(t, f.pool, `SELECT count(*) FROM audit_events`))
}

func TestCreateRole_ConcurrentSameName(t *testing.T) {
	f := setup(t)
	codes := make([]int, 2)
	var wg sync.WaitGroup
	for i := range codes {
		wg.Go(func() { codes[i] = f.create(t, map[string]any{"name": "Paralelo", "permissions": []string{}}).Code })
	}
	wg.Wait()
	slices.Sort(codes)
	require.Equal(t, []int{http.StatusCreated, http.StatusConflict}, codes)
	require.Equal(t, 1, rbactest.Count(t, f.pool, `SELECT count(*) FROM roles WHERE name = 'Paralelo'`))
}

func TestCreateRole_Validation(t *testing.T) {
	f := setup(t)
	rejected := []struct {
		name        string
		permissions []string
		location    string
	}{
		{"", []string{}, "body.name"},
		{"   ", []string{}, "body.name"},
		{strings.Repeat("a", 51), []string{}, "body.name"},
		{"Ok", []string{"*"}, "body.permissions"},
		{"Ok", []string{"nope:x"}, "body.permissions"},
		{"Ok", []string{"users:read", "users:read"}, "body.permissions"},
	}
	roles := rbactest.Count(t, f.pool, `SELECT count(*) FROM roles`)
	for _, c := range rejected {
		rec := f.create(t, map[string]any{"name": c.name, "permissions": c.permissions})
		require.Equal(t, http.StatusUnprocessableEntity, rec.Code, "%q %v", c.name, c.permissions)
		var locations []string
		for _, e := range testkit.JSON[problem](t, rec).Errors {
			locations = append(locations, e.Location)
		}
		require.Contains(t, locations, c.location, "%q %v", c.name, c.permissions)
	}
	require.Equal(t, roles, rbactest.Count(t, f.pool, `SELECT count(*) FROM roles`))

	accepted := []map[string]any{
		{"name": "a", "permissions": []string{"users:read"}},
		{"name": strings.Repeat("b", 50), "permissions": []string{"users:read"}},
		{"name": "Vazio", "permissions": []string{}},
	}
	for _, body := range accepted {
		rec := f.create(t, body)
		require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
	}
}
