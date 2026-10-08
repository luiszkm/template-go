package listactions_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/danielgtaylor/huma/v2"
	"github.com/stretchr/testify/require"

	"github.com/luiszkm/template-go/internal/features/audit/audittest"
	listactions "github.com/luiszkm/template-go/internal/features/audit/list_actions"
	"github.com/luiszkm/template-go/internal/platform/deps"
	"github.com/luiszkm/template-go/internal/platform/op"
	"github.com/luiszkm/template-go/internal/platform/testkit"
)

type empty struct{}

func declare(api huma.API, _ deps.Deps) error {
	noop := func(context.Context, *empty) (*empty, error) { return &empty{}, nil }
	for _, s := range []op.Spec{
		{ID: "b1", Method: http.MethodPost, Path: "/b1", Permission: "x:y", AuditAction: "b.done"},
		{ID: "a", Method: http.MethodPost, Path: "/a", Permission: "x:y", AuditAction: "a.done"},
		{ID: "b2", Method: http.MethodDelete, Path: "/b2", Permission: "x:y", AuditAction: "b.done"},
	} {
		if err := op.Register(api, s, noop); err != nil {
			return err
		}
	}
	return nil
}

func TestListActions_ReturnsCatalogue(t *testing.T) {
	pool := testkit.MigratedDB(t)
	h := audittest.Serve(t, pool, listactions.Register, declare)
	caller := testkit.SignIn(t, pool, "audit:read")

	rec := testkit.Do(t, h, testkit.Request{Method: http.MethodGet, Path: "/api/v1/audit/actions", Cookie: caller.Cookie})
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	require.Equal(t, []string{"a.done", "b.done"}, testkit.JSON[struct {
		Items []string `json:"items"`
	}](t, rec).Items)
}
