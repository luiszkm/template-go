package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/luiszkm/template-go/internal/platform/testkit"
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
	if err = testkit.CopyRepo(src, dst); err != nil {
		return dst, "", err
	}
	out, err = testkit.RunIn(dst, nil, "task", "new:slice", "FEATURE=demo", "NAME=get_thing")
	return dst, out, err
}

func TestMain(m *testing.M) {
	code := m.Run()
	if genRoot != "" {
		_ = os.RemoveAll(genRoot)
	}
	os.Exit(code)
}

func TestGenerated_TaskCheckPasses(t *testing.T) {
	root := generatedRepo(t)
	out, err := testkit.RunIn(root, []string{innerEnv + "=1"}, "task", "check")
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

	out, err := testkit.RunIn(app, nil, "go", "test", "-count=1", "-v", "./internal/features/demo/get_thing")
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

	out, err := testkit.RunIn(app, nil, "go", "test", "-count=1", "-v", "-run", "^TestEndpoint_NotImplemented$", "./internal/features/demo/get_thing")
	require.NoError(t, err, out)
	require.Contains(t, out, "--- PASS: TestEndpoint_NotImplemented")
}
