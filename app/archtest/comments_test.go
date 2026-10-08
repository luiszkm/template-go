package archtest_test

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/luiszkm/template-go/archtest"
)

func TestComments_RejectsCommentsOutsideAnnotations(t *testing.T) {
	v, err := archtest.CheckComments(filepath.Join("testdata", "comments"))
	require.NoError(t, err)
	require.Equal(t, []string{
		"Dockerfile:1: comment in code (AGENTS.md rule 8)",
		"bad.go:1: comment in code (AGENTS.md rule 8)",
		"bad.go:9: comment in code (AGENTS.md rule 8)",
		"bad.go:10: comment in code (AGENTS.md rule 8)",
		"bad.sql:2: comment in code (AGENTS.md rule 8)",
		"bad.yml:1: comment in code (AGENTS.md rule 8)",
		"web/bad.ts:1: comment in code (AGENTS.md rule 8)",
		"web/bad.ts:2: comment in code (AGENTS.md rule 8)",
		"web/bad.ts:3: comment in code (AGENTS.md rule 8)",
	}, v)
}

func TestComments_RepositoryIsClean(t *testing.T) {
	v, err := archtest.CheckComments(repoRoot)
	require.NoError(t, err)
	require.Empty(t, v)
}
