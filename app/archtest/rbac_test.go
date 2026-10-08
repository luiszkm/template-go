package archtest_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/luiszkm/template-go/archtest"
)

func TestWebFeatures_DoNotImportEachOther(t *testing.T) {
	v, err := archtest.CheckWebFeatureImports(filepath.Join("testdata", "webfeatures", "src"))
	require.NoError(t, err)
	require.Len(t, v, 1)
	require.Contains(t, v[0], `"rbac" imports feature "users"`)

	v, err = archtest.CheckWebFeatureImports(filepath.Join(repoRoot, "web", "src"))
	require.NoError(t, err)
	require.Empty(t, v)
}

func TestWebSharedSession_Moved(t *testing.T) {
	session := readRepo(t, "web/src/lib/session.ts")
	for _, name := range []string{"export const meQuery", "export function useMe", "export function can"} {
		require.Contains(t, session, name)
	}
	problems := readRepo(t, "web/src/lib/problems.ts")
	for _, name := range []string{"export function fieldErrors", "export const forbidden"} {
		require.Contains(t, problems, name)
	}
	for _, gone := range []string{"web/src/features/users/session.ts", "web/src/features/users/problems.ts"} {
		_, err := os.Stat(filepath.Join(repoRoot, gone))
		require.ErrorIs(t, err, os.ErrNotExist, gone)
	}
}
