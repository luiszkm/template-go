package httpx

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
)

// API metadata published in the OpenAPI document.
const (
	APITitle   = "Template API"
	APIVersion = "1.0.0"
)

// NewAPI creates the Huma API on mux with the project conventions: Problem errors carrying
// request_id, no $schema links in bodies, contract at /api/openapi.json and docs at /api/docs.
func NewAPI(mux *http.ServeMux) huma.API {
	InstallProblems()
	cfg := huma.DefaultConfig(APITitle, APIVersion)
	cfg.CreateHooks = nil
	cfg.OpenAPIPath = "/api/openapi"
	cfg.DocsPath = "/api/docs"
	cfg.SchemasPath = "/api/schemas"
	cfg.Transformers = append(cfg.Transformers, ProblemTransformer)
	return humago.New(mux, cfg)
}
