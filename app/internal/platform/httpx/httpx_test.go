package httpx_test

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/luiszkm/template-go/internal/platform/httpx"
)

type logBuf struct{ bytes.Buffer }

// intField returns m[key] as an int, failing unless the JSON value is an integer.
func intField(t *testing.T, m map[string]any, key string) int {
	t.Helper()
	raw, err := json.Marshal(m[key])
	require.NoError(t, err)
	var n int
	require.NoError(t, json.Unmarshal(raw, &n), "%s must be an integer, got %s", key, raw)
	return n
}

func (b *logBuf) entries(t *testing.T) []map[string]any {
	t.Helper()
	var out []map[string]any
	for line := range strings.SplitSeq(strings.TrimSpace(b.String()), "\n") {
		if line == "" {
			continue
		}
		var m map[string]any
		require.NoError(t, json.Unmarshal([]byte(line), &m), line)
		out = append(out, m)
	}
	return out
}

// server mirrors the production chain: Huma API with Problem errors, a /api/ 404 catch-all and a panicking route.
func server(t *testing.T) (http.Handler, *logBuf) {
	t.Helper()
	buf := &logBuf{}
	log := slog.New(slog.NewJSONHandler(buf, nil))
	mux := http.NewServeMux()
	api := httpx.NewAPI(mux)
	huma.Register(api, huma.Operation{OperationID: "unavailable", Method: http.MethodGet, Path: "/api/unavailable"},
		func(context.Context, *struct{}) (*struct{}, error) {
			return nil, huma.Error503ServiceUnavailable("db down")
		})
	mux.HandleFunc("GET /api/panic", func(http.ResponseWriter, *http.Request) { panic("boom") })
	mux.Handle("/api/", httpx.NotFound())
	return httpx.Chain(mux, log), buf
}

func do(h http.Handler, method, path string, hdr map[string]string) *httptest.ResponseRecorder {
	req := httptest.NewRequestWithContext(context.Background(), method, path, nil)
	for k, v := range hdr {
		req.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

// C11
func TestProblem_HasRequiredFields(t *testing.T) {
	h, _ := server(t)
	cases := map[int]string{404: "/api/nope", 500: "/api/panic", 503: "/api/unavailable"}
	for status, path := range cases {
		rec := do(h, http.MethodGet, path, nil)
		require.Equal(t, status, rec.Code, path)
		require.Equal(t, "application/problem+json", rec.Header().Get("Content-Type"), path)
		var body map[string]any
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body), path)
		for _, field := range []string{"type", "title", "status", "detail", "request_id"} {
			require.Contains(t, body, field, "%s missing %s", path, field)
		}
		require.Equal(t, status, intField(t, body, "status"), path)
		require.Equal(t, rec.Header().Get(httpx.HeaderRequestID), body["request_id"], path)
		require.NotEmpty(t, body["request_id"], path)
	}
}

// C13
func TestRecover_500WithoutStack(t *testing.T) {
	h, _ := server(t)
	rec := do(h, http.MethodGet, "/api/panic", nil)
	require.Equal(t, http.StatusInternalServerError, rec.Code)
	require.Equal(t, "application/problem+json", rec.Header().Get("Content-Type"))
	var body map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	detail, _ := body["detail"].(string)
	require.NotContains(t, detail, "goroutine")
	require.NotContains(t, detail, ".go:")
	require.NotContains(t, rec.Body.String(), "goroutine")
}

// C14
func TestRecover_LogsErrorWithRequestID(t *testing.T) {
	h, buf := server(t)
	rec := do(h, http.MethodGet, "/api/panic", nil)
	var errs []map[string]any
	for _, e := range buf.entries(t) {
		if e["level"] == "ERROR" {
			errs = append(errs, e)
		}
	}
	require.Len(t, errs, 1)
	require.Equal(t, rec.Header().Get(httpx.HeaderRequestID), errs[0]["request_id"])
}

// C15
func TestRequestID_EchoesIncoming(t *testing.T) {
	h, _ := server(t)
	rec := do(h, http.MethodGet, "/api/nope", map[string]string{httpx.HeaderRequestID: "abc-123"})
	require.Equal(t, "abc-123", rec.Header().Get(httpx.HeaderRequestID))
}

// C16
func TestRequestID_GeneratesUUID(t *testing.T) {
	h, _ := server(t)
	a := do(h, http.MethodGet, "/api/nope", nil).Header().Get(httpx.HeaderRequestID)
	b := do(h, http.MethodGet, "/api/nope", nil).Header().Get(httpx.HeaderRequestID)
	_, err := uuid.Parse(a)
	require.NoError(t, err, a)
	_, err = uuid.Parse(b)
	require.NoError(t, err, b)
	require.NotEqual(t, a, b)
}

// C17
func TestAccessLog_HasAllKeys(t *testing.T) {
	h, buf := server(t)
	do(h, http.MethodGet, "/api/unavailable", map[string]string{httpx.HeaderRequestID: "r-1"})
	entries := buf.entries(t)
	require.Len(t, entries, 1)
	e := entries[0]
	for _, k := range []string{"time", "level", "msg", "request_id", "method", "path", "status", "duration_ms"} {
		require.Contains(t, e, k)
	}
	require.Equal(t, "r-1", e["request_id"])
	require.Equal(t, "GET", e["method"])
	require.Equal(t, "/api/unavailable", e["path"])
	require.Equal(t, http.StatusServiceUnavailable, intField(t, e, "status"))
}

// C57
func TestProblem_InstanceIsRequestPath(t *testing.T) {
	h, _ := server(t)
	cases := map[int]string{404: "/api/nope", 500: "/api/panic", 503: "/api/unavailable"}
	for status, path := range cases {
		rec := do(h, http.MethodGet, path, nil)
		require.Equal(t, status, rec.Code, path)
		require.Equal(t, "application/problem+json", rec.Header().Get("Content-Type"), path)
		var body map[string]any
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body), path)
		require.Equal(t, path, body["instance"], path)
	}
}

// C61
func TestRecover_RepanicsAbortHandler(t *testing.T) {
	buf := &logBuf{}
	log := slog.New(slog.NewJSONHandler(buf, nil))
	h := httpx.Recover(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic(http.ErrAbortHandler)
	}), log)
	rec := httptest.NewRecorder()

	var got any
	func() {
		defer func() { got = recover() }()
		h.ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/abort", nil))
	}()

	require.Equal(t, http.ErrAbortHandler, got, "Recover must re-panic http.ErrAbortHandler")
	require.Empty(t, rec.Body.String(), "no Problem body may be written")
	require.NotEqual(t, http.StatusInternalServerError, rec.Code)
	for _, e := range buf.entries(t) {
		require.NotEqual(t, "ERROR", e["level"], "no ERROR entry may be logged")
	}
}
