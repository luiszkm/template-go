package config

import (
	"fmt"
	"net/netip"
	"strings"
	"time"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	DatabaseURL     string        `env:"DATABASE_URL,required,notEmpty"`
	HTTPAddr        string        `env:"HTTP_ADDR" envDefault:":8080"`
	ShutdownTimeout time.Duration `env:"SHUTDOWN_TIMEOUT" envDefault:"10s"`
	ReadyTimeout    time.Duration `env:"READY_TIMEOUT" envDefault:"2s"`
	LogLevel        string        `env:"LOG_LEVEL" envDefault:"info"`

	SessionTTL     time.Duration `env:"SESSION_TTL" envDefault:"12h"`
	CookieSecure   bool          `env:"COOKIE_SECURE" envDefault:"true"`
	TrustedProxies Prefixes      `env:"TRUSTED_PROXIES"`
}

type Prefixes []netip.Prefix

func (p *Prefixes) UnmarshalText(text []byte) error {
	var out Prefixes
	for part := range strings.SplitSeq(string(text), ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if !strings.Contains(part, "/") {
			addr, err := netip.ParseAddr(part)
			if err != nil {
				return fmt.Errorf("TRUSTED_PROXIES: %w", err)
			}
			out = append(out, netip.PrefixFrom(addr, addr.BitLen()))
			continue
		}
		prefix, err := netip.ParsePrefix(part)
		if err != nil {
			return fmt.Errorf("TRUSTED_PROXIES: %w", err)
		}
		out = append(out, prefix.Masked())
	}
	*p = out
	return nil
}

func Load(environ map[string]string) (Config, error) {
	return env.ParseAsWithOptions[Config](env.Options{Environment: environ})
}
