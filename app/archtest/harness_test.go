package archtest_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

var repoRoot = filepath.Join("..", "..")

func readRepo(t *testing.T, rel string) string {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(repoRoot, rel))
	require.NoError(t, err)
	return string(raw)
}

type taskfile struct {
	Tasks map[string]struct {
		Cmds []any  `yaml:"cmds"`
		Dir  string `yaml:"dir"`
	} `yaml:"tasks"`
}

func loadTaskfile(t *testing.T) taskfile {
	t.Helper()
	var tf taskfile
	require.NoError(t, yaml.Unmarshal([]byte(readRepo(t, "Taskfile.yml")), &tf))
	return tf
}

// calledTasks returns the names of `- task: <name>` entries, in order.
func calledTasks(t *testing.T, tf taskfile, name string) []string {
	t.Helper()
	task, ok := tf.Tasks[name]
	require.True(t, ok, "task %q missing", name)
	var out []string
	for _, c := range task.Cmds {
		m, ok := c.(map[string]any)
		require.True(t, ok, "%s: every step must be `- task: <name>`, got %v", name, c)
		out = append(out, m["task"].(string))
	}
	return out
}

func commands(t *testing.T, tf taskfile, name string) string {
	t.Helper()
	task, ok := tf.Tasks[name]
	require.True(t, ok, "task %q missing", name)
	var out []string
	for _, c := range task.Cmds {
		switch v := c.(type) {
		case string:
			out = append(out, v)
		case map[string]any:
			if cmd, ok := v["cmd"].(string); ok {
				out = append(out, cmd)
			}
			if sub, ok := v["task"].(string); ok {
				out = append(out, commands(t, tf, sub))
			}
		}
	}
	return task.Dir + ": " + strings.Join(out, " && ")
}

// C30
func TestTaskfile_CheckRunsAllSteps(t *testing.T) {
	tf := loadTaskfile(t)
	steps := calledTasks(t, tf, "check")
	require.Equal(t, []string{
		"fmt:check", "lint", "gen:check", "archtest", "test", "web:typecheck", "web:lint", "web:test",
	}, steps)

	require.Contains(t, commands(t, tf, "fmt:check"), "cmd/fmtcheck")
	require.Contains(t, commands(t, tf, "lint"), "golangci-lint run")
	gen := commands(t, tf, "gen:check")
	require.Contains(t, gen, "sqlcrun diff")
	require.Contains(t, gen, "TestOpenAPI_ServedMatchesCommitted")
	require.Contains(t, gen, "gen:check")
	require.Contains(t, commands(t, tf, "archtest"), "./archtest")
	require.Contains(t, commands(t, tf, "test"), "go test -count=1 ./...")
	require.Contains(t, commands(t, tf, "web:typecheck"), "npm run typecheck")
	require.Contains(t, commands(t, tf, "web:lint"), "npm run lint")
	require.Contains(t, commands(t, tf, "web:test"), "npm test")
}

// C55
func TestTaskfile_CheckFastSteps(t *testing.T) {
	tf := loadTaskfile(t)
	require.Equal(t, []string{"fmt:check", "vet", "archtest", "web:typecheck"}, calledTasks(t, tf, "check:fast"))
	require.Contains(t, commands(t, tf, "vet"), "go vet ./...")
}

// C31
func TestModuleLayout(t *testing.T) {
	require.True(t, strings.HasPrefix(readRepo(t, "app/go.mod"), "module github.com/luiszkm/template-go\n"))
	tools := readRepo(t, "app/tools/go.mod")
	for _, tool := range []string{
		"github.com/sqlc-dev/sqlc/cmd/sqlc",
		"github.com/pressly/goose/v3/cmd/goose",
		"github.com/golangci/golangci-lint/v2/cmd/golangci-lint",
	} {
		require.Contains(t, tools, "\t"+tool+"\n", "tool directive for %s", tool)
	}
	require.NoFileExists(t, filepath.Join(repoRoot, "go.mod"))
}

// C32
func TestDependencies_Declared(t *testing.T) {
	gomod := readRepo(t, "app/go.mod")
	for _, mod := range []string{
		"github.com/danielgtaylor/huma/v2 ", "github.com/jackc/pgx/v5 ", "github.com/pressly/goose/v3 ",
		"github.com/caarlos0/env/v11 ", "github.com/stretchr/testify ", "github.com/testcontainers/testcontainers-go ",
		"go.opentelemetry.io/otel ",
	} {
		require.Contains(t, gomod, "\t"+mod, "app/go.mod must require %s", strings.TrimSpace(mod))
	}

	var pkg struct {
		Dependencies    map[string]string `json:"dependencies"`
		DevDependencies map[string]string `json:"devDependencies"`
	}
	require.NoError(t, json.Unmarshal([]byte(readRepo(t, "web/package.json")), &pkg))
	all := map[string]string{}
	for k, v := range pkg.Dependencies {
		all[k] = v
	}
	for k, v := range pkg.DevDependencies {
		all[k] = v
	}
	for _, name := range []string{
		"react", "vite", "@tanstack/react-router", "@tanstack/react-query", "tailwindcss", "zod",
		"react-hook-form", "openapi-typescript", "openapi-fetch", "vitest", "@testing-library/react",
		"@playwright/test", "@biomejs/biome",
	} {
		require.Contains(t, all, name)
	}
	require.True(t, strings.HasPrefix(strings.TrimLeft(all["react"], "^~"), "19."), all["react"])
	require.True(t, strings.HasPrefix(strings.TrimLeft(all["tailwindcss"], "^~"), "4."), all["tailwindcss"])
}

// C52
func TestClaudeSettings_WiresHooks(t *testing.T) {
	var settings struct {
		Hooks map[string][]struct {
			Matcher string `json:"matcher"`
			Hooks   []struct {
				Type    string `json:"type"`
				Command string `json:"command"`
			} `json:"hooks"`
		} `json:"hooks"`
	}
	require.NoError(t, json.Unmarshal([]byte(readRepo(t, ".claude/settings.json")), &settings))

	post := settings.Hooks["PostToolUse"]
	require.Len(t, post, 1)
	require.Equal(t, "Edit|Write|MultiEdit", post[0].Matcher)
	require.Len(t, post[0].Hooks, 1)
	require.Contains(t, post[0].Hooks[0].Command, "./cmd/agenthooks gofmt")

	stop := settings.Hooks["Stop"]
	require.Len(t, stop, 1)
	require.Len(t, stop[0].Hooks, 1)
	require.Contains(t, stop[0].Hooks[0].Command, "./cmd/agenthooks stop")
}

func frontMatter(t *testing.T, rel string) (map[string]any, string) {
	t.Helper()
	raw := strings.ReplaceAll(readRepo(t, rel), "\r\n", "\n")
	parts := strings.SplitN(raw, "---\n", 3)
	require.Len(t, parts, 3, "%s must start with --- front-matter ---", rel)
	var fm map[string]any
	require.NoError(t, yaml.Unmarshal([]byte(parts[1]), &fm))
	return fm, parts[2]
}

// C53
func TestAgentRuleFiles(t *testing.T) {
	fm, body := frontMatter(t, ".cursor/rules/agents.mdc")
	require.Equal(t, true, fm["alwaysApply"])
	require.Contains(t, body, "AGENTS.md")

	fm, body = frontMatter(t, ".windsurf/rules/agents.md")
	require.Equal(t, "always_on", fm["trigger"])
	require.Contains(t, body, "AGENTS.md")
}

// C54
func TestCIWorkflow_RunsCheckAndE2E(t *testing.T) {
	var wf struct {
		Jobs map[string]struct {
			ContinueOnError any `yaml:"continue-on-error"`
			Steps           []struct {
				Run             string `yaml:"run"`
				ContinueOnError any    `yaml:"continue-on-error"`
			} `yaml:"steps"`
		} `yaml:"jobs"`
	}
	require.NoError(t, yaml.Unmarshal([]byte(readRepo(t, ".github/workflows/ci.yml")), &wf))
	var found []string
	for _, job := range wf.Jobs {
		require.NotEqual(t, true, job.ContinueOnError)
		for _, s := range job.Steps {
			require.NotEqual(t, true, s.ContinueOnError, "step %q", s.Run)
			if s.Run == "task check" || s.Run == "task e2e" {
				found = append(found, s.Run)
			}
		}
	}
	require.ElementsMatch(t, []string{"task check", "task e2e"}, found)
}
