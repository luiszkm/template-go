package main

import (
	"bytes"
	"context"
	"database/sql"
	"io/fs"
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
	build := exec.Command("go", "build", "-o", bin, ".")
	out, err := build.CombinedOutput()
	require.NoError(t, err, string(out))

	cmd := exec.Command(bin, "serve")
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
	require.NoError(t, db.QueryRow(`SELECT max(version_id) FROM goose_db_version`).Scan(&v))
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
	require.NoError(t, db.QueryRow(`SELECT count(*) FROM goose_db_version`).Scan(&rowsBefore))

	code, stderr = runCmd(t, pg.URL, "migrate", "up")
	require.Equal(t, 0, code, stderr)
	require.Equal(t, before, gooseVersion(t, pg.URL))
	var rowsAfter int
	require.NoError(t, db.QueryRow(`SELECT count(*) FROM goose_db_version`).Scan(&rowsAfter))
	require.Equal(t, rowsBefore, rowsAfter)
}

// C9
func TestServe_DoesNotMigrate(t *testing.T) {
	pg := testkit.StartPostgres(t)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan int, 1)
	go func() {
		var out, errb bytes.Buffer
		done <- run(ctx, []string{"serve"}, map[string]string{"DATABASE_URL": pg.URL, "HTTP_ADDR": "127.0.0.1:0"}, &out, &errb)
	}()
	time.Sleep(time.Second)
	cancel()
	require.Equal(t, 0, <-done)

	db, err := sql.Open("pgx", pg.URL)
	require.NoError(t, err)
	defer func() { _ = db.Close() }()
	var exists bool
	require.NoError(t, db.QueryRow(`SELECT to_regclass('public.goose_db_version') IS NOT NULL`).Scan(&exists))
	require.False(t, exists)
}
