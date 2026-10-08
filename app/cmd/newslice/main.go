package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
)

var validName = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

func main() {
	feature := flag.String("feature", "", "feature folder, snake_case (e.g. users)")
	name := flag.String("name", "", "slice folder, snake_case (e.g. create_user)")
	root := flag.String("root", ".", "Go module root (the app directory)")
	flag.Parse()

	files, err := Generate(*root, *feature, *name)
	if err != nil {
		fmt.Fprintln(os.Stderr, "newslice:", err)
		os.Exit(1)
	}
	for _, f := range files {
		fmt.Println("wrote", f)
	}
}

func Generate(root, feature, name string) ([]string, error) {
	if !validName.MatchString(feature) {
		return nil, fmt.Errorf("invalid FEATURE %q: must match %s", feature, validName)
	}
	if !validName.MatchString(name) {
		return nil, fmt.Errorf("invalid NAME %q: must match %s", name, validName)
	}
	module, err := modulePath(filepath.Join(root, "go.mod"))
	if err != nil {
		return nil, err
	}

	featureDir := filepath.Join(root, "internal", "features", feature)
	sliceDir := filepath.Join(featureDir, name)
	if _, err := os.Stat(sliceDir); err == nil {
		return nil, fmt.Errorf("slice %s already exists", filepath.ToSlash(sliceDir))
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}

	d := data{
		Module:      module,
		Feature:     feature,
		FeaturePkg:  pkgName(feature),
		Name:        name,
		Pkg:         pkgName(name),
		Path:        "/api/v1/" + kebab(feature) + "/" + kebab(name),
		OperationID: kebab(feature) + "-" + kebab(name),
	}

	writes := map[string][]byte{}
	for file, tmpl := range map[string]*template.Template{
		"endpoint.go":     endpointTmpl,
		"queries.sql":     queriesTmpl,
		name + "_test.go": testTmpl,
	} {
		out, err := render(tmpl, d, strings.HasSuffix(file, ".go"))
		if err != nil {
			return nil, err
		}
		writes[filepath.Join(sliceDir, file)] = out
	}

	registerPath := filepath.Join(featureDir, "register.go")
	register, err := os.ReadFile(registerPath)
	newFeature := errors.Is(err, os.ErrNotExist)
	switch {
	case newFeature:
		if register, err = render(featureRegisterTmpl, d, true); err != nil {
			return nil, err
		}
	case err != nil:
		return nil, err
	}
	if register, err = insert(register, registerPath, "slices",
		importSpec(d.Pkg, name, module+"/internal/features/"+feature+"/"+name),
		d.Pkg+".Register(api, d),"); err != nil {
		return nil, err
	}
	writes[registerPath] = register

	if newFeature {
		registryPath := filepath.Join(root, "internal", "features", "registry.go")
		registry, err := os.ReadFile(registryPath)
		if err != nil {
			return nil, err
		}
		if registry, err = insert(registry, registryPath, "features",
			importSpec(d.FeaturePkg, feature, module+"/internal/features/"+feature),
			d.FeaturePkg+".Register(api, d),"); err != nil {
			return nil, err
		}
		writes[registryPath] = registry
	}

	sqlcPath := filepath.Join(root, "sqlc.yaml")
	sqlc, err := os.ReadFile(sqlcPath)
	if err != nil {
		return nil, err
	}
	entry, err := render(sqlcEntryTmpl, d, false)
	if err != nil {
		return nil, err
	}
	writes[sqlcPath] = appendSQLCEntry(sqlc, entry)

	if err := os.MkdirAll(sliceDir, 0o755); err != nil {
		return nil, err
	}
	var written []string
	for path, content := range writes {
		if err := os.WriteFile(path, content, 0o644); err != nil {
			return written, err
		}
		written = append(written, filepath.ToSlash(path))
	}
	return written, nil
}

type data struct {
	Module, Feature, FeaturePkg, Name, Pkg, Path, OperationID string
}

func render(t *template.Template, d data, goSource bool) ([]byte, error) {
	var buf bytes.Buffer
	if err := t.Execute(&buf, d); err != nil {
		return nil, err
	}
	if !goSource {
		return buf.Bytes(), nil
	}
	return format.Source(buf.Bytes())
}

func insert(src []byte, path, kind, importLine, callLine string) ([]byte, error) {
	s := string(src)
	for _, m := range []struct{ marker, line string }{
		{"// " + kind + ":imports", importLine},
		{"// " + kind + ":register", callLine},
	} {
		i := strings.Index(s, m.marker)
		if i < 0 {
			return nil, fmt.Errorf("%s: marker %q not found; restore it to use the generator", filepath.ToSlash(path), m.marker)
		}
		lineStart := strings.LastIndex(s[:i], "\n") + 1
		s = s[:lineStart] + m.line + "\n" + s[lineStart:]
	}
	return format.Source([]byte(s))
}

func appendSQLCEntry(cfg, entry []byte) []byte {
	s := strings.TrimRight(string(cfg), "\n")
	s = strings.TrimSuffix(s, " []")
	return []byte(s + "\n" + string(entry))
}

func modulePath(goMod string) (string, error) {
	raw, err := os.ReadFile(goMod)
	if err != nil {
		return "", err
	}
	for line := range strings.SplitSeq(string(raw), "\n") {
		if rest, ok := strings.CutPrefix(strings.TrimSpace(line), "module "); ok {
			return strings.TrimSpace(rest), nil
		}
	}
	return "", fmt.Errorf("%s: no module line", goMod)
}

func importSpec(pkg, folder, path string) string {
	if pkg == folder {
		return fmt.Sprintf("%q", path)
	}
	return fmt.Sprintf("%s %q", pkg, path)
}

func pkgName(s string) string { return strings.ReplaceAll(s, "_", "") }
func kebab(s string) string   { return strings.ReplaceAll(s, "_", "-") }

var endpointTmpl = template.Must(template.New("endpoint").Parse(`package {{.Pkg}}

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	"{{.Module}}/internal/platform/deps"
	"{{.Module}}/internal/platform/op"
)

const permission op.Permission = "{{.Feature}}:{{.Name}}"

type Input struct{}

type Output struct {
	Body struct {
		Message string ` + "`json:\"message\"`" + `
	}
}

func Register(api huma.API, _ deps.Deps) error {
	return op.Register(api, op.Spec{
		ID:         "{{.OperationID}}",
		Method:     http.MethodGet,
		Path:       "{{.Path}}",
		Summary:    "{{.Feature}}: {{.Name}} (not implemented yet)",
		Tags:       []string{"{{.Feature}}"},
		Permission: permission,
	}, func(_ context.Context, _ *Input) (*Output, error) {
		return nil, huma.Error501NotImplemented("{{.Feature}}/{{.Name}} is not implemented yet")
	})
}
`))

var queriesTmpl = template.Must(template.New("queries").Parse(`-- name: Placeholder :one
SELECT 1::int AS one;
`))

var testTmpl = template.Must(template.New("test").Parse(`package {{.Pkg}}_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	{{.Pkg}} "{{.Module}}/internal/features/{{.Feature}}/{{.Name}}"
	"{{.Module}}/internal/platform/testkit"
)

func TestEndpoint_NotImplemented(t *testing.T) {
	pool := testkit.MigratedDB(t)
	api, h, d := testkit.NewAPI(t, pool, testkit.APIOptions{})
	require.NoError(t, {{.Pkg}}.Register(api, d))

	rec := testkit.Do(t, h, testkit.Request{Method: http.MethodGet, Path: "{{.Path}}"})
	require.Equal(t, http.StatusUnauthorized, rec.Code)

	user := testkit.SignIn(t, pool, "{{.Feature}}:{{.Name}}")
	rec = testkit.Do(t, h, testkit.Request{Method: http.MethodGet, Path: "{{.Path}}", Cookie: user.Cookie})
	require.Equal(t, http.StatusNotImplemented, rec.Code)
	require.Equal(t, "application/problem+json", rec.Header().Get("Content-Type"))
}
`))

var featureRegisterTmpl = template.Must(template.New("feature").Parse(`package {{.FeaturePkg}}

import (
	"errors"

	"github.com/danielgtaylor/huma/v2"

	"{{.Module}}/internal/platform/deps"
	// slices:imports
)

func Register(api huma.API, d deps.Deps) error {
	return errors.Join(
	// slices:register
	)
}
`))

var sqlcEntryTmpl = template.Must(template.New("sqlc").Parse(`  - engine: postgresql
    schema: migrations
    queries: internal/features/{{.Feature}}/{{.Name}}/queries.sql
    gen:
      go:
        package: db
        out: internal/features/{{.Feature}}/{{.Name}}/db
        sql_package: pgx/v5
`))
