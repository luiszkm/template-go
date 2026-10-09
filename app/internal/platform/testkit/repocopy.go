package testkit

import (
	"context"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

var repoCopySkipDirs = map[string]bool{
	".git": true, "node_modules": true, "bin": true, "dist": true, ".task": true,
	"test-results": true, "playwright-report": true,
}

func CopyRepo(src, dst string) error {
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
				if repoCopySkipDirs[d.Name()] && !strings.HasSuffix(filepath.ToSlash(rel), "webui/dist") {
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

func RunIn(dir string, env []string, name string, args ...string) (string, error) {
	cmd := exec.CommandContext(context.Background(), name, args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), env...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}
