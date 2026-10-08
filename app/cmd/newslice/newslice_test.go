package main

import (
	"crypto/sha256"
	"encoding/hex"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

const module = "github.com/luiszkm/template-go"

func fixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	write := func(rel, content string) {
		p := filepath.Join(root, rel)
		require.NoError(t, os.MkdirAll(filepath.Dir(p), 0o755))
		require.NoError(t, os.WriteFile(p, []byte(content), 0o644))
	}
	copyReal := func(rel string) {
		raw, err := os.ReadFile(filepath.Join("..", "..", rel))
		require.NoError(t, err)
		write(rel, string(raw))
	}
	write("go.mod", "module "+module+"\n\ngo 1.26\n")
	copyReal("internal/features/registry.go")
	copyReal("sqlc.yaml")
	_, err := Generate(root, "users", "list_users")
	require.NoError(t, err)
	return root
}

func read(t *testing.T, root, rel string) string {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(root, rel))
	require.NoError(t, err)
	return string(raw)
}

func treeHash(t *testing.T, root string) map[string]string {
	t.Helper()
	sums := map[string]string{}
	require.NoError(t, filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		raw, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		sum := sha256.Sum256(raw)
		sums[p] = hex.EncodeToString(sum[:])
		return nil
	}))
	return sums
}

func TestGenerate_CreatesSliceAndRegisters(t *testing.T) {
	root := fixture(t)
	entriesBefore := strings.Count(read(t, root, "sqlc.yaml"), "- engine:")
	registryBefore := read(t, root, "internal/features/registry.go")

	_, err := Generate(root, "users", "create_user")
	require.NoError(t, err)

	for _, f := range []string{"endpoint.go", "queries.sql", "create_user_test.go"} {
		require.FileExists(t, filepath.Join(root, "internal/features/users/create_user", f))
	}
	register := read(t, root, "internal/features/users/register.go")
	require.Contains(t, register, `"`+module+`/internal/features/users/create_user"`)
	require.Contains(t, register, "createuser.Register(api, d),")
	require.Contains(t, register, "listusers.Register(api, d),", "existing slice must stay registered")
	require.Equal(t, entriesBefore+1, strings.Count(read(t, root, "sqlc.yaml"), "- engine:"))
	require.Contains(t, read(t, root, "sqlc.yaml"), "queries: internal/features/users/create_user/queries.sql")
	require.Equal(t, registryBefore, read(t, root, "internal/features/registry.go"), "existing feature: registry untouched")
}

func TestGenerate_NewFeatureRegistersFeature(t *testing.T) {
	root := fixture(t)
	_, err := Generate(root, "billing", "get_invoice")
	require.NoError(t, err)

	register := read(t, root, "internal/features/billing/register.go")
	require.Contains(t, register, "package billing")
	require.Contains(t, register, "getinvoice.Register(api, d),")
	registry := read(t, root, "internal/features/registry.go")
	require.Contains(t, registry, `"`+module+`/internal/features/billing"`)
	require.Contains(t, registry, "billing.Register(api, d),")
	require.Contains(t, registry, "users.Register(api, d),")
}

func TestGenerate_RefusesExistingSlice(t *testing.T) {
	root := fixture(t)
	before := treeHash(t, root)
	_, err := Generate(root, "users", "list_users")
	require.ErrorContains(t, err, "already exists")
	require.Equal(t, before, treeHash(t, root))
}

func TestGenerate_RejectsInvalidNames(t *testing.T) {
	cases := []struct{ feature, name string }{
		{"Users", "list"},
		{"1x", "list"},
		{"users", "a-b"},
		{"users", ""},
	}
	for _, c := range cases {
		t.Run(c.feature+"/"+c.name, func(t *testing.T) {
			root := fixture(t)
			before := treeHash(t, root)
			_, err := Generate(root, c.feature, c.name)
			require.Error(t, err)
			require.Equal(t, before, treeHash(t, root))
		})
	}
}

func TestGenerate_MissingMarkerWritesNothing(t *testing.T) {
	root := fixture(t)
	p := filepath.Join(root, "internal/features/users/register.go")
	require.NoError(t, os.WriteFile(p, []byte(strings.ReplaceAll(read(t, root, "internal/features/users/register.go"), "// slices:register", "")), 0o644))
	before := treeHash(t, root)
	_, err := Generate(root, "users", "create_user")
	require.ErrorContains(t, err, "slices:register")
	require.Equal(t, before, treeHash(t, root))
}

func TestMain_FailuresExit1(t *testing.T) {
	bin := filepath.Join(t.TempDir(), "newslice")
	if runtime.GOOS == "windows" {
		bin += ".exe"
	}
	out, err := exec.CommandContext(t.Context(), "go", "build", "-o", bin, ".").CombinedOutput()
	require.NoError(t, err, string(out))

	cases := map[string][]string{
		"invalid name":   {"--feature", "Users", "--name", "list"},
		"existing slice": {"--feature", "users", "--name", "list_users"},
	}
	for name, args := range cases {
		t.Run(name, func(t *testing.T) {
			root := fixture(t)
			out, err := exec.CommandContext(t.Context(), bin, append([]string{"--root", root}, args...)...).CombinedOutput()
			var exitErr *exec.ExitError
			require.ErrorAs(t, err, &exitErr, string(out))
			require.Equal(t, 1, exitErr.ExitCode(), string(out))
		})
	}
}
