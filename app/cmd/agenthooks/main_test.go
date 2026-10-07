package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func payload(t *testing.T, v any) *bytes.Reader {
	t.Helper()
	raw, err := json.Marshal(v)
	require.NoError(t, err)
	return bytes.NewReader(raw)
}

func edit(path string) map[string]any {
	return map[string]any{"hook_event_name": "PostToolUse", "tool_name": "Edit", "tool_input": map[string]any{"file_path": path}}
}

// C50
func TestGofmtHook(t *testing.T) {
	dir := t.TempDir()
	goFile := filepath.Join(dir, "x.go")
	require.NoError(t, os.WriteFile(goFile, []byte("package x\nfunc  F( ) {\nreturn}\n"), 0o644))
	tsFile := filepath.Join(dir, "x.ts")
	tsSrc := "const  a =1\n"
	require.NoError(t, os.WriteFile(tsFile, []byte(tsSrc), 0o644))

	var stderr bytes.Buffer
	require.Equal(t, 0, gofmtHook(payload(t, edit(goFile)), &stderr))
	got, err := os.ReadFile(goFile)
	require.NoError(t, err)
	require.Equal(t, "package x\n\nfunc F() {\n\treturn\n}\n", string(got))

	require.Equal(t, 0, gofmtHook(payload(t, edit(tsFile)), &stderr))
	got, err = os.ReadFile(tsFile)
	require.NoError(t, err)
	require.Equal(t, tsSrc, string(got))
}

// helperCmd re-executes this test binary as a command that exits with the given code.
func helperCmd(code string) []string {
	return []string{os.Args[0], "-test.run=^TestHelperProcess$", "--", code}
}

func TestHelperProcess(t *testing.T) {
	if os.Getenv("AGENTHOOKS_HELPER") != "1" {
		return
	}
	code := os.Args[len(os.Args)-1]
	if code == "fail" {
		_, _ = os.Stdout.WriteString("archtest: feature must not import another feature\n")
		os.Exit(1)
	}
	os.Exit(0)
}

// C51
func TestStopHook(t *testing.T) {
	t.Setenv("AGENTHOOKS_HELPER", "1")

	t.Run("failing check exits 2 with output on stderr", func(t *testing.T) {
		var stderr bytes.Buffer
		require.Equal(t, 2, stopHook(payload(t, map[string]any{"stop_hook_active": false}), &stderr, helperCmd("fail")))
		require.Contains(t, stderr.String(), "feature must not import another feature")
	})

	t.Run("passing check exits 0", func(t *testing.T) {
		var stderr bytes.Buffer
		require.Equal(t, 0, stopHook(payload(t, map[string]any{"stop_hook_active": false}), &stderr, helperCmd("pass")))
	})

	t.Run("stop_hook_active skips the command", func(t *testing.T) {
		var stderr bytes.Buffer
		// The command fails if it runs, so exit 0 with no output proves it was skipped.
		require.Equal(t, 0, stopHook(payload(t, map[string]any{"stop_hook_active": true}), &stderr, helperCmd("fail")))
		require.Empty(t, stderr.String())
	})
}
