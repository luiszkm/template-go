package archtest

import (
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"

	"golang.org/x/tools/go/packages"
)

func CheckImports(dir string) ([]string, error) {
	pkgs, err := packages.Load(&packages.Config{
		Mode:  packages.NeedName | packages.NeedImports | packages.NeedModule,
		Dir:   dir,
		Tests: true,
	}, "./...")
	if err != nil {
		return nil, err
	}
	seen := map[string]bool{}
	var violations []string
	for _, p := range pkgs {
		if len(p.Errors) > 0 {
			return nil, fmt.Errorf("archtest: load %s: %w", p.PkgPath, p.Errors[0])
		}
		if p.Module == nil {
			continue
		}
		mod := p.Module.Path
		from := strings.TrimSuffix(strings.TrimSuffix(p.PkgPath, ".test"), "_test")
		for imp := range p.Imports {
			if v := rule(mod, from, imp); v != "" && !seen[v] {
				seen[v] = true
				violations = append(violations, v)
			}
		}
	}
	sort.Strings(violations)
	return violations, nil
}

func rule(mod, from, imp string) string {
	features := mod + "/internal/features/"
	platform := mod + "/internal/platform"
	if !strings.HasPrefix(imp, features) {
		return ""
	}
	if from == platform || strings.HasPrefix(from, platform+"/") {
		return fmt.Sprintf("platform must not import features: %s imports %s", from, imp)
	}
	if strings.HasPrefix(from, features) {
		if featureOf(features, from) != featureOf(features, imp) {
			return fmt.Sprintf("feature must not import another feature: %s imports %s", from, imp)
		}
	}
	return ""
}

func featureOf(prefix, pkg string) string {
	name, _, _ := strings.Cut(strings.TrimPrefix(pkg, prefix), "/")
	return name
}

var migrationName = regexp.MustCompile(`^\d{14}_[a-z0-9_]+\.sql$`)

func CheckMigrationNames(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var bad []string
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".sql") {
			continue
		}
		if !migrationName.MatchString(e.Name()) {
			bad = append(bad, fmt.Sprintf("migration %s does not match YYYYMMDDHHMMSS_<name>.sql", e.Name()))
		}
	}
	return bad, nil
}
