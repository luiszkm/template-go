package me_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/luiszkm/template-go/internal/features/users/me"
	"github.com/luiszkm/template-go/internal/platform/auth"
	"github.com/luiszkm/template-go/internal/platform/op"
	"github.com/luiszkm/template-go/internal/platform/testkit"
)

type empty struct{}

type body struct {
	ID          string   `json:"id"`
	Email       string   `json:"email"`
	Name        string   `json:"name"`
	Permissions []string `json:"permissions"`
}

func TestMe_ReturnsPermissions(t *testing.T) {
	pool := testkit.MigratedDB(t)
	api, h, d := testkit.NewAPI(t, pool, testkit.APIOptions{})
	require.NoError(t, me.Register(api, d))
	noop := func(context.Context, *empty) (*empty, error) { return &empty{}, nil }
	require.NoError(t, op.Register(api, op.Spec{ID: "z", Method: http.MethodGet, Path: "/z", Permission: "zeta:do"}, noop))
	require.NoError(t, op.Register(api, op.Spec{ID: "r", Method: http.MethodGet, Path: "/r", Permission: "users:read"}, noop))

	cases := []struct {
		name string
		user testkit.User
		want []string
	}{
		{"no role", testkit.SignIn(t, pool), []string{}},
		{"two permissions", testkit.SignIn(t, pool, "users:read", "users:create"), []string{"users:create", "users:read"}},
		{"admin wildcard", testkit.SignIn(t, pool, auth.Wildcard), []string{"users:read", "zeta:do"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			rec := testkit.Do(t, h, testkit.Request{Method: http.MethodGet, Path: "/api/v1/users/me", Cookie: c.user.Cookie})
			require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
			got := testkit.JSON[body](t, rec)
			require.Equal(t, c.user.ID.String(), got.ID)
			require.Equal(t, c.user.Email, got.Email)
			require.Equal(t, "Test", got.Name)
			require.Equal(t, c.want, got.Permissions)
			require.NotContains(t, got.Permissions, "*")
		})
	}
}

func TestMe_401WithoutSession(t *testing.T) {
	pool := testkit.MigratedDB(t)
	api, h, d := testkit.NewAPI(t, pool, testkit.APIOptions{})
	require.NoError(t, me.Register(api, d))
	rec := testkit.Do(t, h, testkit.Request{Method: http.MethodGet, Path: "/api/v1/users/me"})
	require.Equal(t, http.StatusUnauthorized, rec.Code)
	require.Equal(t, "application/problem+json", rec.Header().Get("Content-Type"))
}
