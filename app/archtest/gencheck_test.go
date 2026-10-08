package archtest_test

import (
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func copyApp(t *testing.T) string {
	t.Helper()
	dst := filepath.Join(t.TempDir(), "app")
	copyAppTo(t, dst)
	return dst
}

func copyAppTo(t *testing.T, dst string) {
	t.Helper()
	src, err := filepath.Abs("..")
	require.NoError(t, err)
	require.NoError(t, filepath.WalkDir(src, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(src, p)
		if d.IsDir() {
			if rel == "bin" {
				return filepath.SkipDir
			}
			return os.MkdirAll(filepath.Join(dst, rel), 0o755)
		}
		raw, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(dst, rel), raw, 0o644)
	}))
}

func goIn(t *testing.T, dir string, args ...string) (string, error) {
	t.Helper()
	cmd := exec.CommandContext(t.Context(), "go", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func TestGenCheck_SqlcDriftFails(t *testing.T) {
	app := copyApp(t)
	out, err := goIn(t, app, "run", "./cmd/newslice", "--feature", "gencheck_drift", "--name", "get_thing")
	require.NoError(t, err, out)
	out, err = goIn(t, app, "run", "./cmd/sqlcrun", "generate")
	require.NoError(t, err, out)

	out, err = goIn(t, app, "run", "./cmd/sqlcrun", "diff")
	require.NoError(t, err, "freshly generated code must pass: %s", out)

	generated := filepath.Join(app, "internal", "features", "gencheck_drift", "get_thing", "db", "queries.sql.go")
	raw, err := os.ReadFile(generated)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(generated, append(raw, []byte("\n// hand edit\n")...), 0o644))

	out, err = goIn(t, app, "run", "./cmd/sqlcrun", "diff")
	require.Error(t, err, "edited generated code must fail: %s", out)
}

func TestGenCheck_OpenAPIDriftFails(t *testing.T) {
	app := copyApp(t)
	spec := filepath.Join(app, "openapi.json")
	raw, err := os.ReadFile(spec)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(spec, append(raw, ' '), 0o644))

	out, err := goIn(t, app, "test", "-count=1", "./internal/app", "-run", "^TestOpenAPI_ServedMatchesCommitted$")
	require.Error(t, err, "edited openapi.json must fail the check: %s", out)
}

func TestTaskfile_GatesFailOnFailingStep(t *testing.T) {
	root, err := filepath.EvalSymlinks(t.TempDir())
	require.NoError(t, err)
	raw, err := os.ReadFile(filepath.Join("..", "..", "Taskfile.yml"))
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(root, "Taskfile.yml"), raw, 0o644))
	copyAppTo(t, filepath.Join(root, "app"))
	require.NoError(t, os.WriteFile(filepath.Join(root, "app", "unformatted.go"), []byte("package  main\nfunc  x( ) {}\n"), 0o644))

	for _, gate := range []string{"check", "check:fast"} {
		t.Run(gate, func(t *testing.T) {
			cmd := exec.CommandContext(t.Context(), "task", gate)
			cmd.Dir = root
			out, err := cmd.CombinedOutput()
			var exitErr *exec.ExitError
			require.ErrorAs(t, err, &exitErr, "task %s must fail: %s", gate, out)
			require.NotEqual(t, 0, exitErr.ExitCode())
			require.Contains(t, string(out), "unformatted.go", "must fail at fmt:check on the unformatted file")
		})
	}
}
