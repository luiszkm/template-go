package httpx

import (
	"encoding/json"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
)

// Problem is the single error shape of the API: RFC 9457 plus request_id.
type Problem struct {
	huma.ErrorModel
	RequestID string `json:"request_id" doc:"Correlates the error with the server logs (same value as the X-Request-ID header)."`
}

// InstallProblems makes every Huma error a *Problem with type "about:blank".
// Call once, before any API is created.
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

// ProblemTransformer stamps the request id on every Problem Huma writes.
func ProblemTransformer(ctx huma.Context, _ string, v any) (any, error) {
	if p, ok := v.(*Problem); ok && p.RequestID == "" {
		p.RequestID = RequestIDFrom(ctx.Context())
	}
	return v, nil
}

// WriteProblem writes a Problem outside of Huma (router 404, panic recovery).
func WriteProblem(w http.ResponseWriter, r *http.Request, status int, detail string) {
	p := Problem{
		ErrorModel: huma.ErrorModel{
			Type:   "about:blank",
			Title:  http.StatusText(status),
			Status: status,
			Detail: detail,
		},
		RequestID: RequestIDFrom(r.Context()),
	}
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(p)
}

// NotFound answers every unmatched path with a 404 Problem.
func NotFound() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		WriteProblem(w, r, http.StatusNotFound, "no route matches "+r.Method+" "+r.URL.Path)
	})
}
