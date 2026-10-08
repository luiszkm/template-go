package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"

	"gopkg.in/yaml.v3"
)

func main() {
	raw, err := os.ReadFile("sqlc.yaml")
	if err != nil {
		fmt.Fprintln(os.Stderr, "sqlcrun:", err)
		os.Exit(1)
	}
	var cfg struct {
		SQL []any `yaml:"sql"`
	}
	if err := yaml.Unmarshal(raw, &cfg); err != nil {
		fmt.Fprintln(os.Stderr, "sqlcrun: sqlc.yaml:", err)
		os.Exit(1)
	}
	if len(cfg.SQL) == 0 {
		fmt.Println("sqlcrun: no sqlc packages configured; nothing to do")
		return
	}
	cmd := exec.CommandContext(context.Background(), "go", append([]string{"tool", "-modfile=tools/go.mod", "sqlc"}, os.Args[1:]...)...)
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		if exit, ok := errors.AsType[*exec.ExitError](err); ok {
			os.Exit(exit.ExitCode())
		}
		fmt.Fprintln(os.Stderr, "sqlcrun:", err)
		os.Exit(1)
	}
}
