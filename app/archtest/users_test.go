package archtest_test

import (
	"encoding/json"
	"maps"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDependencies_UsersFeature(t *testing.T) {
	gomod := readRepo(t, "app/go.mod")
	var crypto string
	for line := range strings.Lines(gomod) {
		if strings.HasPrefix(line, "\tgolang.org/x/crypto ") {
			crypto = line
		}
	}
	require.NotEmpty(t, crypto, "app/go.mod must require golang.org/x/crypto")
	require.NotContains(t, crypto, "// indirect")

	var pkg struct {
		Dependencies    map[string]string `json:"dependencies"`
		DevDependencies map[string]string `json:"devDependencies"`
	}
	require.NoError(t, json.Unmarshal([]byte(readRepo(t, "web/package.json")), &pkg))
	all := map[string]string{}
	maps.Copy(all, pkg.Dependencies)
	maps.Copy(all, pkg.DevDependencies)
	require.Contains(t, all, "@radix-ui/react-dialog")
	require.Contains(t, all, "@radix-ui/react-label")
}

func TestAgentsDoc_MentionsUsersContract(t *testing.T) {
	agents := readRepo(t, "AGENTS.md")
	require.Contains(t, agents, "Authenticated")
	require.Contains(t, agents, "api users create-admin")
}
