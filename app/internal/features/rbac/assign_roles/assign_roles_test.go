package assignroles_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"slices"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"

	assignroles "github.com/luiszkm/template-go/internal/features/rbac/assign_roles"
	"github.com/luiszkm/template-go/internal/features/rbac/rbactest"
	"github.com/luiszkm/template-go/internal/platform/auth"
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
	h := rbactest.Serve(t, pool, assignroles.Register, rbactest.SessionProbe)
	return fixture{pool: pool, h: h, caller: testkit.SignIn(t, pool, "rbac:assign")}
}

func (f fixture) put(t *testing.T, userID string, roleIDs ...uuid.UUID) *httptest.ResponseRecorder {
	t.Helper()
	if roleIDs == nil {
		roleIDs = []uuid.UUID{}
	}
	return testkit.Do(t, f.h, testkit.Request{Method: http.MethodPut, Path: "/api/v1/rbac/users/" + userID + "/roles",
		Body: map[string]any{"role_ids": roleIDs}, Cookie: f.caller.Cookie})
}

type state struct {
	roles    []uuid.UUID
	sessions int
	audits   int
}

func (f fixture) state(t *testing.T, userID uuid.UUID) state {
	t.Helper()
	rows, err := f.pool.Query(t.Context(), `SELECT role_id FROM user_roles WHERE user_id = $1 ORDER BY role_id`, userID)
	require.NoError(t, err)
	roles := []uuid.UUID{}
	for rows.Next() {
		var id uuid.UUID
		require.NoError(t, rows.Scan(&id))
		roles = append(roles, id)
	}
	require.NoError(t, rows.Err())
	return state{roles: roles,
		sessions: rbactest.Count(t, f.pool, `SELECT count(*) FROM sessions WHERE user_id = $1`, userID),
		audits:   rbactest.Count(t, f.pool, `SELECT count(*) FROM audit_events`)}
}

func sorted(ids ...uuid.UUID) []uuid.UUID {
	return slices.SortedFunc(slices.Values(ids), func(a, b uuid.UUID) int { return slices.Compare(a[:], b[:]) })
}

func (f fixture) cookie(t *testing.T, userID uuid.UUID) *http.Cookie {
	t.Helper()
	token, err := auth.IssueSession(t.Context(), f.pool, userID)
	require.NoError(t, err)
	return &http.Cookie{Name: auth.CookieName, Value: token}
}

func (f fixture) probe(t *testing.T, c *http.Cookie) int {
	t.Helper()
	return testkit.Do(t, f.h, testkit.Request{Method: http.MethodGet, Path: rbactest.SessionProbePath, Cookie: c}).Code
}

func TestAssignRoles_ReplacesAndRevokes(t *testing.T) {
	f := setup(t)
	alfa := rbactest.CreateRole(t, f.pool, "Alfa")
	beta := rbactest.CreateRole(t, f.pool, "Beta")
	gama := rbactest.CreateRole(t, f.pool, "Gama")
	user := rbactest.InsertUser(t, f.pool)
	rbactest.Assign(t, f.pool, user, alfa)
	first, second := f.cookie(t, user), f.cookie(t, user)
	require.Equal(t, http.StatusOK, f.probe(t, first))

	rec := f.put(t, user.String(), gama, beta)
	require.Equal(t, http.StatusNoContent, rec.Code, rec.Body.String())
	require.Equal(t, sorted(beta, gama), f.state(t, user).roles)
	require.Zero(t, f.state(t, user).sessions)
	require.Equal(t, http.StatusUnauthorized, f.probe(t, first))
	require.Equal(t, http.StatusUnauthorized, f.probe(t, second))

	require.Equal(t, 1, rbactest.Count(t, f.pool, `SELECT count(*) FROM audit_events`))
	var before, after []byte
	require.NoError(t, f.pool.QueryRow(t.Context(),
		`SELECT before, after FROM audit_events WHERE action = 'user.roles_changed' AND resource_type = 'user' AND resource_id = $1 AND actor_id = $2`,
		user.String(), f.caller.ID).Scan(&before, &after))
	require.JSONEq(t, `{"roles":["Alfa"]}`, string(before))
	require.JSONEq(t, `{"roles":["Beta","Gama"]}`, string(after))
}

func TestAssignRoles_SameSetIsNoop(t *testing.T) {
	f := setup(t)
	beta := rbactest.CreateRole(t, f.pool, "Beta")
	gama := rbactest.CreateRole(t, f.pool, "Gama")
	user := rbactest.InsertUser(t, f.pool)
	rbactest.Assign(t, f.pool, user, beta, gama)
	f.cookie(t, user)
	was := f.state(t, user)

	require.Equal(t, http.StatusNoContent, f.put(t, user.String(), gama, beta).Code)
	require.Equal(t, was, f.state(t, user))
	require.Zero(t, f.state(t, user).audits)
}

func TestAssignRoles_InvalidRoleIds(t *testing.T) {
	f := setup(t)
	beta := rbactest.CreateRole(t, f.pool, "Beta")
	alfa := rbactest.CreateRole(t, f.pool, "Alfa")
	user := rbactest.InsertUser(t, f.pool)
	rbactest.Assign(t, f.pool, user, alfa)
	f.cookie(t, user)
	was := f.state(t, user)

	for _, ids := range [][]uuid.UUID{{beta, uuid.New()}, {beta, beta}} {
		rec := f.put(t, user.String(), ids...)
		require.Equal(t, http.StatusUnprocessableEntity, rec.Code, rec.Body.String())
		require.Contains(t, rec.Body.String(), `"location":"body.role_ids"`)
		require.Equal(t, was, f.state(t, user))
	}
}

func TestAssignRoles_404And422(t *testing.T) {
	f := setup(t)
	require.Equal(t, http.StatusNotFound, f.put(t, uuid.NewString()).Code)
	require.Equal(t, http.StatusUnprocessableEntity, f.put(t, "abc").Code)
}

func TestAssignRoles_LastAdminGuard(t *testing.T) {
	cases := []struct {
		name   string
		other  string
		target string
		keep   bool
		want   int
	}{
		{"only active admin loses admin", "none", "active", false, http.StatusConflict},
		{"other admin is deactivated", "deactivated", "active", false, http.StatusConflict},
		{"another active admin remains", "active", "active", false, http.StatusNoContent},
		{"only admin keeps admin and gains a role", "none", "active", true, http.StatusNoContent},
		{"deactivated admin loses admin while one active remains", "active", "deactivated", false, http.StatusNoContent},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			f := setup(t)
			admin := rbactest.AdminRoleID(t, f.pool)
			extra := rbactest.CreateRole(t, f.pool, "Extra")
			target := rbactest.InsertUser(t, f.pool)
			rbactest.Assign(t, f.pool, target, admin)
			f.cookie(t, target)
			if c.target == "deactivated" {
				rbactest.Deactivate(t, f.pool, target)
			}
			if c.other != "none" {
				other := rbactest.InsertUser(t, f.pool)
				rbactest.Assign(t, f.pool, other, admin)
				if c.other == "deactivated" {
					rbactest.Deactivate(t, f.pool, other)
				}
			}
			roles := []uuid.UUID{extra}
			if c.keep {
				roles = append(roles, admin)
			}
			was := f.state(t, target)

			rec := f.put(t, target.String(), roles...)
			require.Equal(t, c.want, rec.Code, rec.Body.String())
			if c.want == http.StatusConflict {
				require.Equal(t, was, f.state(t, target))
			}
		})
	}
}

func TestAssignRoles_ConcurrentLastAdmin(t *testing.T) {
	for range 10 {
		f := setup(t)
		admin := rbactest.AdminRoleID(t, f.pool)
		a, b := rbactest.InsertUser(t, f.pool), rbactest.InsertUser(t, f.pool)
		rbactest.Assign(t, f.pool, a, admin)
		rbactest.Assign(t, f.pool, b, admin)

		codes := make([]int, 2)
		var wg sync.WaitGroup
		for i, user := range []uuid.UUID{a, b} {
			wg.Go(func() { codes[i] = f.put(t, user.String()).Code })
		}
		wg.Wait()
		slices.Sort(codes)
		require.Equal(t, []int{http.StatusNoContent, http.StatusConflict}, codes)
		require.Equal(t, 1, rbactest.Count(t, f.pool,
			`SELECT count(*) FROM user_roles ur JOIN users u ON u.id = ur.user_id WHERE ur.role_id = $1 AND u.deactivated_at IS NULL`, admin))
	}
}

func TestAssignRoles_RollsBackTogether(t *testing.T) {
	f := setup(t)
	alfa := rbactest.CreateRole(t, f.pool, "Alfa")
	beta := rbactest.CreateRole(t, f.pool, "Beta")
	user := rbactest.InsertUser(t, f.pool)
	rbactest.Assign(t, f.pool, user, alfa)
	f.cookie(t, user)
	_, err := f.pool.Exec(t.Context(), `
		CREATE FUNCTION fail_user_roles() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RAISE EXCEPTION 'injected'; END; $$;
		CREATE TRIGGER fail_user_roles BEFORE INSERT ON user_roles FOR EACH ROW EXECUTE FUNCTION fail_user_roles();`)
	require.NoError(t, err)
	was := f.state(t, user)

	require.Equal(t, http.StatusInternalServerError, f.put(t, user.String(), beta).Code)
	require.Equal(t, was, f.state(t, user))
}

func TestAssignRoles_AuditCarriesSortedNames(t *testing.T) {
	f := setup(t)
	user := rbactest.InsertUser(t, f.pool)
	zeta := rbactest.CreateRole(t, f.pool, "Zeta")
	beta := rbactest.CreateRole(t, f.pool, "Beta")
	require.Equal(t, http.StatusNoContent, f.put(t, user.String(), zeta, beta).Code)
	var after []byte
	require.NoError(t, f.pool.QueryRow(t.Context(), `SELECT after FROM audit_events WHERE action = 'user.roles_changed'`).Scan(&after))
	var got struct {
		Roles []string `json:"roles"`
	}
	require.NoError(t, json.Unmarshal(after, &got))
	require.Equal(t, []string{"Beta", "Zeta"}, got.Roles)
}
