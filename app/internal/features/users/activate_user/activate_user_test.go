package activateuser_test

import (
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	activateuser "github.com/luiszkm/template-go/internal/features/users/activate_user"
	"github.com/luiszkm/template-go/internal/features/users/login"
	"github.com/luiszkm/template-go/internal/features/users/userstest"
	"github.com/luiszkm/template-go/internal/platform/testkit"
)

func activate(t *testing.T, h http.Handler, id string, c *http.Cookie) int {
	t.Helper()
	return testkit.Do(t, h, testkit.Request{Method: http.MethodPost, Path: "/api/v1/users/" + id + "/activate", Cookie: c}).Code
}

func TestActivateUser_Reactivates(t *testing.T) {
	pool := testkit.MigratedDB(t)
	h := userstest.Serve(t, pool, testkit.APIOptions{}, activateuser.Register, login.Register)
	admin := testkit.SignIn(t, pool, "users:activate")
	id := userstest.CreateUser(t, pool, "off@x.com", "senha-longa-123")
	_, err := pool.Exec(t.Context(), `UPDATE users SET deactivated_at = now() WHERE id = $1`, id)
	require.NoError(t, err)
	signIn := func() int {
		return testkit.Do(t, h, testkit.Request{Method: http.MethodPost, Path: "/api/v1/users/session",
			Body: map[string]string{"email": "off@x.com", "password": "senha-longa-123"}}).Code
	}
	require.Equal(t, http.StatusUnauthorized, signIn())

	require.Equal(t, http.StatusNoContent, activate(t, h, id.String(), admin.Cookie))
	require.Equal(t, 1, userstest.Count(t, pool, `SELECT count(*) FROM users WHERE id = $1 AND deactivated_at IS NULL`, id))
	require.Equal(t, 1, userstest.Count(t, pool,
		`SELECT count(*) FROM audit_events WHERE action = 'user.activated' AND resource_id = $1 AND actor_id = $2`, id.String(), admin.ID))
	require.Equal(t, http.StatusNoContent, signIn())
}

func TestActivateUser_AlreadyActiveNoop(t *testing.T) {
	pool := testkit.MigratedDB(t)
	h := userstest.Serve(t, pool, testkit.APIOptions{}, activateuser.Register)
	admin := testkit.SignIn(t, pool, "users:activate")
	id := userstest.InsertUser(t, pool, "on@x.com", "x")

	require.Equal(t, http.StatusNoContent, activate(t, h, id.String(), admin.Cookie))
	require.Equal(t, 1, userstest.Count(t, pool, `SELECT count(*) FROM users WHERE id = $1 AND deactivated_at IS NULL`, id))
	require.Equal(t, 0, userstest.Count(t, pool, `SELECT count(*) FROM audit_events`))
}

func TestActivateUser_404And422(t *testing.T) {
	pool := testkit.MigratedDB(t)
	h := userstest.Serve(t, pool, testkit.APIOptions{}, activateuser.Register)
	admin := testkit.SignIn(t, pool, "users:activate")
	require.Equal(t, http.StatusNotFound, activate(t, h, uuid.NewString(), admin.Cookie))
	require.Equal(t, http.StatusUnprocessableEntity, activate(t, h, "abc", admin.Cookie))
}
