package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	oldModule = "example.com/old/app"
	oldName   = "Old Name"
)

var fixtureFiles = map[string]string{
	"app/go.mod":                         "module " + oldModule + "\n\ngo 1.26\n",
	"app/tools/go.mod":                   "module " + oldModule + "/tools\n\ngo 1.26\n",
	"app/internal/x/x.go":                "package x\n\nimport _ \"" + oldModule + "/internal/y\"\n",
	"app/.golangci.yml":                  "      local-prefixes:\n        - " + oldModule + "\n",
	"app/internal/platform/httpx/api.go": "package httpx\n\nconst (\n\tAPITitle   = \"" + oldName + " API\"\n\tAPIVersion = \"1.0.0\"\n)\n",
	"app/internal/render/render.go":      "package render\n\nimport \"text/template\"\n\nvar Template = template.New(\"Template\")\n",
	"web/index.html":                     "<html>\n  <head>\n    <title>" + oldName + "</title>\n  </head>\n</html>\n",
	"web/src/routes/_authed/index.tsx":   "<h1 className=\"text-2xl font-semibold\">" + oldName + "</h1>\n",
	"web/package.json":                   "{\n  \"name\": \"app-web\",\n  \"private\": true\n}\n",
	"web/package-lock.json":              "{\n  \"name\": \"app-web\",\n  \"lockfileVersion\": 3,\n  \"packages\": {\n    \"\": {\n      \"name\": \"app-web\"\n    }\n  }\n}\n",
	"web/notes.txt":                      "module " + oldModule + " lives here\n",
	"other.txt":                          "example.com/old/appx stays, and so does the Template word\n",
}

func fixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	for rel, content := range fixtureFiles {
		write(t, root, rel, content)
	}
	return root
}

func write(t *testing.T, root, rel, content string) {
	t.Helper()
	p := filepath.Join(root, filepath.FromSlash(rel))
	require.NoError(t, os.MkdirAll(filepath.Dir(p), 0o755))
	require.NoError(t, os.WriteFile(p, []byte(content), 0o644))
}

func read(t *testing.T, root, rel string) string {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
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

func runRename(t *testing.T, args ...string) (code int, stdout, stderr string) {
	t.Helper()
	var out, errb bytes.Buffer
	code = run(args, &out, &errb)
	return code, out.String(), errb.String()
}

func TestRename_ReplacesModuleEverywhere(t *testing.T) {
	root := fixture(t)
	code, out, errb := runRename(t, "--root", root, "--module", "github.com/acme/foo", "--name", oldName)
	require.Equal(t, 0, code, errb)

	require.Equal(t, "module github.com/acme/foo\n\ngo 1.26\n", read(t, root, "app/go.mod"))
	require.Equal(t, "package x\n\nimport _ \"github.com/acme/foo/internal/y\"\n", read(t, root, "app/internal/x/x.go"))
	require.Equal(t, "      local-prefixes:\n        - github.com/acme/foo\n", read(t, root, "app/.golangci.yml"))
	require.Equal(t, "module github.com/acme/foo lives here\n", read(t, root, "web/notes.txt"))
	require.Equal(t, fixtureFiles["other.txt"], read(t, root, "other.txt"))

	lines := strings.Split(strings.TrimSpace(out), "\n")
	slices.Sort(lines)
	require.Equal(t, []string{
		"wrote app/.golangci.yml",
		"wrote app/go.mod",
		"wrote app/internal/x/x.go",
		"wrote app/tools/go.mod",
		"wrote web/notes.txt",
		"wrote web/package-lock.json",
		"wrote web/package.json",
	}, lines)
}

func TestRename_ToolsModule(t *testing.T) {
	root := fixture(t)
	_, err := Rename(root, "github.com/acme/foo", oldName)
	require.NoError(t, err)
	require.Equal(t, "module github.com/acme/foo/tools\n\ngo 1.26\n", read(t, root, "app/tools/go.mod"))
}

func TestRename_DisplayName(t *testing.T) {
	root := fixture(t)
	_, err := Rename(root, oldModule, "Foo Bar")
	require.NoError(t, err)
	require.Contains(t, read(t, root, "app/internal/platform/httpx/api.go"), "\tAPITitle   = \"Foo Bar API\"\n")
	require.Contains(t, read(t, root, "web/index.html"), "<title>Foo Bar</title>")
	require.Equal(t, "<h1 className=\"text-2xl font-semibold\">Foo Bar</h1>\n", read(t, root, "web/src/routes/_authed/index.tsx"))
}

func TestRename_WebPackageName(t *testing.T) {
	for _, c := range []struct{ module, pkg string }{
		{"github.com/acme/AIGateway", "aigateway-web"},
		{"github.com/acme/foo/v2", "foo-web"},
	} {
		t.Run(c.pkg, func(t *testing.T) {
			root := fixture(t)
			_, err := Rename(root, c.module, oldName)
			require.NoError(t, err)
			require.Equal(t, "{\n  \"name\": \""+c.pkg+"\",\n  \"private\": true\n}\n", read(t, root, "web/package.json"))
			lock := read(t, root, "web/package-lock.json")
			require.Equal(t, 2, strings.Count(lock, "\"name\": \""+c.pkg+"\""))
			require.NotContains(t, lock, "app-web")
		})
	}
}

func TestRename_LeavesExcludedFiles(t *testing.T) {
	excluded := []string{".specs", ".git", "node_modules", "bin", "dist", ".task", "test-results", "playwright-report"}
	for _, dir := range excluded {
		t.Run(dir, func(t *testing.T) {
			root := fixture(t)
			rel := "web/" + dir + "/keep.txt"
			if dir == ".specs" || dir == ".git" {
				rel = dir + "/keep.txt"
			}
			write(t, root, rel, "module "+oldModule+"\n")
			_, err := Rename(root, "github.com/acme/foo", "Foo Bar")
			require.NoError(t, err)
			require.Equal(t, "module "+oldModule+"\n", read(t, root, rel))
		})
	}

	t.Run("binary and other words", func(t *testing.T) {
		root := fixture(t)
		blob := "\x00\x01" + oldModule + "\x00"
		write(t, root, "app/blob.bin", blob)
		_, err := Rename(root, "github.com/acme/foo", "Foo Bar")
		require.NoError(t, err)
		require.Equal(t, blob, read(t, root, "app/blob.bin"))
		require.Equal(t, fixtureFiles["app/internal/render/render.go"], read(t, root, "app/internal/render/render.go"))
		require.Equal(t, fixtureFiles["other.txt"], read(t, root, "other.txt"))
	})
}

func TestRename_RejectsInvalidInput(t *testing.T) {
	for _, c := range []struct{ label, module, name, field string }{
		{"module with spaces", "Not A Path", "Foo", "module"},
		{"module with dot element", "github.com/acme/.foo", "Foo", "module"},
		{"empty name", "github.com/acme/foo", "", "name"},
		{"name of 61", "github.com/acme/foo", strings.Repeat("a", 61), "name"},
		{"name with quote", "github.com/acme/foo", `Fo"o`, "name"},
		{"name with lt", "github.com/acme/foo", "Fo<o", "name"},
		{"name with gt", "github.com/acme/foo", "Fo>o", "name"},
		{"name with backslash", "github.com/acme/foo", `Fo\o`, "name"},
		{"name with open brace", "github.com/acme/foo", "Fo{o", "name"},
		{"name with close brace", "github.com/acme/foo", "Fo}o", "name"},
		{"name with newline", "github.com/acme/foo", "Fo\no", "name"},
	} {
		t.Run(c.label, func(t *testing.T) {
			root := fixture(t)
			before := treeHash(t, root)
			_, err := Rename(root, c.module, c.name)
			require.ErrorContains(t, err, c.field)
			require.Equal(t, before, treeHash(t, root))
		})
	}

	for _, name := range []string{"F", strings.Repeat("a", 60)} {
		t.Run(fmt.Sprintf("accepts name of %d", len(name)), func(t *testing.T) {
			root := fixture(t)
			_, err := Rename(root, "github.com/acme/foo", name)
			require.NoError(t, err)
			require.Contains(t, read(t, root, "web/index.html"), "<title>"+name+"</title>")
		})
	}
}

func TestMain_FailuresExit1(t *testing.T) {
	for _, c := range []struct {
		label string
		args  []string
		want  string
	}{
		{"missing module", []string{"--name", "Foo"}, "usage:"},
		{"missing name", []string{"--module", "github.com/acme/foo"}, "usage:"},
		{"invalid input", []string{"--module", "Not A Path", "--name", "Foo"}, "module"},
	} {
		t.Run(c.label, func(t *testing.T) {
			root := fixture(t)
			before := treeHash(t, root)
			code, out, errb := runRename(t, append([]string{"--root", root}, c.args...)...)
			require.Equal(t, 1, code)
			require.Contains(t, errb, c.want)
			require.Empty(t, out)
			require.Equal(t, before, treeHash(t, root))
		})
	}
}

func TestRename_SameValuesChangesNothing(t *testing.T) {
	root := fixture(t)
	before := treeHash(t, root)
	code, out, errb := runRename(t, "--root", root, "--module", oldModule, "--name", oldName)
	require.Equal(t, 0, code, errb)
	require.NotContains(t, out, "wrote")
	require.Equal(t, before, treeHash(t, root))
}

func TestRename_ModuleAtEndOfFile(t *testing.T) {
	root := fixture(t)
	write(t, root, "web/tail.txt", "see "+oldModule)
	_, err := Rename(root, "github.com/acme/foo", oldName)
	require.NoError(t, err)
	require.Equal(t, "see github.com/acme/foo", read(t, root, "web/tail.txt"))
}

func TestRename_NameRunesAndCarriageReturn(t *testing.T) {
	t.Run("rejects carriage return", func(t *testing.T) {
		root := fixture(t)
		before := treeHash(t, root)
		_, err := Rename(root, "github.com/acme/foo", "Fo\ro")
		require.ErrorContains(t, err, "name")
		require.Equal(t, before, treeHash(t, root))
	})
	t.Run("accepts 60 multibyte characters", func(t *testing.T) {
		root := fixture(t)
		name := strings.Repeat("ç", 60)
		_, err := Rename(root, "github.com/acme/foo", name)
		require.NoError(t, err)
		require.Contains(t, read(t, root, "web/index.html"), "<title>"+name+"</title>")
	})
	t.Run("rejects 61 multibyte characters", func(t *testing.T) {
		root := fixture(t)
		_, err := Rename(root, "github.com/acme/foo", strings.Repeat("ç", 61))
		require.ErrorContains(t, err, "name")
	})
}

func TestRename_ErrorPaths(t *testing.T) {
	for _, c := range []struct {
		label   string
		rel     string
		content string
		want    string
	}{
		{"go.mod without module line", "app/go.mod", "go 1.26\n", "app/go.mod"},
		{"index.html without title", "web/index.html", "<html></html>\n", "web/index.html"},
		{"invalid package.json", "web/package.json", "{not json", "web/package.json"},
	} {
		t.Run(c.label, func(t *testing.T) {
			root := fixture(t)
			write(t, root, c.rel, c.content)
			before := treeHash(t, root)
			code, out, errb := runRename(t, "--root", root, "--module", "github.com/acme/foo", "--name", "Foo")
			require.Equal(t, 1, code)
			require.Contains(t, errb, c.want)
			require.Empty(t, out)
			require.Equal(t, before, treeHash(t, root))
		})
	}

	t.Run("missing go.mod", func(t *testing.T) {
		root := fixture(t)
		require.NoError(t, os.Remove(filepath.Join(root, "app", "go.mod")))
		before := treeHash(t, root)
		code, out, errb := runRename(t, "--root", root, "--module", "github.com/acme/foo", "--name", "Foo")
		require.Equal(t, 1, code)
		require.Contains(t, errb, "go.mod")
		require.Empty(t, out)
		require.Equal(t, before, treeHash(t, root))
	})

	t.Run("write fails after partial writes", func(t *testing.T) {
		root := fixture(t)
		locked := filepath.Join(root, "web", "package.json")
		require.NoError(t, os.Chmod(locked, 0o444))
		t.Cleanup(func() { _ = os.Chmod(locked, 0o644) })
		code, out, errb := runRename(t, "--root", root, "--module", "github.com/acme/foo", "--name", oldName)
		require.Equal(t, 1, code)
		require.Contains(t, out, "wrote app/go.mod\n")
		require.NotContains(t, out, "wrote web/package.json")
		require.Contains(t, errb, "rename:")
		require.Contains(t, errb, "package.json")
		require.Equal(t, "module github.com/acme/foo\n\ngo 1.26\n", read(t, root, "app/go.mod"))
	})
}
