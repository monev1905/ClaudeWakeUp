//go:build windows

package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unsafe"
)

var getSystemDirectoryW = kernel32.NewProc("GetSystemDirectoryW")

func platformClaudeCommand(ctx context.Context, claudePath string, args []string) (*exec.Cmd, error) {
	extension := strings.ToLower(filepath.Ext(claudePath))
	if extension != ".bat" && extension != ".cmd" {
		cmd := exec.CommandContext(ctx, claudePath, args...)
		configureProcessTreeCancellation(cmd)
		return cmd, nil
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
	configureProcessTreeCancellation(cmd)
	return cmd, nil
}

func systemCommandInterpreter() (string, error) {
	systemDir, err := windowsSystemDirectory()
	if err != nil {
		return "", err
	}
	return verifiedSystemExecutable(systemDir, "cmd.exe")
}

func windowsSystemDirectory() (string, error) {
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
	return syscall.UTF16ToString(buffer[:length]), nil
}

func verifiedSystemExecutable(systemDirectory, name string) (string, error) {
	path := filepath.Join(systemDirectory, name)
	pointer, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return "", err
	}
	attributes, err := syscall.GetFileAttributes(pointer)
	if err != nil || attributes&syscall.FILE_ATTRIBUTE_DIRECTORY != 0 {
		return "", fmt.Errorf("Windows system executable %s was not found", name)
	}
	return path, nil
}

func configureProcessTreeCancellation(cmd *exec.Cmd) {
	cmd.WaitDelay = 5 * time.Second
	cmd.Cancel = func() error {
		if cmd.Process == nil {
			return os.ErrProcessDone
		}

		systemDir, err := windowsSystemDirectory()
		if err == nil {
			if taskkillPath, pathErr := verifiedSystemExecutable(systemDir, "taskkill.exe"); pathErr == nil {
				killCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				killer := exec.CommandContext(killCtx, taskkillPath, "/PID", strconv.Itoa(cmd.Process.Pid), "/T", "/F")
				killer.Stdout = io.Discard
				killer.Stderr = io.Discard
				if killErr := killer.Run(); killErr == nil {
					return nil
				}
			}
		}
		return cmd.Process.Kill()
	}
}
