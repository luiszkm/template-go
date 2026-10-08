package main

import (
	"bytes"
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/luiszkm/template-go/internal/platform/testkit"
)

func createAdminCmd(t *testing.T, url, stdin string, args ...string) (int, string, string) {
	t.Helper()
	var out, errb bytes.Buffer
	code := run(context.Background(), append([]string{"users", "create-admin"}, args...),
		map[string]string{"DATABASE_URL": url}, strings.NewReader(stdin), &out, &errb)
	return code, out.String(), errb.String()
}

func TestCreateAdmin_CreatesActiveAdmin(t *testing.T) {
	pool := testkit.MigratedDB(t)
	code, stdout, stderr := createAdminCmd(t, testkit.URLOf(pool), "senha-longa-123\n", "--email", "  Ana@X.com ", "--name", "Ana")
	require.Equal(t, 0, code, stderr)
	id, err := uuid.Parse(strings.TrimSpace(stdout))
	require.NoError(t, err, stdout)

	var email, hash string
	var active bool
	require.NoError(t, pool.QueryRow(t.Context(),
		`SELECT email, password_hash, deactivated_at IS NULL FROM users WHERE id = $1`, id).Scan(&email, &hash, &active))
	require.Equal(t, "ana@x.com", email)
	require.True(t, active)
	require.Regexp(t, `^\$argon2id\$v=19\$m=65536,t=3,p=4\$[A-Za-z0-9+/]{22}\$[A-Za-z0-9+/]{43}$`, hash)

	var roles int
	require.NoError(t, pool.QueryRow(t.Context(),
		`SELECT count(*) FROM user_roles ur JOIN roles r ON r.id = ur.role_id WHERE ur.user_id = $1 AND r.name = 'admin'`, id).Scan(&roles))
	require.Equal(t, 1, roles)

	var events int
	require.NoError(t, pool.QueryRow(t.Context(),
		`SELECT count(*) FROM audit_events WHERE action = 'user.created' AND actor_id IS NULL AND resource_id = $1`, id.String()).Scan(&events))
	require.Equal(t, 1, events)
}

func TestCreateAdmin_ExistingEmailExits1(t *testing.T) {
	pool := testkit.MigratedDB(t)
	code, _, stderr := createAdminCmd(t, testkit.URLOf(pool), "senha-longa-123\n", "--email", "ana@x.com", "--name", "Ana")
	require.Equal(t, 0, code, stderr)

	code, _, stderr = createAdminCmd(t, testkit.URLOf(pool), "senha-longa-123\n", "--email", "ANA@x.com", "--name", "Ana")
	require.Equal(t, 1, code)
	require.Contains(t, stderr, "already exists")
	var users int
	require.NoError(t, pool.QueryRow(t.Context(), `SELECT count(*) FROM users`).Scan(&users))
	require.Equal(t, 1, users)
}

func TestCreateAdmin_InvalidInputExits2(t *testing.T) {
	pool := testkit.MigratedDB(t)
	cases := []struct {
		name, stdin string
		args        []string
	}{
		{"no email", "senha-longa-123\n", []string{"--name", "Ana"}},
		{"no name", "senha-longa-123\n", []string{"--email", "a@x.com"}},
		{"password 11", strings.Repeat("p", 11) + "\n", []string{"--email", "a@x.com", "--name", "Ana"}},
		{"password 129", strings.Repeat("p", 129) + "\n", []string{"--email", "a@x.com", "--name", "Ana"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			code, _, stderr := createAdminCmd(t, testkit.URLOf(pool), c.stdin, c.args...)
			require.Equal(t, 2, code)
			require.True(t, strings.HasPrefix(stderr, "usage:"), stderr)
		})
	}
	var users int
	require.NoError(t, pool.QueryRow(t.Context(), `SELECT count(*) FROM users`).Scan(&users))
	require.Zero(t, users)
}

func TestServe_AppliesSessionConfig(t *testing.T) {
	pool := testkit.MigratedDB(t)
	url := testkit.URLOf(pool)
	code, _, stderr := createAdminCmd(t, url, "senha-longa-123\n", "--email", "ana@x.com", "--name", "Ana")
	require.Equal(t, 0, code, stderr)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan int, 1)
	addr := freeAddr(t)
	go func() {
		var out, errb bytes.Buffer
		done <- run(ctx, []string{"serve"}, map[string]string{"DATABASE_URL": url, "HTTP_ADDR": addr,
			"SESSION_TTL": "1h", "COOKIE_SECURE": "false"}, nil, &out, &errb)
	}()
	waitServing(t, "http://"+addr+"/healthz", done)

	req, err := http.NewRequestWithContext(t.Context(), http.MethodPost, "http://"+addr+"/api/v1/users/session",
		strings.NewReader(`{"email":"ana@x.com","password":"senha-longa-123"}`))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	_ = res.Body.Close()
	cancel()
	require.Equal(t, 0, <-done)

	require.Equal(t, http.StatusNoContent, res.StatusCode)
	cookie := res.Header.Get("Set-Cookie")
	require.Contains(t, cookie, "Max-Age=3600")
	require.NotContains(t, cookie, "Secure")
}
