package httpx

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
)

const (
	APITitle   = "Template API"
	APIVersion = "1.0.0"
)

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
