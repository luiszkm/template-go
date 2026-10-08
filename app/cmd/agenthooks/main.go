// Command agenthooks implements the Claude Code hooks wired in .claude/settings.json.
//
//	agenthooks gofmt   PostToolUse: gofmt the edited file when it is a .go file
//	agenthooks stop    Stop: run `task check:fast`; on failure exit 2 so Claude sees stderr
//
// Both read the hook's JSON payload from stdin.
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

// gofmtHook formats the file named in tool_input.file_path. It never blocks the agent:
// a file that does not parse is left alone and reported on stderr.
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

// stopHook runs cmd unless Claude is already continuing because of this hook
// (stop_hook_active), which would otherwise loop forever on a red check.
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
