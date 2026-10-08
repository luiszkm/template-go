package login_test

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"slices"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/argon2"

	"github.com/luiszkm/template-go/internal/features/users/login"
	"github.com/luiszkm/template-go/internal/features/users/me"
	"github.com/luiszkm/template-go/internal/features/users/password"
	"github.com/luiszkm/template-go/internal/features/users/userstest"
	"github.com/luiszkm/template-go/internal/platform/deps"
	"github.com/luiszkm/template-go/internal/platform/testkit"
)

const pw = "senha-longa-123"

type countingVerifier struct{ calls atomic.Int32 }

func (c *countingVerifier) verify(p, hash string) (bool, error) {
	c.calls.Add(1)
	return password.Verify(p, hash)
}

func (c *countingVerifier) register(api huma.API, d deps.Deps) error {
	return login.RegisterWithVerifier(api, d, c.verify)
}

func serve(t *testing.T, pool *pgxpool.Pool, o testkit.APIOptions) http.Handler {
	t.Helper()
	return userstest.Serve(t, pool, o, login.Register, me.Register)
}

type attempt struct {
	email, password string
	cookie          *http.Cookie
	remoteAddr      string
	xff             string
}

func post(t *testing.T, h http.Handler, a attempt) *httptest.ResponseRecorder {
	t.Helper()
	r := testkit.Request{Method: http.MethodPost, Path: "/api/v1/users/session",
		Body: map[string]string{"email": a.email, "password": a.password}, Cookie: a.cookie, RemoteAddr: a.remoteAddr}
	if a.xff != "" {
		r.Header = http.Header{"X-Forwarded-For": {a.xff}}
	}
	return testkit.Do(t, h, r)
}

func meStatus(t *testing.T, h http.Handler, c *http.Cookie) int {
	t.Helper()
	return testkit.Do(t, h, testkit.Request{Method: http.MethodGet, Path: "/api/v1/users/me", Cookie: c}).Code
}

func sessionCookie(t *testing.T, rec *httptest.ResponseRecorder) *http.Cookie {
	t.Helper()
	return userstest.SessionCookie(t, rec.Result())
}

func legacyHash(t *testing.T, p string) string {
	t.Helper()
	salt := make([]byte, 16)
	_, err := rand.Read(salt)
	require.NoError(t, err)
	enc := base64.RawStdEncoding
	return fmt.Sprintf("$argon2id$v=19$m=19456,t=2,p=1$%s$%s", enc.EncodeToString(salt),
		enc.EncodeToString(argon2.IDKey([]byte(p), salt, 2, 19456, 1, 32)))
}

func TestLogin_RehashesOutdatedHash(t *testing.T) {
	pool := testkit.MigratedDB(t)
	h := serve(t, pool, testkit.APIOptions{})
	id := userstest.InsertUser(t, pool, "old@x.com", legacyHash(t, pw))

	require.Equal(t, http.StatusNoContent, post(t, h, attempt{email: "old@x.com", password: pw}).Code)
	var stored string
	require.NoError(t, pool.QueryRow(t.Context(), `SELECT password_hash FROM users WHERE id = $1`, id).Scan(&stored))
	require.True(t, strings.HasPrefix(stored, "$argon2id$v=19$m=65536,t=3,p=4$"), stored)
	require.Equal(t, http.StatusNoContent, post(t, h, attempt{email: "old@x.com", password: pw}).Code)
}

func TestLogin_SetsSessionCookie(t *testing.T) {
	for _, secure := range []bool{true, false} {
		t.Run("secure="+strconv.FormatBool(secure), func(t *testing.T) {
			pool := testkit.MigratedDB(t)
			h := serve(t, pool, testkit.APIOptions{CookieSecure: secure})
			id := userstest.CreateUser(t, pool, "ana@x.com", pw)

			rec := post(t, h, attempt{email: "ana@x.com", password: pw})
			require.Equal(t, http.StatusNoContent, rec.Code)
			headers := rec.Header().Values("Set-Cookie")
			require.Len(t, headers, 1)
			attrs := strings.Split(headers[0], "; ")
			require.Regexp(t, `^session=[A-Za-z0-9_-]{43}$`, attrs[0])
			for _, want := range []string{"Path=/", "Max-Age=43200", "HttpOnly", "SameSite=Lax"} {
				require.Contains(t, attrs, want)
			}
			require.Equal(t, secure, slices.Contains(attrs, "Secure"))
			require.Equal(t, 1, userstest.Count(t, pool,
				`SELECT count(*) FROM audit_events WHERE action = 'session.created' AND actor_id = $1`, id))
		})
	}
}

func TestLogin_StoresOnlyTokenHash(t *testing.T) {
	pool := testkit.MigratedDB(t)
	h := serve(t, pool, testkit.APIOptions{})
	userstest.CreateUser(t, pool, "ana@x.com", pw)

	token := sessionCookie(t, post(t, h, attempt{email: "ana@x.com", password: pw})).Value
	sum := sha256.Sum256([]byte(token))
	var hashHex, userID, createdAt string
	require.NoError(t, pool.QueryRow(t.Context(),
		`SELECT encode(token_hash, 'hex'), user_id::text, created_at::text FROM sessions`).Scan(&hashHex, &userID, &createdAt))
	require.Equal(t, fmt.Sprintf("%x", sum), hashHex)
	for _, column := range []string{hashHex, userID, createdAt} {
		require.NotEqual(t, token, column)
		require.NotContains(t, column, token)
	}
	require.Equal(t, 0, userstest.Count(t, pool,
		`SELECT count(*) FROM sessions WHERE token_hash = convert_to($1, 'UTF8')`, token))
}

func TestLogin_RotatesExistingSession(t *testing.T) {
	pool := testkit.MigratedDB(t)
	h := serve(t, pool, testkit.APIOptions{})
	userstest.CreateUser(t, pool, "ana@x.com", pw)

	first := sessionCookie(t, post(t, h, attempt{email: "ana@x.com", password: pw}))
	second := sessionCookie(t, post(t, h, attempt{email: "ana@x.com", password: pw, cookie: first}))
	require.NotEqual(t, first.Value, second.Value)
	sum := sha256.Sum256([]byte(first.Value))
	require.Equal(t, 0, userstest.Count(t, pool, `SELECT count(*) FROM sessions WHERE token_hash = $1`, sum[:]))
	require.Equal(t, http.StatusUnauthorized, meStatus(t, h, first))
	require.Equal(t, http.StatusOK, meStatus(t, h, second))
}

func problemWithoutRequest(t *testing.T, rec *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	body := testkit.JSON[map[string]any](t, rec)
	delete(body, "request_id")
	delete(body, "instance")
	return body
}

func TestLogin_FailuresAreIndistinguishable(t *testing.T) {
	pool := testkit.MigratedDB(t)
	h := serve(t, pool, testkit.APIOptions{})
	userstest.CreateUser(t, pool, "ana@x.com", pw)
	off := userstest.CreateUser(t, pool, "off@x.com", pw)
	_, err := pool.Exec(t.Context(), `UPDATE users SET deactivated_at = now() WHERE id = $1`, off)
	require.NoError(t, err)

	unknown := post(t, h, attempt{email: "nobody@x.com", password: pw, remoteAddr: "192.0.2.1:1"})
	wrong := post(t, h, attempt{email: "ana@x.com", password: "senha-errada-123", remoteAddr: "192.0.2.2:1"})
	deactivated := post(t, h, attempt{email: "off@x.com", password: pw, remoteAddr: "192.0.2.3:1"})
	for _, rec := range []*httptest.ResponseRecorder{unknown, wrong, deactivated} {
		require.Equal(t, http.StatusUnauthorized, rec.Code)
		require.Equal(t, "invalid email or password", testkit.JSON[map[string]any](t, rec)["detail"])
	}
	require.Equal(t, problemWithoutRequest(t, unknown), problemWithoutRequest(t, wrong))
	require.Equal(t, problemWithoutRequest(t, unknown), problemWithoutRequest(t, deactivated))
}

func TestLogin_UnknownEmailVerifiesDummyHash(t *testing.T) {
	pool := testkit.MigratedDB(t)
	var hashes []string
	verify := func(p, hash string) (bool, error) {
		hashes = append(hashes, hash)
		return password.Verify(p, hash)
	}
	h := userstest.Serve(t, pool, testkit.APIOptions{}, func(api huma.API, d deps.Deps) error {
		return login.RegisterWithVerifier(api, d, verify)
	})

	require.Equal(t, http.StatusUnauthorized, post(t, h, attempt{email: "nobody@x.com", password: pw}).Code)
	require.Equal(t, []string{password.DummyHash()}, hashes)
}

func TestLogin_InvalidBody422(t *testing.T) {
	pool := testkit.MigratedDB(t)
	h := serve(t, pool, testkit.APIOptions{})
	bodies := map[string]any{
		"email not an address": map[string]string{"email": "nope", "password": pw},
		"email absent":         map[string]string{"password": pw},
		"password empty":       map[string]string{"email": "ana@x.com", "password": ""},
		"password 129 chars":   map[string]string{"email": "ana@x.com", "password": strings.Repeat("a", 129)},
	}
	for name, body := range bodies {
		t.Run(name, func(t *testing.T) {
			rec := testkit.Do(t, h, testkit.Request{Method: http.MethodPost, Path: "/api/v1/users/session", Body: body})
			require.Equal(t, http.StatusUnprocessableEntity, rec.Code, rec.Body.String())
		})
	}
}

func requireRetryAfter(t *testing.T, rec *httptest.ResponseRecorder) {
	t.Helper()
	require.Equal(t, http.StatusTooManyRequests, rec.Code, rec.Body.String())
	secs, err := strconv.Atoi(rec.Header().Get("Retry-After"))
	require.NoError(t, err)
	require.GreaterOrEqual(t, secs, 1)
	require.LessOrEqual(t, secs, 900)
}

func TestLogin_RateLimitPerEmail(t *testing.T) {
	pool := testkit.MigratedDB(t)
	v := &countingVerifier{}
	h := userstest.Serve(t, pool, testkit.APIOptions{}, v.register)
	userstest.CreateUser(t, pool, "x@y.com", pw)
	userstest.CreateUser(t, pool, "four@y.com", pw)

	for _, email := range []string{"x@y.com", "ghost@y.com"} {
		for i := range 5 {
			rec := post(t, h, attempt{email: email, password: "senha-errada-123", remoteAddr: fmt.Sprintf("192.0.2.%d:1", i+1)})
			require.Equal(t, http.StatusUnauthorized, rec.Code)
		}
		before := v.calls.Load()
		requireRetryAfter(t, post(t, h, attempt{email: email, password: pw, remoteAddr: "192.0.2.99:1"}))
		require.Equal(t, before, v.calls.Load(), "password must not be verified once limited")
	}

	for i := range 4 {
		require.Equal(t, http.StatusUnauthorized,
			post(t, h, attempt{email: "four@y.com", password: "senha-errada-123", remoteAddr: fmt.Sprintf("198.51.100.%d:1", i+1)}).Code)
	}
	require.Equal(t, http.StatusNoContent, post(t, h, attempt{email: "four@y.com", password: pw, remoteAddr: "198.51.100.50:1"}).Code)
}

func TestLogin_RateLimitPerIP(t *testing.T) {
	pool := testkit.MigratedDB(t)
	h := serve(t, pool, testkit.APIOptions{})
	userstest.CreateUser(t, pool, "ana@x.com", pw)

	for i := range 20 {
		rec := post(t, h, attempt{email: fmt.Sprintf("u%d@x.com", i), password: "senha-errada-123", remoteAddr: "203.0.113.1:1"})
		require.Equal(t, http.StatusUnauthorized, rec.Code)
	}
	requireRetryAfter(t, post(t, h, attempt{email: "ana@x.com", password: pw, remoteAddr: "203.0.113.1:2"}))
	require.Equal(t, http.StatusNoContent, post(t, h, attempt{email: "ana@x.com", password: pw, remoteAddr: "203.0.113.2:1"}).Code)

	for i := range 20 {
		rec := post(t, h, attempt{email: fmt.Sprintf("s%d@x.com", i), password: "senha-errada-123",
			remoteAddr: "203.0.113.3:1", xff: fmt.Sprintf("10.9.9.%d", i)})
		require.Equal(t, http.StatusUnauthorized, rec.Code)
	}
	requireRetryAfter(t, post(t, h, attempt{email: "ana@x.com", password: pw, remoteAddr: "203.0.113.3:1", xff: "10.9.9.200"}))
}

func insertAttempts(t *testing.T, pool *pgxpool.Pool, kind, key string, n int, age time.Duration) {
	t.Helper()
	for range n {
		_, err := pool.Exec(t.Context(), `INSERT INTO login_attempts (kind, key, at) VALUES ($1, $2, now() - $3::interval)`,
			kind, key, age.String())
		require.NoError(t, err)
	}
}

func TestLogin_RateLimitWindowIs15Minutes(t *testing.T) {
	pool := testkit.MigratedDB(t)
	h := serve(t, pool, testkit.APIOptions{})
	userstest.CreateUser(t, pool, "old@x.com", pw)
	userstest.CreateUser(t, pool, "recent@x.com", pw)

	insertAttempts(t, pool, "email", "old@x.com", 5, 16*time.Minute)
	insertAttempts(t, pool, "email", "recent@x.com", 5, 14*time.Minute)
	require.Equal(t, http.StatusNoContent, post(t, h, attempt{email: "old@x.com", password: pw}).Code)
	requireRetryAfter(t, post(t, h, attempt{email: "recent@x.com", password: pw}))

	insertAttempts(t, pool, "ip", "198.51.100.7", 20, 16*time.Minute)
	require.Equal(t, http.StatusNoContent, post(t, h, attempt{email: "old@x.com", password: pw, remoteAddr: "198.51.100.7:1"}).Code)
	insertAttempts(t, pool, "ip", "198.51.100.8", 20, 14*time.Minute)
	requireRetryAfter(t, post(t, h, attempt{email: "old@x.com", password: pw, remoteAddr: "198.51.100.8:1"}))
}

func TestLogin_SuccessClearsEmailFailures(t *testing.T) {
	pool := testkit.MigratedDB(t)
	h := serve(t, pool, testkit.APIOptions{})
	userstest.CreateUser(t, pool, "ana@x.com", pw)
	fail := func() {
		for i := range 4 {
			require.Equal(t, http.StatusUnauthorized,
				post(t, h, attempt{email: "ana@x.com", password: "senha-errada-123", remoteAddr: fmt.Sprintf("192.0.2.%d:1", i+1)}).Code)
		}
	}
	fail()
	require.Equal(t, http.StatusNoContent, post(t, h, attempt{email: "ana@x.com", password: pw}).Code)
	fail()
	require.Equal(t, http.StatusNoContent, post(t, h, attempt{email: "ana@x.com", password: pw}).Code)
}

func TestLogin_FailureLogsWarnWithoutEmail(t *testing.T) {
	pool := testkit.MigratedDB(t)
	var buf bytes.Buffer
	log := slog.New(slog.NewJSONHandler(&buf, nil))
	h := serve(t, pool, testkit.APIOptions{Logger: log})

	rec := post(t, h, attempt{email: "secret.person@x.com", password: pw, remoteAddr: "192.0.2.77:1"})
	require.Equal(t, http.StatusUnauthorized, rec.Code)

	var warns []map[string]any
	for line := range strings.Lines(buf.String()) {
		var entry map[string]any
		require.NoError(t, json.Unmarshal([]byte(line), &entry))
		if entry["level"] == "WARN" {
			warns = append(warns, entry)
		}
	}
	require.Len(t, warns, 1)
	require.Equal(t, rec.Header().Get("X-Request-ID"), warns[0]["request_id"])
	require.Equal(t, "192.0.2.77", warns[0]["ip"])
	require.NotContains(t, buf.String(), "secret.person")
}

func TestLogin_PrunesOldAttempts(t *testing.T) {
	pool := testkit.MigratedDB(t)
	h := serve(t, pool, testkit.APIOptions{})
	insertAttempts(t, pool, "email", "stale@x.com", 3, 20*time.Minute)

	require.Equal(t, http.StatusUnauthorized, post(t, h, attempt{email: "nobody@x.com", password: pw}).Code)
	require.Equal(t, 0, userstest.Count(t, pool, `SELECT count(*) FROM login_attempts WHERE at <= now() - interval '15 minutes'`))
}
