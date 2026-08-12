//go:build !windows

package main

import (
	"context"
	"os/exec"
)

func platformClaudeCommand(ctx context.Context, claudePath string, args []string) (*exec.Cmd, error) {
	return exec.CommandContext(ctx, claudePath, args...), nil
}
