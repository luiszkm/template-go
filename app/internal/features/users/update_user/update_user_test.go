package updateuser_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"

	updateuser "github.com/luiszkm/template-go/internal/features/users/update_user"
	"github.com/luiszkm/template-go/internal/features/users/userstest"
	"github.com/luiszkm/template-go/internal/platform/testkit"
)

type fixture struct {
	pool   *pgxpool.Pool
	h      http.Handler
	editor testkit.User
}

func setup(t *testing.T) fixture {
	t.Helper()
	pool := testkit.MigratedDB(t)
	return fixture{pool: pool, h: userstest.Serve(t, pool, testkit.APIOptions{}, updateuser.Register),
		editor: testkit.SignIn(t, pool, "users:update")}
}

func (f fixture) patch(t *testing.T, path string, body any) *httptest.ResponseRecorder {
	t.Helper()
	return testkit.Do(t, f.h, testkit.Request{Method: http.MethodPatch, Path: path, Body: body, Cookie: f.editor.Cookie})
}

func (f fixture) stored(t *testing.T, id uuid.UUID) (string, string) {
	t.Helper()
	var email, name string
	require.NoError(t, f.pool.QueryRow(t.Context(), `SELECT email, name FROM users WHERE id = $1`, id).Scan(&email, &name))
	return email, name
}

func (f fixture) lastAudit(t *testing.T, id uuid.UUID) (map[string]any, map[string]any) {
	t.Helper()
	var before, after []byte
	require.NoError(t, f.pool.QueryRow(t.Context(),
		`SELECT before, after FROM audit_events WHERE action = 'user.updated' AND resource_id = $1 ORDER BY id DESC LIMIT 1`,
		id.String()).Scan(&before, &after))
	var b, a map[string]any
	require.NoError(t, json.Unmarshal(before, &b))
	require.NoError(t, json.Unmarshal(after, &a))
	return b, a
}

func changedKeys(before, after map[string]any) []string {
	var keys []string
	for k, v := range after {
		if before[k] != v {
			keys = append(keys, k)
		}
	}
	return keys
}

func TestUpdateUser_PartialFields(t *testing.T) {
	f := setup(t)
	id := userstest.InsertUser(t, f.pool, "ana@x.com", "x")
	path := "/api/v1/users/" + id.String()

	rec := f.patch(t, path, map[string]string{"name": "Ana Maria"})
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	require.Equal(t, "Ana Maria", testkit.JSON[map[string]any](t, rec)["name"])
	email, name := f.stored(t, id)
	require.Equal(t, []string{"ana@x.com", "Ana Maria"}, []string{email, name})
	require.ElementsMatch(t, []string{"name"}, changedKeys(f.lastAudit(t, id)))

	rec = f.patch(t, path, map[string]string{"email": "Ana2@X.com"})
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	email, name = f.stored(t, id)
	require.Equal(t, []string{"ana2@x.com", "Ana Maria"}, []string{email, name})
	require.ElementsMatch(t, []string{"email"}, changedKeys(f.lastAudit(t, id)))

	rec = f.patch(t, path, map[string]string{"email": "ana3@x.com", "name": "Ana"})
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	body := testkit.JSON[map[string]any](t, rec)
	require.Equal(t, "ana3@x.com", body["email"])
	require.Equal(t, "Ana", body["name"])
	require.ElementsMatch(t, []string{"email", "name"}, changedKeys(f.lastAudit(t, id)))
	require.Equal(t, 3, userstest.Count(t, f.pool, `SELECT count(*) FROM audit_events WHERE action = 'user.updated' AND actor_id = $1`, f.editor.ID))
}

func TestUpdateUser_ConflictAndValidation(t *testing.T) {
	f := setup(t)
	id := userstest.InsertUser(t, f.pool, "ana@x.com", "x")
	userstest.InsertUser(t, f.pool, "bia@x.com", "x")
	path := "/api/v1/users/" + id.String()

	require.Equal(t, http.StatusConflict, f.patch(t, path, map[string]string{"email": "BIA@x.com"}).Code)
	email, _ := f.stored(t, id)
	require.Equal(t, "ana@x.com", email)

	require.Equal(t, http.StatusOK, f.patch(t, path, map[string]string{"email": "ana@x.com"}).Code)

	for name, body := range map[string]any{
		"invalid email": map[string]string{"email": "nope"},
		"empty name":    map[string]string{"name": ""},
		"empty body":    map[string]string{},
	} {
		require.Equal(t, http.StatusUnprocessableEntity, f.patch(t, path, body).Code, name)
	}
}

func TestUpdateUser_404And422(t *testing.T) {
	f := setup(t)
	require.Equal(t, http.StatusNotFound, f.patch(t, "/api/v1/users/"+uuid.NewString(), map[string]string{"name": "X"}).Code)
	require.Equal(t, http.StatusUnprocessableEntity, f.patch(t, "/api/v1/users/abc", map[string]string{"name": "X"}).Code)
}
