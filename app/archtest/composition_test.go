package archtest_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/tools/go/packages"
)

// C22: the binary builds its server through app.New, so tests and production share one assembly.
func TestCompositionRoot_CmdAPIUsesAppNew(t *testing.T) {
	pkgs, err := packages.Load(&packages.Config{Mode: packages.NeedName | packages.NeedImports, Dir: ".."}, "./cmd/api")
	require.NoError(t, err)
	require.Len(t, pkgs, 1)
	imports := pkgs[0].Imports
	require.Contains(t, imports, "github.com/luiszkm/template-go/internal/app")
	for imp := range imports {
		require.False(t, strings.Contains(imp, "/internal/features"), "cmd/api must not wire features itself: %s", imp)
		require.False(t, strings.Contains(imp, "huma/v2"), "cmd/api must not build its own API: %s", imp)
	}
}
