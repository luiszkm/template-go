package archtest_test

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/luiszkm/template-go/archtest"
)

// C23
func TestImports_RejectsCrossFeature(t *testing.T) {
	v, err := archtest.CheckImports(filepath.Join("testdata", "crossfeature"))
	require.NoError(t, err)
	require.Len(t, v, 1)
	require.Contains(t, v[0], "example.com/fix/internal/features/a")
	require.Contains(t, v[0], "example.com/fix/internal/features/b")
}

// C24
func TestImports_RejectsPlatformToFeature(t *testing.T) {
	v, err := archtest.CheckImports(filepath.Join("testdata", "platformtofeature"))
	require.NoError(t, err)
	require.Len(t, v, 1)
	require.Contains(t, v[0], "example.com/fix/internal/platform/x")
	require.Contains(t, v[0], "example.com/fix/internal/features/a")
}

// C25
func TestImports_AllowsSameFeatureAndCompositionRoot(t *testing.T) {
	v, err := archtest.CheckImports(filepath.Join("testdata", "allowed"))
	require.NoError(t, err)
	require.Empty(t, v)
}

// C26
func TestImports_RepositoryIsClean(t *testing.T) {
	v, err := archtest.CheckImports("..")
	require.NoError(t, err)
	require.Empty(t, v)
}

// C10
func TestMigrationNames_RejectsNonTimestamp(t *testing.T) {
	bad, err := archtest.CheckMigrationNames(filepath.Join("testdata", "migrations"))
	require.NoError(t, err)
	require.Len(t, bad, 1)
	require.Contains(t, bad[0], "001_bad.sql")
}

// C10
func TestMigrationNames_RepositoryIsClean(t *testing.T) {
	bad, err := archtest.CheckMigrationNames(filepath.Join("..", "migrations"))
	require.NoError(t, err)
	require.Empty(t, bad)
}
