package createuser_test

import (
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"

	createuser "github.com/luiszkm/template-go/internal/features/users/create_user"
	"github.com/luiszkm/template-go/internal/features/users/userstest"
	"github.com/luiszkm/template-go/internal/platform/testkit"
)

type fixture struct {
	pool  *pgxpool.Pool
	h     http.Handler
	admin testkit.User
}

func setup(t *testing.T) fixture {
	t.Helper()
	pool := testkit.MigratedDB(t)
	return fixture{pool: pool, h: userstest.Serve(t, pool, testkit.APIOptions{}, createuser.Register),
		admin: testkit.SignIn(t, pool, "users:create")}
}

func (f fixture) create(t *testing.T, body map[string]string) *httptest.ResponseRecorder {
	t.Helper()
	return testkit.Do(t, f.h, testkit.Request{Method: http.MethodPost, Path: "/api/v1/users", Body: body, Cookie: f.admin.Cookie})
}

func valid(email string) map[string]string {
	return map[string]string{"email": email, "name": "Bia", "password": "doze-chars12"}
}

func TestCreateUser_201(t *testing.T) {
	f := setup(t)
	rec := f.create(t, valid("B@X.com "))
	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())

	body := testkit.JSON[map[string]any](t, rec)
	keys := make([]string, 0, len(body))
	for k := range body {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	require.Equal(t, []string{"active", "created_at", "email", "id", "name"}, keys)
	require.Equal(t, "b@x.com", body["email"])
	require.Equal(t, "Bia", body["name"])
	require.Equal(t, true, body["active"])
	require.Equal(t, 1, userstest.Count(t, f.pool,
		`SELECT count(*) FROM audit_events WHERE action = 'user.created' AND actor_id = $1 AND resource_id = $2`, f.admin.ID, body["id"]))
}

func TestCreateUser_StoresArgon2idHash(t *testing.T) {
	f := setup(t)
	require.Equal(t, http.StatusCreated, f.create(t, valid("b@x.com")).Code)
	var hash string
	require.NoError(t, f.pool.QueryRow(t.Context(), `SELECT password_hash FROM users WHERE email = 'b@x.com'`).Scan(&hash))
	require.Regexp(t, `^\$argon2id\$v=19\$m=65536,t=3,p=4\$[A-Za-z0-9+/]{22}\$[A-Za-z0-9+/]{43}$`, hash)
	require.NotContains(t, hash, "doze-chars12")
	var audit string
	require.NoError(t, f.pool.QueryRow(t.Context(), `SELECT after::text FROM audit_events WHERE action = 'user.created'`).Scan(&audit))
	require.NotContains(t, audit, "password")
}

func TestCreateUser_DuplicateEmail409(t *testing.T) {
	f := setup(t)
	require.Equal(t, http.StatusCreated, f.create(t, valid("B@x.com")).Code)
	before := userstest.Count(t, f.pool, `SELECT count(*) FROM users`)
	rec := f.create(t, valid(" b@x.com"))
	require.Equal(t, http.StatusConflict, rec.Code, rec.Body.String())
	require.Equal(t, "application/problem+json", rec.Header().Get("Content-Type"))
	require.Equal(t, before, userstest.Count(t, f.pool, `SELECT count(*) FROM users`))
}

func TestCreateUser_ConcurrentSameEmail(t *testing.T) {
	f := setup(t)
	codes := make([]int, 10)
	var wg sync.WaitGroup
	for i := range codes {
		wg.Go(func() { codes[i] = f.create(t, valid("race@x.com")).Code })
	}
	wg.Wait()
	slices.Sort(codes)
	require.Equal(t, []int{201, 409, 409, 409, 409, 409, 409, 409, 409, 409}, codes)
	require.Equal(t, 1, userstest.Count(t, f.pool, `SELECT count(*) FROM users WHERE email = 'race@x.com'`))
}

func TestCreateUser_Validation(t *testing.T) {
	f := setup(t)
	rejected := []struct {
		name, field string
		body        map[string]string
	}{
		{"email not an address", "body.email", map[string]string{"email": "nope", "name": "Bia", "password": "doze-chars12"}},
		{"name empty", "body.name", map[string]string{"email": "a1@x.com", "name": "", "password": "doze-chars12"}},
		{"name 101", "body.name", map[string]string{"email": "a2@x.com", "name": strings.Repeat("n", 101), "password": "doze-chars12"}},
		{"password 11", "body.password", map[string]string{"email": "a3@x.com", "name": "Bia", "password": strings.Repeat("p", 11)}},
		{"password 129", "body.password", map[string]string{"email": "a4@x.com", "name": "Bia", "password": strings.Repeat("p", 129)}},
	}
	for _, c := range rejected {
		t.Run(c.name, func(t *testing.T) {
			rec := f.create(t, c.body)
			require.Equal(t, http.StatusUnprocessableEntity, rec.Code, rec.Body.String())
			var locations []string
			for _, e := range testkit.JSON[struct {
				Errors []struct {
					Location string `json:"location"`
				} `json:"errors"`
			}](t, rec).Errors {
				locations = append(locations, e.Location)
			}
			require.Contains(t, locations, c.field)
		})
	}
	accepted := []map[string]string{
		{"email": "e1@x.com", "name": "B", "password": "doze-chars12"},
		{"email": "e2@x.com", "name": strings.Repeat("n", 100), "password": "doze-chars12"},
		{"email": "e3@x.com", "name": "Bia", "password": strings.Repeat("p", 12)},
		{"email": "e4@x.com", "name": "Bia", "password": strings.Repeat("p", 128)},
	}
	for _, body := range accepted {
		require.Equal(t, http.StatusCreated, f.create(t, body).Code, body["email"])
	}
}
