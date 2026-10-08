package getuser_test

import (
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	getuser "github.com/luiszkm/template-go/internal/features/users/get_user"
	"github.com/luiszkm/template-go/internal/features/users/userstest"
	"github.com/luiszkm/template-go/internal/platform/testkit"
)

func TestGetUser_200(t *testing.T) {
	pool := testkit.MigratedDB(t)
	h := userstest.Serve(t, pool, testkit.APIOptions{}, getuser.Register)
	reader := testkit.SignIn(t, pool, "users:read")
	active := userstest.InsertUser(t, pool, "on@x.com", "x")
	off := userstest.InsertUser(t, pool, "off@x.com", "x")
	_, err := pool.Exec(t.Context(), `UPDATE users SET deactivated_at = now() WHERE id = $1`, off)
	require.NoError(t, err)

	for id, wantActive := range map[uuid.UUID]bool{active: true, off: false} {
		rec := testkit.Do(t, h, testkit.Request{Method: http.MethodGet, Path: "/api/v1/users/" + id.String(), Cookie: reader.Cookie})
		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
		body := testkit.JSON[map[string]any](t, rec)
		for _, key := range []string{"id", "email", "name", "active", "created_at", "deactivated_at"} {
			require.Contains(t, body, key)
		}
		require.Equal(t, id.String(), body["id"])
		require.Equal(t, wantActive, body["active"])
		if wantActive {
			require.Nil(t, body["deactivated_at"])
		} else {
			require.NotEmpty(t, body["deactivated_at"])
		}
	}
}

func TestGetUser_404And422(t *testing.T) {
	pool := testkit.MigratedDB(t)
	h := userstest.Serve(t, pool, testkit.APIOptions{}, getuser.Register)
	reader := testkit.SignIn(t, pool, "users:read")
	require.Equal(t, http.StatusNotFound,
		testkit.Do(t, h, testkit.Request{Method: http.MethodGet, Path: "/api/v1/users/" + uuid.NewString(), Cookie: reader.Cookie}).Code)
	require.Equal(t, http.StatusUnprocessableEntity,
		testkit.Do(t, h, testkit.Request{Method: http.MethodGet, Path: "/api/v1/users/abc", Cookie: reader.Cookie}).Code)
}
