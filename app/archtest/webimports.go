package archtest

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

var featureImport = regexp.MustCompile(`["']@/features/([^/"']+)`)

func CheckWebFeatureImports(srcDir string) ([]string, error) {
	featuresDir := filepath.Join(srcDir, "features")
	var violations []string
	err := filepath.WalkDir(featuresDir, func(path string, e fs.DirEntry, err error) error {
		if err != nil || e.IsDir() || !isScript(path) {
			return err
		}
		rel, err := filepath.Rel(featuresDir, path)
		if err != nil {
			return err
		}
		owner, _, _ := strings.Cut(filepath.ToSlash(rel), "/")
		src, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		for _, m := range featureImport.FindAllStringSubmatch(string(src), -1) {
			if m[1] != owner {
				violations = append(violations, fmt.Sprintf("%s: web feature %q imports feature %q", filepath.ToSlash(rel), owner, m[1]))
			}
		}
		return nil
	})
	sort.Strings(violations)
	return violations, err
}

func isScript(path string) bool {
	ext := filepath.Ext(path)
	return ext == ".ts" || ext == ".tsx"
}
