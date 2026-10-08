package httpx_test

import (
	"net/http"
	"net/http/httptest"
	"net/netip"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/luiszkm/template-go/internal/platform/httpx"
)

func clientIP(t *testing.T, trusted []netip.Prefix, peer string, xff ...string) netip.Addr {
	t.Helper()
	var got netip.Addr
	h := httpx.WithClientIP(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		got = httpx.ClientIP(r.Context())
	}), trusted)
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
	req.RemoteAddr = peer
	for _, v := range xff {
		req.Header.Add("X-Forwarded-For", v)
	}
	h.ServeHTTP(httptest.NewRecorder(), req)
	return got
}

func TestClientIP_TrustedPeerReadsForwardedFor(t *testing.T) {
	trusted := []netip.Prefix{netip.MustParsePrefix("10.0.0.0/8")}
	cases := []struct{ name, xff, want string }{
		{"untrusted client before trusted hop", "203.0.113.9, 10.0.0.7", "203.0.113.9"},
		{"rightmost untrusted wins over a spoofed left entry", "1.1.1.1, 203.0.113.9", "203.0.113.9"},
		{"all hops trusted gives the leftmost", "10.0.0.8, 10.0.0.7", "10.0.0.8"},
		{"unparsable entry gives the peer", "203.0.113.9, garbage", "10.0.0.5"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			require.Equal(t, netip.MustParseAddr(c.want), clientIP(t, trusted, "10.0.0.5:4321", c.xff))
		})
	}
	require.Equal(t, netip.MustParseAddr("10.0.0.5"), clientIP(t, trusted, "10.0.0.5:4321"), "no header gives the peer")
}

func TestClientIP_UntrustedPeerIgnoresForwardedFor(t *testing.T) {
	trusted := []netip.Prefix{netip.MustParsePrefix("10.0.0.0/8")}
	require.Equal(t, netip.MustParseAddr("198.51.100.4"), clientIP(t, trusted, "198.51.100.4:1234", "1.2.3.4"))
	require.Equal(t, netip.MustParseAddr("10.0.0.5"), clientIP(t, nil, "10.0.0.5:1234", "1.2.3.4"))
	require.Equal(t, netip.MustParseAddr("::1"), clientIP(t, nil, "[::1]:1234", "1.2.3.4"))
}
