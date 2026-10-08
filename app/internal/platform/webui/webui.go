// Package webui serves the built SPA embedded in the binary, falling back to index.html.
package webui

import (
	"embed"
	"io/fs"
	"net/http"
	"path"
	"strings"

	"github.com/luiszkm/template-go/internal/platform/httpx"
)

//go:embed all:dist
var dist embed.FS

// Dist returns the embedded build output (copied here by `task build`).
func Dist() fs.FS {
	sub, err := fs.Sub(dist, "dist")
	if err != nil {
		panic(err)
	}
	return sub
}

// Handler serves files from fsys; any path that is not a file gets index.html with 200.
// Only GET and HEAD are served; any other method gets a 405 Problem.
func Handler(fsys fs.FS) http.Handler {
	files := http.FileServerFS(fsys)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			w.Header().Set("Allow", "GET, HEAD")
			httpx.WriteProblem(w, r, http.StatusMethodNotAllowed, "method "+r.Method+" not allowed on "+r.URL.Path)
			return
		}
		name := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
		if name != "" && name != "index.html" {
			if info, err := fs.Stat(fsys, name); err == nil && !info.IsDir() {
				files.ServeHTTP(w, r)
				return
			}
		}
		index, err := fs.ReadFile(fsys, "index.html")
		if err != nil {
			httpx.WriteProblem(w, r, http.StatusNotFound, "web build not found: run `task build`")
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-cache")
		_, _ = w.Write(index)
	})
}
