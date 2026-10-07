package webui_test

import (
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
	webui.Handler(site).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
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
