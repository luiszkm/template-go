package listactions

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	"github.com/luiszkm/template-go/internal/platform/deps"
	"github.com/luiszkm/template-go/internal/platform/op"
)

const permission op.Permission = "audit:read"

type Input struct{}

type AuditActionList struct {
	Items []string `json:"items"`
}

type Output struct {
	Body AuditActionList
}

func Register(api huma.API, _ deps.Deps) error {
	return op.Register(api, op.Spec{
		ID:         "list-audit-actions",
		Method:     http.MethodGet,
		Path:       "/api/v1/audit/actions",
		Summary:    "Every audit action an operation declares",
		Tags:       []string{"audit"},
		Permission: permission,
	}, func(context.Context, *Input) (*Output, error) {
		return &Output{Body: AuditActionList{Items: op.AuditActions(api)}}, nil
	})
}
