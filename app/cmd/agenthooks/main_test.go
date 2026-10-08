package main

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
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
		require.Equal(t, 0, stopHook(payload(t, map[string]any{"stop_hook_active": true}), &stderr, helperCmd("fail")))
		require.Empty(t, stderr.String())
	})
}

func TestGofmtHook_NeverBlocks(t *testing.T) {
	t.Run("malformed JSON payload", func(t *testing.T) {
		var stderr bytes.Buffer
		require.Equal(t, 0, gofmtHook(strings.NewReader(`{"tool_input": {"file_path": `), &stderr))
		require.NotEmpty(t, stderr.String())
	})

	t.Run("file that does not exist", func(t *testing.T) {
		dir := t.TempDir()
		missing := filepath.Join(dir, "missing.go")
		var stderr bytes.Buffer
		require.Equal(t, 0, gofmtHook(payload(t, edit(missing)), &stderr))
		require.NotEmpty(t, stderr.String())
		require.NoFileExists(t, missing)
		entries, err := os.ReadDir(dir)
		require.NoError(t, err)
		require.Empty(t, entries)
	})

	t.Run(".go file that does not parse", func(t *testing.T) {
		goFile := filepath.Join(t.TempDir(), "broken.go")
		src := []byte("package x\nfunc  F( {\nreturn}\n")
		require.NoError(t, os.WriteFile(goFile, src, 0o644))
		var stderr bytes.Buffer
		require.Equal(t, 0, gofmtHook(payload(t, edit(goFile)), &stderr))
		require.NotEmpty(t, stderr.String())
		got, err := os.ReadFile(goFile)
		require.NoError(t, err)
		require.Equal(t, src, got)
	})
}

func TestGofmtHook_WriteFailureNeverBlocks(t *testing.T) {
	goFile := filepath.Join(t.TempDir(), "x.go")
	src := []byte("package x\nfunc  F( ) {\nreturn}\n")
	require.NoError(t, os.WriteFile(goFile, src, 0o644))
	require.NoError(t, os.Chmod(goFile, 0o444))
	t.Cleanup(func() { _ = os.Chmod(goFile, 0o644) })

	var stderr bytes.Buffer
	require.Equal(t, 0, gofmtHook(payload(t, edit(goFile)), &stderr))
	require.NotEmpty(t, stderr.String())
	got, err := os.ReadFile(goFile)
	require.NoError(t, err)
	require.Equal(t, src, got)
}

func TestMain_DispatchFailuresExit1(t *testing.T) {
	bin := buildAgenthooks(t)

	for name, args := range map[string][]string{"no argument": nil, "unknown hook": {"nope"}} {
		t.Run(name, func(t *testing.T) {
			err := exec.CommandContext(t.Context(), bin, args...).Run()
			var exitErr *exec.ExitError
			require.ErrorAs(t, err, &exitErr)
			require.Equal(t, 1, exitErr.ExitCode())
		})
	}
}

func buildAgenthooks(t *testing.T) string {
	t.Helper()
	bin := filepath.Join(t.TempDir(), "agenthooks")
	if runtime.GOOS == "windows" {
		bin += ".exe"
	}
	out, err := exec.CommandContext(t.Context(), "go", "build", "-o", bin, ".").CombinedOutput()
	require.NoError(t, err, string(out))
	return bin
}

func TestMain_GofmtArmFormats(t *testing.T) {
	bin := buildAgenthooks(t)
	goFile := filepath.Join(t.TempDir(), "x.go")
	require.NoError(t, os.WriteFile(goFile, []byte("package x\nfunc  F( ) {\nreturn}\n"), 0o644))

	cmd := exec.CommandContext(t.Context(), bin, "gofmt")
	cmd.Stdin = payload(t, edit(goFile))
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, string(out))

	got, err := os.ReadFile(goFile)
	require.NoError(t, err)
	require.Equal(t, "package x\n\nfunc F() {\n\treturn\n}\n", string(got))
}

func TestMain_StopArmRunsTaskCheckFast(t *testing.T) {
	bin := buildAgenthooks(t)
	stubDir := t.TempDir()
	argsFile := filepath.Join(stubDir, "args.txt")
	if runtime.GOOS == "windows" {
		stub := "@echo off\r\n>\"" + argsFile + "\" echo %*\r\nexit /b 3\r\n"
		require.NoError(t, os.WriteFile(filepath.Join(stubDir, "task.bat"), []byte(stub), 0o755))
	} else {
		stub := "#!/bin/sh\necho \"$@\" > '" + argsFile + "'\nexit 3\n"
		require.NoError(t, os.WriteFile(filepath.Join(stubDir, "task"), []byte(stub), 0o755))
	}

	cmd := exec.CommandContext(t.Context(), bin, "stop")
	cmd.Stdin = strings.NewReader(`{"stop_hook_active":false}`)
	cmd.Env = append(os.Environ(), "PATH="+stubDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	err := cmd.Run()
	var exitErr *exec.ExitError
	require.ErrorAs(t, err, &exitErr)
	require.Equal(t, 2, exitErr.ExitCode())

	args, err := os.ReadFile(argsFile)
	require.NoError(t, err)
	require.Equal(t, "check:fast", strings.TrimSpace(string(args)))
}
