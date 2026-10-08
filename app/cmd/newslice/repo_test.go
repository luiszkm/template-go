package main

import (
	"context"
	"encoding/json"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

const innerEnv = "NEWSLICE_INNER_CHECK"

var (
	genOnce sync.Once
	genRoot string
	genErr  error
	genOut  string
)

func generatedRepo(t *testing.T) string {
	t.Helper()
	if os.Getenv(innerEnv) == "1" {
		t.Skip("inside the repository copy started by TestGenerated_TaskCheckPasses")
	}
	genOnce.Do(func() { genRoot, genOut, genErr = buildGeneratedRepo() })
	require.NoError(t, genErr, genOut)
	return genRoot
}

func buildGeneratedRepo() (root, out string, err error) {
	src, err := filepath.Abs(filepath.Join("..", "..", ".."))
	if err != nil {
		return "", "", err
	}
	dst, err := os.MkdirTemp("", "newslice-repo-")
	if err != nil {
		return "", "", err
	}
	if dst, err = filepath.EvalSymlinks(dst); err != nil {
		return "", "", err
	}
	if err = copyRepo(src, dst); err != nil {
		return dst, "", err
	}
	out, err = runIn(dst, nil, "task", "new:slice", "FEATURE=demo", "NAME=get_thing")
	return dst, out, err
}

func TestMain(m *testing.M) {
	code := m.Run()
	if genRoot != "" {
		_ = os.RemoveAll(genRoot)
	}
	os.Exit(code)
}

var skipDirs = map[string]bool{
	".git": true, "node_modules": true, "bin": true, "dist": true, ".task": true,
	"test-results": true, "playwright-report": true,
}

func copyRepo(src, dst string) error {
	for _, top := range []string{"app", "web", "Taskfile.yml", "AGENTS.md", ".claude", ".cursor", ".github"} {
		if _, err := os.Stat(filepath.Join(src, top)); os.IsNotExist(err) {
			continue
		}
		err := filepath.WalkDir(filepath.Join(src, top), func(p string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			rel, _ := filepath.Rel(src, p)
			target := filepath.Join(dst, rel)
			if d.IsDir() {
				if skipDirs[d.Name()] && !strings.HasSuffix(filepath.ToSlash(rel), "webui/dist") {
					return filepath.SkipDir
				}
				return os.MkdirAll(target, 0o755)
			}
			raw, err := os.ReadFile(p)
			if err != nil {
				return err
			}
			return os.WriteFile(target, raw, 0o644)
		})
		if err != nil {
			return err
		}
	}
	return linkDir(filepath.Join(src, "web", "node_modules"), filepath.Join(dst, "web", "node_modules"))
}

func linkDir(target, link string) error {
	if err := os.Symlink(target, link); err == nil || runtime.GOOS != "windows" {
		return err
	}
	out, err := exec.CommandContext(context.Background(), "cmd", "/c", "mklink", "/J", link, target).CombinedOutput()
	if err != nil {
		return &exec.ExitError{Stderr: out}
	}
	return nil
}

func runIn(dir string, env []string, name string, args ...string) (string, error) {
	cmd := exec.CommandContext(context.Background(), name, args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), env...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func TestGenerated_TaskCheckPasses(t *testing.T) {
	root := generatedRepo(t)
	out, err := runIn(root, []string{innerEnv + "=1"}, "task", "check")
	require.NoError(t, err, out)
}

func TestGenerated_SliceAnswers501(t *testing.T) {
	root := generatedRepo(t)
	app := filepath.Join(root, "app")

	doc, err := os.ReadFile(filepath.Join(app, "openapi.json"))
	require.NoError(t, err)
	var spec struct {
		Paths map[string]json.RawMessage `json:"paths"`
	}
	require.NoError(t, json.Unmarshal(doc, &spec))
	var demo []string
	for p := range spec.Paths {
		if strings.HasPrefix(p, "/api/v1/demo/") {
			demo = append(demo, p)
		}
	}
	require.Equal(t, []string{"/api/v1/demo/get-thing"}, demo)

	endpoint, err := os.ReadFile(filepath.Join(app, "internal", "features", "demo", "get_thing", "endpoint.go"))
	require.NoError(t, err)
	require.Contains(t, string(endpoint), `const permission op.Permission = "demo:get_thing"`)

	test, err := os.ReadFile(filepath.Join(app, "internal", "features", "demo", "get_thing", "get_thing_test.go"))
	require.NoError(t, err)
	require.Contains(t, string(test), "http.StatusNotImplemented")

	out, err := runIn(app, nil, "go", "test", "-count=1", "-v", "./internal/features/demo/get_thing")
	require.NoError(t, err, out)
	require.Contains(t, out, "--- PASS: TestEndpoint_NotImplemented")
}

func TestGeneratedSlice_RequiresSession(t *testing.T) {
	root := generatedRepo(t)
	app := filepath.Join(root, "app")

	test, err := os.ReadFile(filepath.Join(app, "internal", "features", "demo", "get_thing", "get_thing_test.go"))
	require.NoError(t, err)
	require.Contains(t, string(test), "http.StatusUnauthorized")
	require.Contains(t, string(test), `testkit.SignIn(t, pool, "demo:get_thing")`)
	require.Contains(t, string(test), "http.StatusNotImplemented")

	out, err := runIn(app, nil, "go", "test", "-count=1", "-v", "-run", "^TestEndpoint_NotImplemented$", "./internal/features/demo/get_thing")
	require.NoError(t, err, out)
	require.Contains(t, out, "--- PASS: TestEndpoint_NotImplemented")
}
