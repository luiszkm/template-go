package httpx_test

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/luiszkm/template-go/internal/platform/httpx"
)

const leakedCause = "pq: secret host=db user=app"

type named struct {
	Body struct {
		Name string `json:"name" minLength:"3"`
	}
}

type edge struct {
	h     http.Handler
	log   *logBuf
	calls *atomic.Int32
}

func wrapped(t *testing.T, hsts bool) edge {
	t.Helper()
	buf := &logBuf{}
	log := slog.New(slog.NewJSONHandler(buf, nil))
	mux := http.NewServeMux()
	api := httpx.NewAPI(mux)
	var calls atomic.Int32
	reg := func(id, method, path string, h func(context.Context, *struct{}) (*struct{}, error)) {
		huma.Register(api, huma.Operation{OperationID: id, Method: method, Path: path}, h)
	}
	reg("plain", http.MethodGet, "/api/plain", func(context.Context, *struct{}) (*struct{}, error) {
		return nil, errors.New(leakedCause)
	})
	reg("e500", http.MethodGet, "/api/e500", func(context.Context, *struct{}) (*struct{}, error) {
		return nil, huma.Error500InternalServerError("x", errors.New("secret"))
	})
	reg("e502", http.MethodGet, "/api/e502", func(context.Context, *struct{}) (*struct{}, error) {
		return nil, huma.Error502BadGateway("x", errors.New("secret"))
	})
	reg("e503", http.MethodGet, "/api/e503", func(context.Context, *struct{}) (*struct{}, error) {
		return nil, huma.Error503ServiceUnavailable("x", errors.New("secret"))
	})
	reg("down", http.MethodGet, "/api/down", func(context.Context, *struct{}) (*struct{}, error) {
		return nil, huma.Error503ServiceUnavailable("db down")
	})
	huma.Register(api, huma.Operation{OperationID: "named", Method: http.MethodPost, Path: "/api/named"},
		func(context.Context, *named) (*struct{}, error) { return nil, nil })
	count := func(context.Context, *struct{}) (*struct{}, error) { calls.Add(1); return nil, nil }
	for _, m := range []string{http.MethodGet, http.MethodHead, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete} {
		reg("thing-"+strings.ToLower(m), m, "/api/thing", count)
	}
	mux.HandleFunc("GET /api/panic", func(http.ResponseWriter, *http.Request) { panic("boom") })
	mux.Handle("/api/", httpx.NotFound())
	return edge{h: httpx.Wrap(mux, httpx.Options{Logger: log, HSTS: hsts}), log: buf, calls: &calls}
}

func send(h http.Handler, method, path string, hdr map[string]string, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequestWithContext(context.Background(), method, path, strings.NewReader(body))
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	for k, v := range hdr {
		if k == "Host" {
			req.Host = v
			continue
		}
		req.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func problemOf(t *testing.T, rec *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	require.Equal(t, "application/problem+json", rec.Header().Get("Content-Type"), rec.Body.String())
	return decode(t, rec.Body.Bytes())
}

func onlyEntry(t *testing.T, b *logBuf) map[string]any {
	t.Helper()
	entries := b.entries(t)
	require.Len(t, entries, 1)
	return entries[0]
}

func TestProblem_5xxHidesCause(t *testing.T) {
	e := wrapped(t, false)
	rec := send(e.h, http.MethodGet, "/api/plain", nil, "")
	require.Equal(t, http.StatusInternalServerError, rec.Code)
	body := problemOf(t, rec)
	require.Equal(t, "internal server error", body["detail"])
	require.NotContains(t, body, "errors")
	require.NotContains(t, rec.Body.String(), "secret")
}

func TestProblem_No5xxCarriesErrors(t *testing.T) {
	e := wrapped(t, false)
	for path, status := range map[string]int{"/api/e500": 500, "/api/e502": 502, "/api/e503": 503} {
		rec := send(e.h, http.MethodGet, path, nil, "")
		require.Equal(t, status, rec.Code, path)
		require.NotContains(t, problemOf(t, rec), "errors", path)
		require.NotContains(t, rec.Body.String(), "secret", path)
	}
}

func TestAccessLog_5xxLogsCauseAtError(t *testing.T) {
	e := wrapped(t, false)
	rec := send(e.h, http.MethodGet, "/api/plain", nil, "")
	entry := onlyEntry(t, e.log)
	require.Equal(t, "ERROR", entry["level"])
	require.Equal(t, 500, intField(t, entry, "status"))
	require.Equal(t, rec.Header().Get(httpx.HeaderRequestID), entry["request_id"])
	require.Equal(t, leakedCause, entry["error"])
}

func TestAccessLog_5xxWithoutCause(t *testing.T) {
	e := wrapped(t, false)
	require.Equal(t, http.StatusServiceUnavailable, send(e.h, http.MethodGet, "/api/down", nil, "").Code)
	entry := onlyEntry(t, e.log)
	require.Equal(t, "ERROR", entry["level"])
	require.NotContains(t, entry, "error")
}

func TestAccessLog_Below500IsInfo(t *testing.T) {
	e := wrapped(t, false)
	require.Equal(t, http.StatusNotFound, send(e.h, http.MethodGet, "/api/nope", nil, "").Code)
	rec := send(e.h, http.MethodPost, "/api/named", nil, `{"name":"a"}`)
	require.Equal(t, http.StatusUnprocessableEntity, rec.Code)
	errs, ok := problemOf(t, rec)["errors"].([]any)
	require.True(t, ok, rec.Body.String())
	require.Len(t, errs, 1)
	first, ok := errs[0].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "body.name", first["location"])

	entries := e.log.entries(t)
	require.Len(t, entries, 2)
	for _, entry := range entries {
		require.Equal(t, "INFO", entry["level"])
		require.NotContains(t, entry, "error")
	}
}

func TestCrossOrigin_RejectsEveryUnsafeMethod(t *testing.T) {
	e := wrapped(t, false)
	for _, m := range []string{http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete} {
		rec := send(e.h, m, "/api/thing", map[string]string{"Sec-Fetch-Site": "cross-site"}, "")
		require.Equal(t, http.StatusForbidden, rec.Code, m)
		require.Equal(t, "cross-origin request rejected", problemOf(t, rec)["detail"], m)
	}
	require.Zero(t, e.calls.Load())
}

func TestCrossOrigin_OriginMustMatchHost(t *testing.T) {
	e := wrapped(t, false)
	rec := send(e.h, http.MethodPost, "/api/thing", map[string]string{"Host": "app.local", "Origin": "https://evil.example"}, "")
	require.Equal(t, http.StatusForbidden, rec.Code)
	require.Zero(t, e.calls.Load())

	rec = send(e.h, http.MethodPost, "/api/thing", map[string]string{"Host": "app.local", "Origin": "http://app.local"}, "")
	require.Less(t, rec.Code, 300, rec.Body.String())
	require.EqualValues(t, 1, e.calls.Load())
}

func TestCrossOrigin_AllowsSameOriginAndHeaderless(t *testing.T) {
	e := wrapped(t, false)
	for _, hdr := range []map[string]string{
		{"Sec-Fetch-Site": "same-origin"},
		{"Sec-Fetch-Site": "none"},
		nil,
	} {
		rec := send(e.h, http.MethodPost, "/api/thing", hdr, "")
		require.Less(t, rec.Code, 300, "%v: %s", hdr, rec.Body.String())
	}
	require.EqualValues(t, 3, e.calls.Load())
}

func TestCrossOrigin_SafeMethodsPass(t *testing.T) {
	e := wrapped(t, false)
	for _, m := range []string{http.MethodGet, http.MethodHead} {
		rec := send(e.h, m, "/api/thing", map[string]string{"Sec-Fetch-Site": "cross-site"}, "")
		require.Less(t, rec.Code, 300, m)
	}
	require.EqualValues(t, 2, e.calls.Load())
}

func TestSecurityHeaders_OnPanic500(t *testing.T) {
	e := wrapped(t, false)
	rec := send(e.h, http.MethodGet, "/api/panic", nil, "")
	require.Equal(t, http.StatusInternalServerError, rec.Code)
	require.Equal(t, "nosniff", rec.Header().Get("X-Content-Type-Options"))
	require.Equal(t, "DENY", rec.Header().Get("X-Frame-Options"))
	require.Equal(t, "strict-origin-when-cross-origin", rec.Header().Get("Referrer-Policy"))
}

func TestRequestID_AcceptsValidFormat(t *testing.T) {
	e := wrapped(t, false)
	for _, id := range []string{"abc-123", "A.b_C-9", uuid.NewString(), strings.Repeat("a", 64)} {
		rec := send(e.h, http.MethodGet, "/api/nope", map[string]string{httpx.HeaderRequestID: id}, "")
		require.Equal(t, id, rec.Header().Get(httpx.HeaderRequestID))
		require.Equal(t, id, problemOf(t, rec)["request_id"])
	}
}

func TestRequestID_RejectsInvalidFormat(t *testing.T) {
	for _, id := range []string{"", strings.Repeat("a", 65), "a b", "a\tb", `a"b`, "ação", `{"x":1}`} {
		e := wrapped(t, false)
		rec := send(e.h, http.MethodGet, "/api/nope", map[string]string{httpx.HeaderRequestID: id}, "")
		got := rec.Header().Get(httpx.HeaderRequestID)
		_, err := uuid.Parse(got)
		require.NoError(t, err, "%q replaced by %q", id, got)
		require.Equal(t, got, problemOf(t, rec)["request_id"], id)
		require.Equal(t, got, onlyEntry(t, e.log)["request_id"], id)
	}
}

func decode(t *testing.T, raw []byte) map[string]any {
	t.Helper()
	var m map[string]any
	require.NoError(t, json.Unmarshal(raw, &m), string(raw))
	return m
}
