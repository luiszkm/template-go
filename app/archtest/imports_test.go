package archtest_test

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/luiszkm/template-go/archtest"
)

func TestImports_RejectsCrossFeature(t *testing.T) {
	v, err := archtest.CheckImports(filepath.Join("testdata", "crossfeature"))
	require.NoError(t, err)
	require.Len(t, v, 1)
	require.Contains(t, v[0], "example.com/fix/internal/features/a")
	require.Contains(t, v[0], "example.com/fix/internal/features/b")
}

func TestImports_RejectsPlatformToFeature(t *testing.T) {
	v, err := archtest.CheckImports(filepath.Join("testdata", "platformtofeature"))
	require.NoError(t, err)
	require.Len(t, v, 1)
	require.Contains(t, v[0], "example.com/fix/internal/platform/x")
	require.Contains(t, v[0], "example.com/fix/internal/features/a")
}

func TestImports_AllowsSameFeatureAndCompositionRoot(t *testing.T) {
	v, err := archtest.CheckImports(filepath.Join("testdata", "allowed"))
	require.NoError(t, err)
	require.Empty(t, v)
}

func TestImports_RepositoryIsClean(t *testing.T) {
	v, err := archtest.CheckImports("..")
	require.NoError(t, err)
	require.Empty(t, v)
}

func TestMigrationNames_RejectsNonTimestamp(t *testing.T) {
	bad, err := archtest.CheckMigrationNames(filepath.Join("testdata", "migrations"))
	require.NoError(t, err)
	require.Len(t, bad, 1)
	require.Contains(t, bad[0], "001_bad.sql")
}

func TestMigrationNames_RepositoryIsClean(t *testing.T) {
	bad, err := archtest.CheckMigrationNames(filepath.Join("..", "migrations"))
	require.NoError(t, err)
	require.Empty(t, bad)
}
