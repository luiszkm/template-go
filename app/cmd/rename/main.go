package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"golang.org/x/mod/module"
)

const (
	usage         = "usage: rename --module <module path> --name <display name> [--root <repository root>]"
	maxNameLength = 60
	apiTitleFile  = "app/internal/platform/httpx/api.go"
	indexHTMLFile = "web/index.html"
	homeFile      = "web/src/routes/_authed/index.tsx"
	packageFile   = "web/package.json"
	lockFile      = "web/package-lock.json"
)

var (
	skippedDirs = map[string]bool{
		".specs": true, ".git": true, "node_modules": true, "bin": true, "dist": true, ".task": true,
		"test-results": true, "playwright-report": true,
	}
	forbiddenNameChars = "\"<>\\{}\r\n"
	titlePattern       = regexp.MustCompile(`<title>([^<]*)</title>`)
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("rename", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	root := flags.String("root", ".", "repository root")
	modulePath := flags.String("module", "", "new Go module path")
	name := flags.String("name", "", "new display name")
	if err := flags.Parse(args); err != nil || *modulePath == "" || !flagSet(flags, "name") {
		fmt.Fprintln(stderr, usage)
		return 1
	}
	written, err := Rename(*root, *modulePath, *name)
	for _, f := range written {
		fmt.Fprintln(stdout, "wrote", f)
	}
	if err != nil {
		fmt.Fprintln(stderr, "rename:", err)
		return 1
	}
	return 0
}

func flagSet(flags *flag.FlagSet, name string) bool {
	found := false
	flags.Visit(func(f *flag.Flag) { found = found || f.Name == name })
	return found
}

func Rename(root, newModule, newName string) ([]string, error) {
	if err := module.CheckPath(newModule); err != nil {
		return nil, fmt.Errorf("invalid module %q: %w", newModule, err)
	}
	if err := checkName(newName); err != nil {
		return nil, err
	}
	current, err := readCurrent(root)
	if err != nil {
		return nil, err
	}
	edits, err := planEdits(root, current, names{module: newModule, display: newName, pkg: webPackage(newModule)})
	if err != nil {
		return nil, err
	}
	return applyEdits(root, edits)
}

func checkName(name string) error {
	if name == "" || len([]rune(name)) > maxNameLength {
		return fmt.Errorf("invalid name %q: must have 1 to %d characters", name, maxNameLength)
	}
	if strings.ContainsAny(name, forbiddenNameChars) {
		return fmt.Errorf("invalid name %q: must not contain any of %q", name, forbiddenNameChars)
	}
	return nil
}

func webPackage(modulePath string) string {
	prefix, _, _ := module.SplitPathVersion(modulePath)
	return strings.ToLower(path.Base(prefix)) + "-web"
}

type names struct {
	module  string
	display string
	pkg     string
}

func readCurrent(root string) (names, error) {
	goMod, err := readFile(root, "app/go.mod")
	if err != nil {
		return names{}, err
	}
	modulePath, ok := moduleLine(goMod)
	if !ok {
		return names{}, errors.New("app/go.mod: no module line")
	}
	html, err := readFile(root, indexHTMLFile)
	if err != nil {
		return names{}, err
	}
	title := titlePattern.FindStringSubmatch(html)
	if title == nil {
		return names{}, fmt.Errorf("%s: no <title>", indexHTMLFile)
	}
	raw, err := readFile(root, packageFile)
	if err != nil {
		return names{}, err
	}
	var pkg struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal([]byte(raw), &pkg); err != nil {
		return names{}, fmt.Errorf("%s: %w", packageFile, err)
	}
	return names{module: modulePath, display: title[1], pkg: pkg.Name}, nil
}

func moduleLine(goMod string) (string, bool) {
	for line := range strings.SplitSeq(goMod, "\n") {
		if rest, ok := strings.CutPrefix(strings.TrimSpace(line), "module "); ok {
			return strings.TrimSpace(rest), true
		}
	}
	return "", false
}

func readFile(root, rel string) (string, error) {
	raw, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
	return string(raw), err
}

func planEdits(root string, from, to names) (map[string]string, error) {
	modulePattern := regexp.MustCompile(regexp.QuoteMeta(from.module) + `([^A-Za-z0-9._~-]|$)`)
	nameEdits := map[string][][2]string{
		apiTitleFile:  {{`"` + from.display + ` API"`, `"` + to.display + ` API"`}},
		indexHTMLFile: {{"<title>" + from.display + "</title>", "<title>" + to.display + "</title>"}},
		homeFile:      {{">" + from.display + "</h1>", ">" + to.display + "</h1>"}},
		packageFile:   {{`"name": "` + from.pkg + `"`, `"name": "` + to.pkg + `"`}},
		lockFile:      {{`"name": "` + from.pkg + `"`, `"name": "` + to.pkg + `"`}},
	}
	edits := map[string]string{}
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if p != root && skippedDirs[d.Name()] {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if !d.Type().IsRegular() {
			return nil
		}
		raw, err := os.ReadFile(p)
		if err != nil || bytes.IndexByte(raw, 0) >= 0 {
			return err
		}
		rel, err := filepath.Rel(root, p)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		text := modulePattern.ReplaceAllString(string(raw), to.module+"${1}")
		for _, pair := range nameEdits[rel] {
			text = strings.ReplaceAll(text, pair[0], pair[1])
		}
		if text != string(raw) {
			edits[rel] = text
		}
		return nil
	})
	return edits, err
}

func applyEdits(root string, edits map[string]string) ([]string, error) {
	paths := make([]string, 0, len(edits))
	for rel := range edits {
		paths = append(paths, rel)
	}
	sort.Strings(paths)
	written := make([]string, 0, len(paths))
	for _, rel := range paths {
		p := filepath.Join(root, filepath.FromSlash(rel))
		info, err := os.Stat(p)
		if err != nil {
			return written, err
		}
		if err := os.WriteFile(p, []byte(edits[rel]), info.Mode().Perm()); err != nil {
			return written, err
		}
		written = append(written, rel)
	}
	return written, nil
}
