package archtest_test

import (
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDependencies_NoOpenTelemetry(t *testing.T) {
	out, err := exec.CommandContext(t.Context(), "go", "list", "-deps", "../cmd/api").CombinedOutput()
	require.NoError(t, err, string(out))
	for pkg := range strings.Lines(string(out)) {
		require.False(t, strings.HasPrefix(pkg, "go.opentelemetry.io/"), "cmd/api links %s", strings.TrimSpace(pkg))
	}
	for line := range strings.Lines(readRepo(t, "app/go.mod")) {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "go.opentelemetry.io/") {
			require.True(t, strings.HasSuffix(line, "// indirect"), "app/go.mod requires %s directly", line)
		}
	}
}

func TestTaskfile_CheckRunsVuln(t *testing.T) {
	tf := loadTaskfile(t)
	require.Contains(t, calledTasks(t, tf, "check"), "vuln")
	require.Contains(t, commands(t, tf, "vuln"), "govulncheck ./...")
	require.Equal(t, "app", tf.Tasks["vuln"].Dir)
	require.Contains(t, readRepo(t, "app/tools/go.mod"), "\tgolang.org/x/vuln/cmd/govulncheck\n")
}
