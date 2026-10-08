package config_test

import (
	"net/netip"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/luiszkm/template-go/internal/platform/config"
)

func TestDefaults_ShutdownTimeoutIs10s(t *testing.T) {
	cfg, err := config.Load(map[string]string{"DATABASE_URL": "postgres://x"})
	require.NoError(t, err)
	require.Equal(t, 10*time.Second, cfg.ShutdownTimeout)
	require.Equal(t, 2*time.Second, cfg.ReadyTimeout)
}

func TestLoad_MissingDatabaseURLNamesVariable(t *testing.T) {
	_, err := config.Load(map[string]string{})
	require.ErrorContains(t, err, "DATABASE_URL")
}

func TestDefaults_UsersConfig(t *testing.T) {
	cfg, err := config.Load(map[string]string{"DATABASE_URL": "postgres://x"})
	require.NoError(t, err)
	require.Equal(t, 12*time.Hour, cfg.SessionTTL)
	require.True(t, cfg.CookieSecure)
	require.Empty(t, cfg.TrustedProxies)
}

func TestTrustedProxies_Parse(t *testing.T) {
	cfg, err := config.Load(map[string]string{"DATABASE_URL": "postgres://x", "TRUSTED_PROXIES": "10.0.0.0/8, 192.168.1.1/32"})
	require.NoError(t, err)
	require.Equal(t, config.Prefixes{netip.MustParsePrefix("10.0.0.0/8"), netip.MustParsePrefix("192.168.1.1/32")}, cfg.TrustedProxies)

	_, err = config.Load(map[string]string{"DATABASE_URL": "postgres://x", "TRUSTED_PROXIES": "nope"})
	require.ErrorContains(t, err, "TRUSTED_PROXIES")
}
