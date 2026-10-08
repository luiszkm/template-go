package httpx

import (
	"encoding/json"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
)

type Problem struct {
	huma.ErrorModel
	RequestID string `json:"request_id" doc:"Correlates the error with the server logs (same value as the X-Request-ID header)."`
}

func InstallProblems() {
	huma.NewError = func(status int, msg string, errs ...error) huma.StatusError {
		details := make([]*huma.ErrorDetail, 0, len(errs))
		for _, err := range errs {
			if err == nil {
				continue
			}
			if d, ok := err.(huma.ErrorDetailer); ok {
				details = append(details, d.ErrorDetail())
				continue
			}
			details = append(details, &huma.ErrorDetail{Message: err.Error()})
		}
		return &Problem{ErrorModel: huma.ErrorModel{
			Type:   "about:blank",
			Title:  http.StatusText(status),
			Status: status,
			Detail: msg,
			Errors: details,
		}}
	}
}

func ProblemTransformer(ctx huma.Context, _ string, v any) (any, error) {
	if p, ok := v.(*Problem); ok {
		if p.RequestID == "" {
			p.RequestID = RequestIDFrom(ctx.Context())
		}
		if p.Instance == "" {
			p.Instance = ctx.URL().Path
		}
	}
	return v, nil
}

func WriteProblem(w http.ResponseWriter, r *http.Request, status int, detail string) {
	p := Problem{
		ErrorModel: huma.ErrorModel{
			Type:     "about:blank",
			Title:    http.StatusText(status),
			Status:   status,
			Detail:   detail,
			Instance: r.URL.Path,
		},
		RequestID: RequestIDFrom(r.Context()),
	}
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(p)
}

func NotFound() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		WriteProblem(w, r, http.StatusNotFound, "no route matches "+r.Method+" "+r.URL.Path)
	})
}
