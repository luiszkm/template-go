package main

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/luiszkm/template-go/internal/platform/testkit"
)

const (
	innerEnv         = "RENAME_INNER_CHECK"
	newsliceInnerEnv = "NEWSLICE_INNER_CHECK"
	templateModule   = "github.com/luiszkm/template-go"
)

var (
	renamedOnce sync.Once
	renamedRoot string
	renamedErr  error
	renamedOut  string
)

func renamedRepo(t *testing.T) string {
	t.Helper()
	if os.Getenv(innerEnv) == "1" || os.Getenv(newsliceInnerEnv) == "1" {
		t.Skip("inside a repository copy started by another copy test")
	}
	renamedOnce.Do(func() { renamedRoot, renamedOut, renamedErr = buildRenamedRepo() })
	require.NoError(t, renamedErr, renamedOut)
	return renamedRoot
}

func repoRoot() (string, error) {
	return filepath.Abs(filepath.Join("..", "..", ".."))
}

func buildRenamedRepo() (root, out string, err error) {
	src, err := repoRoot()
	if err != nil {
		return "", "", err
	}
	dst, err := os.MkdirTemp("", "rename-repo-")
	if err != nil {
		return "", "", err
	}
	if dst, err = filepath.EvalSymlinks(dst); err != nil {
		return "", "", err
	}
	if err = testkit.CopyRepo(src, dst); err != nil {
		return dst, "", err
	}
	out, err = testkit.RunIn(dst, nil, "task", "rename", "MODULE=github.com/acme/foo", "NAME=Foo Bar")
	return dst, out, err
}

func TestMain(m *testing.M) {
	code := m.Run()
	if renamedRoot != "" {
		_ = os.RemoveAll(renamedRoot)
	}
	os.Exit(code)
}

func TestRenamed_RepoCarriesNewNames(t *testing.T) {
	root := renamedRepo(t)

	var stale []string
	require.NoError(t, filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() && d.Name() == "node_modules" {
			return filepath.SkipDir
		}
		if !d.Type().IsRegular() {
			return nil
		}
		raw, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		if strings.Contains(string(raw), templateModule) {
			stale = append(stale, p)
		}
		return nil
	}))
	require.Empty(t, stale)

	spec, err := os.ReadFile(filepath.Join(root, "app", "openapi.json"))
	require.NoError(t, err)
	require.Contains(t, string(spec), `"title":"Foo Bar API"`)

	pkg, err := os.ReadFile(filepath.Join(root, "web", "package.json"))
	require.NoError(t, err)
	require.Contains(t, string(pkg), `"name": "foo-web"`)
}

func TestRenamed_TaskCheckPasses(t *testing.T) {
	root := renamedRepo(t)
	out, err := testkit.RunIn(root, []string{innerEnv + "=1", newsliceInnerEnv + "=1"}, "task", "check")
	require.NoError(t, err, out)
}

func TestTaskfileAndDocs_DeclareRename(t *testing.T) {
	root, err := repoRoot()
	require.NoError(t, err)

	agents, err := os.ReadFile(filepath.Join(root, "AGENTS.md"))
	require.NoError(t, err)
	require.Regexp(t, "(?m)^\\| `task rename MODULE=<m> NAME=<n>` \\| .+ \\|$", string(agents))

	taskfile, err := os.ReadFile(filepath.Join(root, "Taskfile.yml"))
	require.NoError(t, err)
	block := regexp.MustCompile(`(?ms)^  rename:\n(.*?)\n\n`).FindStringSubmatch(string(taskfile))
	require.NotNil(t, block, "Taskfile.yml must declare a rename task")
	body := block[1]
	require.Regexp(t, `vars: \[MODULE, NAME\]`, body)
	run := strings.Index(body, "go run ./cmd/rename")
	gen := strings.Index(body, "task: gen")
	require.GreaterOrEqual(t, run, 0)
	require.Greater(t, gen, run)
}
