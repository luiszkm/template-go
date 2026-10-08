package logout_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/luiszkm/template-go/internal/features/users/logout"
	"github.com/luiszkm/template-go/internal/features/users/me"
	"github.com/luiszkm/template-go/internal/features/users/userstest"
	"github.com/luiszkm/template-go/internal/platform/testkit"
)

func TestLogout_DeletesSession(t *testing.T) {
	pool := testkit.MigratedDB(t)
	h := userstest.Serve(t, pool, testkit.APIOptions{}, logout.Register, me.Register)
	u := testkit.SignIn(t, pool)

	rec := testkit.Do(t, h, testkit.Request{Method: http.MethodDelete, Path: "/api/v1/users/session", Cookie: u.Cookie})
	require.Equal(t, http.StatusNoContent, rec.Code, rec.Body.String())
	cookies := rec.Header().Values("Set-Cookie")
	require.Len(t, cookies, 1)
	require.Contains(t, cookies[0], "session=;")
	require.Contains(t, cookies[0], "Path=/")
	require.Contains(t, cookies[0], "Max-Age=0")
	require.Equal(t, 0, userstest.Count(t, pool, `SELECT count(*) FROM sessions WHERE user_id = $1`, u.ID))
	require.Equal(t, 1, userstest.Count(t, pool,
		`SELECT count(*) FROM audit_events WHERE action = 'session.deleted' AND actor_id = $1`, u.ID))
	require.Equal(t, http.StatusUnauthorized,
		testkit.Do(t, h, testkit.Request{Method: http.MethodGet, Path: "/api/v1/users/me", Cookie: u.Cookie}).Code)
}

func TestLogout_WithoutSession401(t *testing.T) {
	pool := testkit.MigratedDB(t)
	h := userstest.Serve(t, pool, testkit.APIOptions{}, logout.Register)
	rec := testkit.Do(t, h, testkit.Request{Method: http.MethodDelete, Path: "/api/v1/users/session"})
	require.Equal(t, http.StatusUnauthorized, rec.Code)
}
