package health_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"

	"github.com/luiszkm/template-go/internal/platform/health"
	"github.com/luiszkm/template-go/internal/platform/httpx"
	"github.com/luiszkm/template-go/internal/platform/testkit"
)

// untouchable fails the test if the probe reaches the database.
type untouchable struct{ t *testing.T }

func (u untouchable) QueryRow(context.Context, string, ...any) pgx.Row {
	u.t.Fatal("healthz must not touch the database")
	return nil
}

func handler(t *testing.T, db health.DB, timeout time.Duration) http.Handler {
	t.Helper()
	mux := http.NewServeMux()
	require.NoError(t, health.Register(httpx.NewAPI(mux), db, timeout))
	return httpx.Chain(mux, testkit.DiscardLogger())
}

func get(h http.Handler, path string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
	return rec
}

// C1
func TestHealthz_OKWithoutDatabase(t *testing.T) {
	rec := get(handler(t, untouchable{t}, time.Second), "/healthz")
	require.Equal(t, http.StatusOK, rec.Code)
	require.JSONEq(t, `{"status":"ok"}`, rec.Body.String())
}

func pool(t *testing.T, url string) *pgxpool.Pool {
	t.Helper()
	p, err := pgxpool.New(context.Background(), url)
	require.NoError(t, err)
	t.Cleanup(p.Close)
	return p
}

// C2
func TestReadyz_ReadyWhenDatabaseAnswers(t *testing.T) {
	pg := testkit.StartPostgres(t)
	rec := get(handler(t, pool(t, pg.URL), 2*time.Second), "/readyz")
	require.Equal(t, http.StatusOK, rec.Code)
	require.JSONEq(t, `{"status":"ready"}`, rec.Body.String())
}

// C3
func TestReadyz_503WhenDatabaseDown(t *testing.T) {
	pg := testkit.StartPostgres(t)
	h := handler(t, pool(t, pg.URL), 2*time.Second)
	require.NoError(t, pg.Container.Stop(context.Background(), nil))

	start := time.Now()
	rec := get(h, "/readyz")
	require.Less(t, time.Since(start), 3*time.Second)
	require.Equal(t, http.StatusServiceUnavailable, rec.Code)
	require.Equal(t, "application/problem+json", rec.Header().Get("Content-Type"))
	var body map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.Equal(t, float64(503), body["status"])
}

// A database that never answers is cut off by the timeout, not by the driver.
func TestReadyz_503WhenDatabaseHangs(t *testing.T) {
	rec := get(handler(t, hanging{}, 200*time.Millisecond), "/readyz")
	require.Equal(t, http.StatusServiceUnavailable, rec.Code)
}

type hanging struct{}

func (hanging) QueryRow(context.Context, string, ...any) pgx.Row { return hangRow{} }

type hangRow struct{}

func (hangRow) Scan(...any) error { select {} }
