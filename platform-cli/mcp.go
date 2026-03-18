package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func runMcp(args []string) error {
	mcpBin, err := findMcpBin()
	if err != nil {
		return fmt.Errorf("mcp binary not found; run `make build-mcp` first: %w", err)
	}
	cmd := exec.Command(mcpBin, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func findMcpBin() (string, error) {
	// 1. Check PATH
	if p, err := exec.LookPath("mcp"); err == nil {
		return p, nil
	}
	// 2. Check bin/mcp relative to executable
	exe, err := os.Executable()
	if err == nil {
		candidate := filepath.Join(filepath.Dir(exe), "mcp")
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
	}
	return "", errors.New("mcp binary not found")
}
