//go:build windows

package main

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"
)

var getSystemDirectoryW = kernel32.NewProc("GetSystemDirectoryW")

func platformClaudeCommand(ctx context.Context, claudePath string, args []string) (*exec.Cmd, error) {
	extension := strings.ToLower(filepath.Ext(claudePath))
	if extension != ".bat" && extension != ".cmd" {
		return exec.CommandContext(ctx, claudePath, args...), nil
	}

	cmdPath, err := systemCommandInterpreter()
	if err != nil {
		return nil, err
	}
	commandLine, err := windowsBatchCommandLine(claudePath, args)
	if err != nil {
		return nil, err
	}
	cmd := exec.CommandContext(ctx, cmdPath)
	cmd.SysProcAttr = &syscall.SysProcAttr{CmdLine: commandLine}
	return cmd, nil
}

func systemCommandInterpreter() (string, error) {
	buffer := make([]uint16, 32768)
	length, _, callErr := getSystemDirectoryW.Call(
		uintptr(unsafe.Pointer(&buffer[0])),
		uintptr(len(buffer)),
	)
	if length == 0 {
		return "", fmt.Errorf("GetSystemDirectoryW failed: %w", callErr)
	}
	if length >= uintptr(len(buffer)) {
		return "", fmt.Errorf("Windows system directory path is too long")
	}
	cmdPath := filepath.Join(syscall.UTF16ToString(buffer[:length]), "cmd.exe")
	if info, err := syscall.GetFileAttributes(syscall.StringToUTF16Ptr(cmdPath)); err != nil || info&syscall.FILE_ATTRIBUTE_DIRECTORY != 0 {
		return "", fmt.Errorf("trusted Windows command interpreter was not found")
	}
	return cmdPath, nil
}
