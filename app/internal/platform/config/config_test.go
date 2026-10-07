package config_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/luiszkm/template-go/internal/platform/config"
)

// C6: the shutdown drain defaults to exactly 10s.
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
