package httpx

import (
	"log/slog"
	"net/http"
	"net/netip"
	"strings"
)

const (
	spaPolicy = "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; " +
		"font-src 'self'; connect-src 'self'; object-src 'none'; base-uri 'self'; form-action 'self'; frame-ancestors 'none'"
	apiPolicy = "default-src 'none'; frame-ancestors 'none'"
	hstsValue = "max-age=31536000"
)

type Options struct {
	Logger         *slog.Logger
	TrustedProxies []netip.Prefix
	HSTS           bool
}

func Wrap(h http.Handler, o Options) http.Handler {
	return SecurityHeaders(WithClientIP(Chain(CrossOrigin(h), o.Logger), o.TrustedProxies), o.HSTS)
}

func CrossOrigin(next http.Handler) http.Handler {
	cop := http.NewCrossOriginProtection()
	cop.SetDenyHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		WriteProblem(w, r, http.StatusForbidden, "cross-origin request rejected")
	}))
	return cop.Handler(next)
}

func SecurityHeaders(next http.Handler, hsts bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		h.Set("Content-Security-Policy", policyFor(r.URL.Path))
		if hsts {
			h.Set("Strict-Transport-Security", hstsValue)
		}
		next.ServeHTTP(w, r)
	})
}

func policyFor(path string) string {
	if strings.HasPrefix(path, "/api/") || path == "/healthz" || path == "/readyz" {
		return apiPolicy
	}
	return spaPolicy
}
