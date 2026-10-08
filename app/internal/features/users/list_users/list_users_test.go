package listusers_test

import (
	"fmt"
	"net/http"
	"slices"
	"testing"

	"github.com/stretchr/testify/require"

	listusers "github.com/luiszkm/template-go/internal/features/users/list_users"
	"github.com/luiszkm/template-go/internal/features/users/userstest"
	"github.com/luiszkm/template-go/internal/platform/testkit"
)

type page struct {
	Items []struct {
		Email string `json:"email"`
	} `json:"items"`
	Total int `json:"total"`
}

func emails(p page) []string {
	out := make([]string, 0, len(p.Items))
	for _, i := range p.Items {
		out = append(out, i.Email)
	}
	return out
}

func TestListUsers_PagesByEmail(t *testing.T) {
	pool := testkit.MigratedDB(t)
	h := userstest.Serve(t, pool, testkit.APIOptions{}, listusers.Register)
	reader := testkit.SignIn(t, pool, "users:read")
	for i := range 51 {
		userstest.InsertUser(t, pool, fmt.Sprintf("u%02d@x.com", 50-i), "x")
	}

	rec := testkit.Do(t, h, testkit.Request{Method: http.MethodGet, Path: "/api/v1/users", Cookie: reader.Cookie})
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	first := testkit.JSON[page](t, rec)
	require.Equal(t, 52, first.Total)
	require.Len(t, first.Items, 50)
	require.True(t, slices.IsSorted(emails(first)), emails(first))

	rec = testkit.Do(t, h, testkit.Request{Method: http.MethodGet, Path: "/api/v1/users?offset=50", Cookie: reader.Cookie})
	require.Equal(t, http.StatusOK, rec.Code)
	last := testkit.JSON[page](t, rec)
	require.Len(t, last.Items, 2)
	all := append(emails(first), emails(last)...)
	require.True(t, slices.IsSorted(all), all)
	require.Len(t, slices.Compact(all), 52)
}

func TestListUsers_InvalidPage422(t *testing.T) {
	pool := testkit.MigratedDB(t)
	h := userstest.Serve(t, pool, testkit.APIOptions{}, listusers.Register)
	reader := testkit.SignIn(t, pool, "users:read")
	for query, want := range map[string]int{
		"limit=0": 422, "limit=101": 422, "offset=-1": 422, "limit=1": 200, "limit=100": 200,
	} {
		rec := testkit.Do(t, h, testkit.Request{Method: http.MethodGet, Path: "/api/v1/users?" + query, Cookie: reader.Cookie})
		require.Equal(t, want, rec.Code, query)
	}
}
