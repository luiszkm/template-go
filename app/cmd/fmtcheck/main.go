// Command fmtcheck lists Go files that are not gofmt-formatted and exits 1 if there are any.
// It is `gofmt -l` without needing gofmt or a POSIX shell on PATH.
package main

import (
	"bytes"
	"fmt"
	"go/format"
	"io/fs"
	"os"
	"path/filepath"
)

func main() {
	var bad []string
	err := filepath.WalkDir(".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() && (d.Name() == "testdata" || d.Name() == "node_modules" || d.Name() == ".git") {
			return filepath.SkipDir
		}
		if d.IsDir() || filepath.Ext(path) != ".go" {
			return nil
		}
		src, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		out, err := format.Source(src)
		if err != nil || !bytes.Equal(src, out) {
			bad = append(bad, path)
		}
		return nil
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, "fmtcheck:", err)
		os.Exit(1)
	}
	for _, f := range bad {
		fmt.Println(f)
	}
	if len(bad) > 0 {
		fmt.Fprintf(os.Stderr, "fmtcheck: %d file(s) need gofmt -w\n", len(bad))
		os.Exit(1)
	}
}
