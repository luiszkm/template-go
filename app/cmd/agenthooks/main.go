package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"go/format"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: agenthooks gofmt|stop")
		os.Exit(1)
	}
	var code int
	switch os.Args[1] {
	case "gofmt":
		code = gofmtHook(os.Stdin, os.Stderr)
	case "stop":
		code = stopHook(os.Stdin, os.Stderr, []string{"task", "check:fast"})
	default:
		fmt.Fprintln(os.Stderr, "agenthooks: unknown hook", os.Args[1])
		code = 1
	}
	os.Exit(code)
}

func gofmtHook(stdin io.Reader, stderr io.Writer) int {
	var payload struct {
		ToolInput struct {
			FilePath string `json:"file_path"`
		} `json:"tool_input"`
	}
	if err := json.NewDecoder(stdin).Decode(&payload); err != nil {
		fmt.Fprintln(stderr, "agenthooks gofmt: bad payload:", err)
		return 0
	}
	path := payload.ToolInput.FilePath
	if !strings.EqualFold(filepath.Ext(path), ".go") {
		return 0
	}
	src, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintln(stderr, "agenthooks gofmt:", err)
		return 0
	}
	out, err := format.Source(src)
	if err != nil {
		fmt.Fprintf(stderr, "agenthooks gofmt: %s does not parse: %v\n", path, err)
		return 0
	}
	if !bytes.Equal(src, out) {
		if err := os.WriteFile(path, out, 0o644); err != nil {
			fmt.Fprintln(stderr, "agenthooks gofmt:", err)
		}
	}
	return 0
}

func stopHook(stdin io.Reader, stderr io.Writer, cmd []string) int {
	var payload struct {
		StopHookActive bool `json:"stop_hook_active"`
	}
	_ = json.NewDecoder(stdin).Decode(&payload)
	if payload.StopHookActive {
		return 0
	}
	c := exec.CommandContext(context.Background(), cmd[0], cmd[1:]...)
	out, err := c.CombinedOutput()
	if err != nil {
		fmt.Fprintf(stderr, "%s failed - fix before finishing:\n%s\n", strings.Join(cmd, " "), out)
		return 2
	}
	return 0
}
