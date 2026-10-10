package main

import (
	"bytes"
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/luiszkm/template-go/internal/platform/auth"
	"github.com/luiszkm/template-go/internal/platform/config"
	"github.com/luiszkm/template-go/internal/platform/testkit"
)

const unreachableDB = "postgres://app:app@127.0.0.1:1/app?sslmode=disable&connect_timeout=5"

func TestNewServer_Timeouts(t *testing.T) {
	cfg, err := config.Load(map[string]string{"DATABASE_URL": "postgres://x"})
	require.NoError(t, err)
	srv := newServer(cfg, http.NotFoundHandler())
	require.Equal(t, 30*time.Second, srv.ReadTimeout)
	require.Equal(t, 30*time.Second, srv.WriteTimeout)
	require.Equal(t, 120*time.Second, srv.IdleTimeout)
	require.Equal(t, 10*time.Second, srv.ReadHeaderTimeout)

	cfg, err = config.Load(map[string]string{"DATABASE_URL": "postgres://x",
		"HTTP_READ_TIMEOUT": "5s", "HTTP_WRITE_TIMEOUT": "6s", "HTTP_IDLE_TIMEOUT": "7s"})
	require.NoError(t, err)
	srv = newServer(cfg, http.NotFoundHandler())
	require.Equal(t, 5*time.Second, srv.ReadTimeout)
	require.Equal(t, 6*time.Second, srv.WriteTimeout)
	require.Equal(t, 7*time.Second, srv.IdleTimeout)
}

func TestServe_InvalidDurationExits1(t *testing.T) {
	pool := testkit.MigratedDB(t)
	for _, name := range []string{"HTTP_READ_TIMEOUT", "HTTP_WRITE_TIMEOUT", "HTTP_IDLE_TIMEOUT", "SESSION_SWEEP_INTERVAL"} {
		var out, errb bytes.Buffer
		code := run(t.Context(), []string{"serve"}, map[string]string{"DATABASE_URL": testkit.URLOf(pool),
			"HTTP_ADDR": "127.0.0.1:0", name: "abc"}, nil, &out, &errb)
		require.Equal(t, 1, code, name)
		require.True(t, strings.HasPrefix(errb.String(), "config:"), "%s: %s", name, errb.String())
	}
}

func TestServe_SweepsExpiredSessions(t *testing.T) {
	pool := testkit.MigratedDB(t)
	old := testkit.SignIn(t, pool)
	_, err := pool.Exec(t.Context(), `UPDATE sessions SET created_at = now() - interval '2 hours' WHERE user_id = $1`, old.ID)
	require.NoError(t, err)
	fresh := testkit.SignIn(t, pool)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan int, 1)
	addr := freeAddr(t)
	go func() {
		var out, errb bytes.Buffer
		done <- run(ctx, []string{"serve"}, map[string]string{"DATABASE_URL": testkit.URLOf(pool), "HTTP_ADDR": addr,
			"SESSION_TTL": "1h"}, nil, &out, &errb)
	}()
	waitServing(t, "http://"+addr+"/healthz", done)

	count := func(token string) int {
		var n int
		require.NoError(t, pool.QueryRow(t.Context(), `SELECT count(*) FROM sessions WHERE token_hash = $1`,
			auth.HashToken(token)).Scan(&n))
		return n
	}
	require.Eventually(t, func() bool { return count(old.Cookie.Value) == 0 }, 5*time.Second, 20*time.Millisecond)
	require.Equal(t, 1, count(fresh.Cookie.Value))
	cancel()
	require.Equal(t, 0, <-done)
}

func TestServe_UnreachableDatabaseExits1(t *testing.T) {
	var out, errb bytes.Buffer
	code := run(t.Context(), []string{"serve"}, map[string]string{"DATABASE_URL": unreachableDB, "HTTP_ADDR": "127.0.0.1:0"},
		nil, &out, &errb)
	require.Equal(t, 1, code)
	require.Contains(t, errb.String(), "db: ping:")
	require.NotContains(t, out.String(), "listening")
}

func TestCreateAdmin_UnreachableDatabaseExits1(t *testing.T) {
	code, _, stderr := createAdminCmd(t, unreachableDB, "senha-longa-123\n", "--email", "ana@x.com", "--name", "Ana")
	require.Equal(t, 1, code)
	require.Contains(t, stderr, "db: ping:")
}
