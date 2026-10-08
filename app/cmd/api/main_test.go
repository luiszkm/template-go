package main

import (
	"bytes"
	"context"
	"database/sql"
	"io/fs"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/luiszkm/template-go/internal/platform/testkit"
	"github.com/luiszkm/template-go/migrations"
)

// C4
func TestServe_MissingDatabaseURLExits1(t *testing.T) {
	bin := filepath.Join(t.TempDir(), "api")
	if runtime.GOOS == "windows" {
		bin += ".exe"
	}
	build := exec.CommandContext(t.Context(), "go", "build", "-o", bin, ".")
	out, err := build.CombinedOutput()
	require.NoError(t, err, string(out))

	cmd := exec.CommandContext(t.Context(), bin, "serve")
	var env []string
	for _, kv := range os.Environ() {
		if !strings.HasPrefix(strings.ToUpper(kv), "DATABASE_URL=") {
			env = append(env, kv)
		}
	}
	cmd.Env = env
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err = cmd.Run()
	var exitErr *exec.ExitError
	require.ErrorAs(t, err, &exitErr)
	require.Equal(t, 1, exitErr.ExitCode())
	require.Contains(t, stderr.String(), "DATABASE_URL")
}

func newestMigrationVersion(t *testing.T) int64 {
	t.Helper()
	names, err := fs.Glob(migrations.FS, "*.sql")
	require.NoError(t, err)
	require.NotEmpty(t, names)
	var newest int64
	for _, n := range names {
		v, err := strconv.ParseInt(strings.SplitN(n, "_", 2)[0], 10, 64)
		require.NoError(t, err, n)
		newest = max(newest, v)
	}
	return newest
}

func gooseVersion(t *testing.T, url string) int64 {
	t.Helper()
	db, err := sql.Open("pgx", url)
	require.NoError(t, err)
	defer func() { _ = db.Close() }()
	var v int64
	require.NoError(t, db.QueryRowContext(t.Context(), `SELECT max(version_id) FROM goose_db_version`).Scan(&v))
	return v
}

func runCmd(t *testing.T, url string, args ...string) (int, string) {
	t.Helper()
	var out, errb bytes.Buffer
	code := run(context.Background(), args, map[string]string{"DATABASE_URL": url}, &out, &errb)
	return code, errb.String()
}

// C7
func TestMigrateUp_AppliesPending(t *testing.T) {
	pg := testkit.StartPostgres(t)
	code, stderr := runCmd(t, pg.URL, "migrate", "up")
	require.Equal(t, 0, code, stderr)
	require.Equal(t, newestMigrationVersion(t), gooseVersion(t, pg.URL))
}

// C8
func TestMigrateUp_RerunIsNoop(t *testing.T) {
	pg := testkit.StartPostgres(t)
	code, stderr := runCmd(t, pg.URL, "migrate", "up")
	require.Equal(t, 0, code, stderr)
	before := gooseVersion(t, pg.URL)

	db, err := sql.Open("pgx", pg.URL)
	require.NoError(t, err)
	defer func() { _ = db.Close() }()
	var rowsBefore int
	require.NoError(t, db.QueryRowContext(t.Context(), `SELECT count(*) FROM goose_db_version`).Scan(&rowsBefore))

	code, stderr = runCmd(t, pg.URL, "migrate", "up")
	require.Equal(t, 0, code, stderr)
	require.Equal(t, before, gooseVersion(t, pg.URL))
	var rowsAfter int
	require.NoError(t, db.QueryRowContext(t.Context(), `SELECT count(*) FROM goose_db_version`).Scan(&rowsAfter))
	require.Equal(t, rowsBefore, rowsAfter)
}

// C9
func TestServe_DoesNotMigrate(t *testing.T) {
	pg := testkit.StartPostgres(t)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan int, 1)
	addr := freeAddr(t)
	go func() {
		var out, errb bytes.Buffer
		done <- run(ctx, []string{"serve"}, map[string]string{"DATABASE_URL": pg.URL, "HTTP_ADDR": addr}, &out, &errb)
	}()
	waitServing(t, "http://"+addr+"/healthz", done)
	cancel()
	require.Equal(t, 0, <-done)

	db, err := sql.Open("pgx", pg.URL)
	require.NoError(t, err)
	defer func() { _ = db.Close() }()
	var exists bool
	require.NoError(t, db.QueryRowContext(t.Context(), `SELECT to_regclass('public.goose_db_version') IS NOT NULL`).Scan(&exists))
	require.False(t, exists)
}

// freeAddr returns a loopback address with a port that was free a moment ago.
func freeAddr(t *testing.T) string {
	t.Helper()
	var lc net.ListenConfig
	ln, err := lc.Listen(t.Context(), "tcp", "127.0.0.1:0")
	require.NoError(t, err)
	addr := ln.Addr().String()
	require.NoError(t, ln.Close())
	return addr
}

// waitServing polls url until it answers 200, failing if the server exits or the deadline passes.
func waitServing(t *testing.T, url string, done <-chan int) {
	t.Helper()
	client := &http.Client{Timeout: time.Second}
	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		select {
		case code := <-done:
			t.Fatalf("serve exited with %d before serving", code)
		default:
		}
		req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, url, nil)
		require.NoError(t, err)
		resp, err := client.Do(req)
		if err == nil {
			_ = resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return
			}
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatalf("%s did not answer 200 within the deadline", url)
}

// C62
func TestRun_UnknownCommandExits2(t *testing.T) {
	var out, errb bytes.Buffer
	code := run(context.Background(), []string{"frobnicate"}, map[string]string{}, &out, &errb)
	require.Equal(t, 2, code)
	require.Contains(t, errb.String(), "usage: api serve | api migrate up | api openapi")
}

// C63
func TestMigrateUp_UnreachableDatabaseExits1(t *testing.T) {
	code, stderr := runCmd(t, "postgres://app:app@127.0.0.1:1/app?sslmode=disable&connect_timeout=5", "migrate", "up")
	require.Equal(t, 1, code, stderr)
	require.Contains(t, stderr, "migrate up:")
}
