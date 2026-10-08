package changepassword_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	changepassword "github.com/luiszkm/template-go/internal/features/users/change_password"
	"github.com/luiszkm/template-go/internal/features/users/login"
	"github.com/luiszkm/template-go/internal/features/users/me"
	"github.com/luiszkm/template-go/internal/features/users/userstest"
	"github.com/luiszkm/template-go/internal/platform/testkit"
)

const old = "senha-antiga-123"

func setup(t *testing.T) (http.Handler, func() *http.Cookie) {
	t.Helper()
	pool := testkit.MigratedDB(t)
	h := userstest.Serve(t, pool, testkit.APIOptions{}, changepassword.Register, login.Register, me.Register)
	userstest.CreateUser(t, pool, "ana@x.com", old)
	signIn := func() *http.Cookie {
		rec := signInWith(t, h, old)
		require.Equal(t, http.StatusNoContent, rec.Code)
		return userstest.SessionCookie(t, rec.Result())
	}
	return h, signIn
}

func signInWith(t *testing.T, h http.Handler, pw string) *httptest.ResponseRecorder {
	t.Helper()
	return testkit.Do(t, h, testkit.Request{Method: http.MethodPost, Path: "/api/v1/users/session",
		Body: map[string]string{"email": "ana@x.com", "password": pw}})
}

func change(t *testing.T, h http.Handler, c *http.Cookie, current, next string) *httptest.ResponseRecorder {
	t.Helper()
	return testkit.Do(t, h, testkit.Request{Method: http.MethodPut, Path: "/api/v1/users/me/password", Cookie: c,
		Body: map[string]string{"current_password": current, "new_password": next}})
}

func meStatus(t *testing.T, h http.Handler, c *http.Cookie) int {
	t.Helper()
	return testkit.Do(t, h, testkit.Request{Method: http.MethodGet, Path: "/api/v1/users/me", Cookie: c}).Code
}

func TestChangePassword_RevokesOtherSessions(t *testing.T) {
	h, signIn := setup(t)
	current, other := signIn(), signIn()

	require.Equal(t, http.StatusNoContent, change(t, h, current, old, "senha-nova-1234").Code)
	require.Equal(t, http.StatusUnauthorized, meStatus(t, h, other))
	require.Equal(t, http.StatusOK, meStatus(t, h, current))
	require.Equal(t, http.StatusUnauthorized, signInWith(t, h, old).Code)
	require.Equal(t, http.StatusNoContent, signInWith(t, h, "senha-nova-1234").Code)
}

func TestChangePassword_AuditsWithoutPassword(t *testing.T) {
	pool := testkit.MigratedDB(t)
	h := userstest.Serve(t, pool, testkit.APIOptions{}, changepassword.Register)
	id := userstest.CreateUser(t, pool, "ana@x.com", old)
	u := testkit.SignIn(t, pool)
	_, err := pool.Exec(t.Context(), `UPDATE sessions SET user_id = $1 WHERE user_id = $2`, id, u.ID)
	require.NoError(t, err)

	require.Equal(t, http.StatusNoContent, change(t, h, u.Cookie, old, "senha-nova-1234").Code)
	require.Equal(t, 1, userstest.Count(t, pool,
		`SELECT count(*) FROM audit_events WHERE action = 'user.password_changed' AND actor_id = $1`, id))
	var after string
	require.NoError(t, pool.QueryRow(t.Context(),
		`SELECT coalesce(before::text, '') || coalesce(after::text, '') FROM audit_events WHERE action = 'user.password_changed' AND actor_id = $1`,
		id).Scan(&after))
	require.NotContains(t, after, "password")
}

func TestChangePassword_StoresArgon2idHash(t *testing.T) {
	pool := testkit.MigratedDB(t)
	h := userstest.Serve(t, pool, testkit.APIOptions{}, changepassword.Register)
	id := userstest.CreateUser(t, pool, "ana@x.com", old)
	u := testkit.SignIn(t, pool)
	_, err := pool.Exec(t.Context(), `UPDATE sessions SET user_id = $1 WHERE user_id = $2`, id, u.ID)
	require.NoError(t, err)

	require.Equal(t, http.StatusNoContent, change(t, h, u.Cookie, old, "senha-nova-1234").Code)
	var hash string
	require.NoError(t, pool.QueryRow(t.Context(), `SELECT password_hash FROM users WHERE id = $1`, id).Scan(&hash))
	require.Regexp(t, `^\$argon2id\$v=19\$m=65536,t=3,p=4\$[A-Za-z0-9+/]{22}\$[A-Za-z0-9+/]{43}$`, hash)
	require.NotContains(t, hash, "senha-nova-1234")
}

func errorLocations(t *testing.T, rec *httptest.ResponseRecorder) []string {
	t.Helper()
	var out []string
	for _, e := range testkit.JSON[struct {
		Errors []struct {
			Location string `json:"location"`
		} `json:"errors"`
	}](t, rec).Errors {
		out = append(out, e.Location)
	}
	return out
}

func TestChangePassword_Validation422(t *testing.T) {
	h, signIn := setup(t)
	current, other := signIn(), signIn()

	rec := change(t, h, current, "senha-errada-123", "senha-nova-1234")
	require.Equal(t, http.StatusUnprocessableEntity, rec.Code)
	require.Equal(t, []string{"body.current_password"}, errorLocations(t, rec))
	require.Equal(t, http.StatusOK, meStatus(t, h, current))
	require.Equal(t, http.StatusOK, meStatus(t, h, other))
	require.Equal(t, http.StatusNoContent, signInWith(t, h, old).Code)

	rec = change(t, h, current, old, strings.Repeat("n", 11))
	require.Equal(t, http.StatusUnprocessableEntity, rec.Code)
	require.Equal(t, []string{"body.new_password"}, errorLocations(t, rec))
}
