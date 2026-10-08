package webui_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/require"

	"github.com/luiszkm/template-go/internal/platform/webui"
)

var site = fstest.MapFS{
	"index.html":    {Data: []byte("<html>index</html>")},
	"assets/app.js": {Data: []byte("console.log('app')")},
}

func get(path string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	webui.Handler(site).ServeHTTP(rec, httptest.NewRequestWithContext(context.Background(), http.MethodGet, path, nil))
	return rec
}

// C44
func TestHandler_FallsBackToIndex(t *testing.T) {
	rec := get("/users/42")
	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "<html>index</html>", rec.Body.String())
}

// C45
func TestHandler_ServesStaticFile(t *testing.T) {
	rec := get("/assets/app.js")
	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "console.log('app')", rec.Body.String())
}

func TestHandler_RootServesIndex(t *testing.T) {
	rec := get("/")
	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "<html>index</html>", rec.Body.String())
}

// C56
func TestHandler_MissingBuildIs404Problem(t *testing.T) {
	empty := fstest.MapFS{"assets/app.js": {Data: []byte("console.log('app')")}}
	rec := httptest.NewRecorder()
	webui.Handler(empty).ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/users/42", nil))
	require.Equal(t, http.StatusNotFound, rec.Code)
	require.Equal(t, "application/problem+json", rec.Header().Get("Content-Type"))
}

// C70
func TestHandler_NonGetIs405Problem(t *testing.T) {
	rec := httptest.NewRecorder()
	webui.Handler(site).ServeHTTP(rec, httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/users", nil))
	require.Equal(t, http.StatusMethodNotAllowed, rec.Code)
	require.Equal(t, "application/problem+json", rec.Header().Get("Content-Type"))
	require.Equal(t, "GET, HEAD", rec.Header().Get("Allow"))
}
