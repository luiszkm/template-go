package archtest_test

import (
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// copyApp copies the app module (without build output) into a temp dir.
func copyApp(t *testing.T) string {
	t.Helper()
	src, err := filepath.Abs("..")
	require.NoError(t, err)
	dst := filepath.Join(t.TempDir(), "app")
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
	return dst
}

func goIn(t *testing.T, dir string, args ...string) (string, error) {
	t.Helper()
	cmd := exec.Command("go", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// C27
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

// C28
func TestGenCheck_OpenAPIDriftFails(t *testing.T) {
	app := copyApp(t)
	spec := filepath.Join(app, "openapi.json")
	raw, err := os.ReadFile(spec)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(spec, append(raw, ' '), 0o644))

	out, err := goIn(t, app, "test", "-count=1", "./internal/app", "-run", "^TestOpenAPI_ServedMatchesCommitted$")
	require.Error(t, err, "edited openapi.json must fail the check: %s", out)
}
