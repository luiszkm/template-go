package app_test

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/stretchr/testify/require"

	"github.com/luiszkm/template-go/internal/app"
	"github.com/luiszkm/template-go/internal/platform/deps"
	"github.com/luiszkm/template-go/internal/platform/op"
	"github.com/luiszkm/template-go/internal/platform/testkit"
)

var site = fstest.MapFS{"index.html": {Data: []byte("<html>spa</html>")}}

func newServer(t *testing.T) http.Handler {
	t.Helper()
	h, err := app.New(app.Options{Logger: testkit.DiscardLogger(), Web: site})
	require.NoError(t, err)
	return h
}

func get(h http.Handler, path string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
	return rec
}

// C12
func TestRouting_UnknownAPIPathIs404Problem(t *testing.T) {
	rec := get(newServer(t), "/api/does-not-exist")
	require.Equal(t, http.StatusNotFound, rec.Code)
	require.Equal(t, "application/problem+json", rec.Header().Get("Content-Type"))
}

// C46
func TestRouting_APIAndHealthNotSwallowedBySPA(t *testing.T) {
	h := newServer(t)
	for _, p := range []string{"/api/x", "/healthz", "/readyz"} {
		rec := get(h, p)
		require.NotContains(t, rec.Body.String(), "<html>spa</html>", p)
	}
	require.Equal(t, "<html>spa</html>", get(h, "/some/page").Body.String())
}

// C18
func TestOpenAPI_ServedMatchesCommitted(t *testing.T) {
	rec := get(newServer(t), "/api/openapi.json")
	require.Equal(t, http.StatusOK, rec.Code)
	committed, err := os.ReadFile(filepath.Join("..", "..", "openapi.json"))
	require.NoError(t, err, "app/openapi.json must be committed; run `task gen:openapi`")
	require.Equal(t, string(committed), rec.Body.String(), "run `task gen:openapi`")
	var doc struct {
		OpenAPI string `json:"openapi"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &doc))
	require.True(t, strings.HasPrefix(doc.OpenAPI, "3.1"), doc.OpenAPI)
}

// C22
func TestNew_FailsOnInvalidOperation(t *testing.T) {
	type empty struct{}
	_, err := app.New(app.Options{Logger: testkit.DiscardLogger(), Web: site,
		Register: func(api huma.API, _ deps.Deps) error {
			return op.Register(api, op.Spec{ID: "no-permission-op", Method: http.MethodGet, Path: "/api/v1/x"},
				func(context.Context, *empty) (*empty, error) { return nil, nil })
		}})
	require.Error(t, err)
	require.Contains(t, err.Error(), "no-permission-op")
}

func listen(t *testing.T) net.Listener {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	return ln
}

// C5
func TestRun_ShutdownDrainsInFlight(t *testing.T) {
	entered, release := make(chan struct{}), make(chan struct{})
	srv := &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		close(entered)
		<-release
		w.WriteHeader(http.StatusOK)
	})}
	ln := listen(t)
	addr := "http://" + ln.Addr().String()
	ctx, cancel := context.WithCancel(context.Background())
	runErr := make(chan error, 1)
	go func() { runErr <- app.Run(ctx, srv, ln, 10*time.Second, testkit.DiscardLogger()) }()

	status := make(chan int, 1)
	go func() {
		resp, err := http.Get(addr + "/slow")
		if err != nil {
			status <- -1
			return
		}
		resp.Body.Close()
		status <- resp.StatusCode
	}()
	<-entered
	cancel()

	require.Eventually(t, func() bool {
		c, err := net.DialTimeout("tcp", ln.Addr().String(), 200*time.Millisecond)
		if err != nil {
			return true
		}
		c.Close()
		return false
	}, 3*time.Second, 50*time.Millisecond, "new connections must be refused during shutdown")

	close(release)
	require.Equal(t, http.StatusOK, <-status)
	select {
	case err := <-runErr:
		require.NoError(t, err)
	case <-time.After(5 * time.Second):
		t.Fatal("Run did not return")
	}
}

// C6
func TestRun_ShutdownTimeoutBoundsDrain(t *testing.T) {
	entered := make(chan struct{})
	block := make(chan struct{})
	t.Cleanup(func() { close(block) })
	srv := &http.Server{Handler: http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		close(entered)
		<-block
	})}
	ln := listen(t)
	ctx, cancel := context.WithCancel(context.Background())
	runErr := make(chan error, 1)
	go func() { runErr <- app.Run(ctx, srv, ln, 200*time.Millisecond, testkit.DiscardLogger()) }()
	go func() {
		resp, err := http.Get("http://" + ln.Addr().String() + "/stuck")
		if err == nil {
			resp.Body.Close()
		}
	}()
	<-entered
	start := time.Now()
	cancel()
	select {
	case err := <-runErr:
		require.True(t, err == nil || errors.Is(err, http.ErrServerClosed), "%v", err)
		require.Less(t, time.Since(start), 2*time.Second)
	case <-time.After(5 * time.Second):
		t.Fatal("Run kept waiting past the shutdown timeout")
	}
}
