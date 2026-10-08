package deactivateuser_test

import (
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	deactivateuser "github.com/luiszkm/template-go/internal/features/users/deactivate_user"
	"github.com/luiszkm/template-go/internal/features/users/me"
	"github.com/luiszkm/template-go/internal/features/users/userstest"
	"github.com/luiszkm/template-go/internal/platform/auth"
	"github.com/luiszkm/template-go/internal/platform/testkit"
)

func deactivate(t *testing.T, h http.Handler, id string, c *http.Cookie) int {
	t.Helper()
	return testkit.Do(t, h, testkit.Request{Method: http.MethodPost, Path: "/api/v1/users/" + id + "/deactivate", Cookie: c}).Code
}

func meStatus(t *testing.T, h http.Handler, c *http.Cookie) int {
	t.Helper()
	return testkit.Do(t, h, testkit.Request{Method: http.MethodGet, Path: "/api/v1/users/me", Cookie: c}).Code
}

func TestDeactivateUser_RevokesSessions(t *testing.T) {
	pool := testkit.MigratedDB(t)
	h := userstest.Serve(t, pool, testkit.APIOptions{}, deactivateuser.Register, me.Register)
	admin := testkit.SignIn(t, pool, "users:deactivate")
	target := testkit.SignIn(t, pool)
	token, err := auth.IssueSession(t.Context(), pool, target.ID)
	require.NoError(t, err)
	second := &http.Cookie{Name: auth.CookieName, Value: token}
	require.Equal(t, http.StatusOK, meStatus(t, h, target.Cookie))

	require.Equal(t, http.StatusNoContent, deactivate(t, h, target.ID.String(), admin.Cookie))
	require.Equal(t, 1, userstest.Count(t, pool, `SELECT count(*) FROM users WHERE id = $1 AND deactivated_at IS NOT NULL`, target.ID))
	require.Equal(t, 0, userstest.Count(t, pool, `SELECT count(*) FROM sessions WHERE user_id = $1`, target.ID))
	require.Equal(t, 1, userstest.Count(t, pool,
		`SELECT count(*) FROM audit_events WHERE action = 'user.deactivated' AND resource_id = $1 AND actor_id = $2`, target.ID.String(), admin.ID))
	require.Equal(t, http.StatusUnauthorized, meStatus(t, h, target.Cookie))
	require.Equal(t, http.StatusUnauthorized, meStatus(t, h, second))
}

func TestDeactivateUser_SelfIs409(t *testing.T) {
	pool := testkit.MigratedDB(t)
	h := userstest.Serve(t, pool, testkit.APIOptions{}, deactivateuser.Register, me.Register)
	admin := testkit.SignIn(t, pool, "users:deactivate")

	require.Equal(t, http.StatusConflict, deactivate(t, h, admin.ID.String(), admin.Cookie))
	require.Equal(t, 0, userstest.Count(t, pool, `SELECT count(*) FROM users WHERE id = $1 AND deactivated_at IS NOT NULL`, admin.ID))
	require.Equal(t, http.StatusOK, meStatus(t, h, admin.Cookie))
}

func TestDeactivateUser_AlreadyDeactivatedNoop(t *testing.T) {
	pool := testkit.MigratedDB(t)
	h := userstest.Serve(t, pool, testkit.APIOptions{}, deactivateuser.Register)
	admin := testkit.SignIn(t, pool, "users:deactivate")
	id := userstest.InsertUser(t, pool, "off@x.com", "x")
	_, err := pool.Exec(t.Context(), `UPDATE users SET deactivated_at = '2026-01-01T00:00:00Z' WHERE id = $1`, id)
	require.NoError(t, err)

	require.Equal(t, http.StatusNoContent, deactivate(t, h, id.String(), admin.Cookie))
	require.Equal(t, 1, userstest.Count(t, pool, `SELECT count(*) FROM users WHERE id = $1 AND deactivated_at = '2026-01-01T00:00:00Z'`, id))
	require.Equal(t, 0, userstest.Count(t, pool, `SELECT count(*) FROM audit_events`))
}

func TestDeactivateUser_404And422(t *testing.T) {
	pool := testkit.MigratedDB(t)
	h := userstest.Serve(t, pool, testkit.APIOptions{}, deactivateuser.Register)
	admin := testkit.SignIn(t, pool, "users:deactivate")
	require.Equal(t, http.StatusNotFound, deactivate(t, h, uuid.NewString(), admin.Cookie))
	require.Equal(t, http.StatusUnprocessableEntity, deactivate(t, h, "abc", admin.Cookie))
}
