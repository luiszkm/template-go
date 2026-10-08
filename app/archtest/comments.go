package archtest

import (
	"fmt"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var skippedDirs = map[string]bool{
	".git": true, "node_modules": true, "dist": true, "bin": true, ".task": true, "testdata": true,
	"test-results": true, "playwright-report": true, ".claude": true, ".cursor": true, ".specs": true,
}

var (
	goAnnotation  = regexp.MustCompile(`^//(go:|nolint|\s*(slices|features):(imports|register)$)`)
	goGenerated   = regexp.MustCompile(`(?m)^// Code generated .* DO NOT EDIT\.$`)
	sqlAnnotation = regexp.MustCompile(`^-- (\+goose |name: )`)
)

func CheckComments(root string) ([]string, error) {
	var violations []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if path != root && skippedDirs[d.Name()] {
				return filepath.SkipDir
			}
			return nil
		}
		check := commentChecker(d.Name())
		if check == nil {
			return nil
		}
		src, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(root, path)
		lines, err := check(path, string(src))
		if err != nil {
			return fmt.Errorf("archtest: %s: %w", rel, err)
		}
		for _, line := range lines {
			violations = append(violations, fmt.Sprintf("%s:%d: comment in code (AGENTS.md rule 8)", filepath.ToSlash(rel), line))
		}
		return nil
	})
	return violations, err
}

type checker func(path, src string) ([]int, error)

func commentChecker(name string) checker {
	switch {
	case strings.HasSuffix(name, ".go"):
		return goComments
	case strings.HasSuffix(name, ".d.ts"), strings.HasSuffix(name, ".gen.ts"):
		return nil
	case strings.HasSuffix(name, ".ts"), strings.HasSuffix(name, ".tsx"), strings.HasSuffix(name, ".js"),
		strings.HasSuffix(name, ".mjs"), strings.HasSuffix(name, ".css"):
		return scriptComments
	case strings.HasSuffix(name, ".sql"):
		return sqlComments
	case strings.HasSuffix(name, ".yml"), strings.HasSuffix(name, ".yaml"), name == "Dockerfile":
		return hashComments
	}
	return nil
}

func goComments(path, src string) ([]int, error) {
	if goGenerated.MatchString(src) {
		return nil, nil
	}
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, src, parser.ParseComments)
	if err != nil {
		return nil, err
	}
	var lines []int
	for _, group := range file.Comments {
		for _, c := range group.List {
			if !goAnnotation.MatchString(c.Text) {
				lines = append(lines, fset.Position(c.Slash).Line)
			}
		}
	}
	return lines, nil
}

func scriptComments(_, src string) ([]int, error) {
	var lines []int
	for i, line := range strings.Split(src, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "// biome-ignore") || strings.HasPrefix(trimmed, "{/* biome-ignore") {
			continue
		}
		if strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "/*") || strings.HasPrefix(trimmed, "{/*") ||
			hasTrailingComment(line) {
			lines = append(lines, i+1)
		}
	}
	return lines, nil
}

func hasTrailingComment(line string) bool {
	for idx := strings.Index(line, " // "); idx != -1; {
		before := line[:idx]
		if strings.TrimSpace(before) != "" && strings.Count(before, `"`)%2 == 0 &&
			strings.Count(before, "'")%2 == 0 && strings.Count(before, "`")%2 == 0 {
			return true
		}
		next := strings.Index(line[idx+1:], " // ")
		if next == -1 {
			return false
		}
		idx += next + 1
	}
	return false
}

func sqlComments(_, src string) ([]int, error) {
	var lines []int
	for i, line := range strings.Split(src, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "--") && !sqlAnnotation.MatchString(trimmed) {
			lines = append(lines, i+1)
		}
	}
	return lines, nil
}

func hashComments(_, src string) ([]int, error) {
	var lines []int
	for i, line := range strings.Split(src, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "#") {
			lines = append(lines, i+1)
		}
	}
	return lines, nil
}
