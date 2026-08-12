package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

func resolveClaudeCLI() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("cannot determine the working directory: %w", err)
	}
	executable, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("cannot determine the application directory: %w", err)
	}

	blocked := []string{cwd, filepath.Dir(executable)}
	if strings.EqualFold(filepath.Base(filepath.Dir(executable)), "dist") {
		blocked = append(blocked, filepath.Dir(filepath.Dir(executable)))
	}

	path, err := findClaudeCLI(os.Getenv("PATH"), executableExtensions(), blocked)
	if err != nil {
		return "", err
	}
	return path, nil
}

func findClaudeCLI(pathValue string, extensions, blockedDirectories []string) (string, error) {
	blocked := make([]string, 0, len(blockedDirectories))
	for _, dir := range blockedDirectories {
		if normalized, err := normalizeDirectory(dir); err == nil {
			blocked = append(blocked, normalized)
		}
	}

	for _, entry := range filepath.SplitList(pathValue) {
		entry = strings.TrimSpace(strings.Trim(entry, `"`))
		if entry == "" || !filepath.IsAbs(entry) {
			continue
		}
		dir, err := normalizeDirectory(entry)
		if err != nil || directoryIsBlocked(dir, blocked) {
			continue
		}

		for _, extension := range extensions {
			candidate := filepath.Join(dir, "claude"+extension)
			info, err := os.Stat(candidate)
			if err != nil || info.IsDir() {
				continue
			}
			absolute, err := filepath.Abs(candidate)
			if err != nil {
				continue
			}
			return filepath.Clean(absolute), nil
		}
	}

	return "", errors.New("Claude CLI was not found in a trusted absolute PATH directory")
}

func executableExtensions() []string {
	if runtime.GOOS != "windows" {
		return []string{""}
	}

	value := os.Getenv("PATHEXT")
	if value == "" {
		return []string{".com", ".exe", ".bat", ".cmd"}
	}
	var extensions []string
	for _, extension := range strings.Split(value, ";") {
		extension = strings.ToLower(strings.TrimSpace(extension))
		if extension == "" {
			continue
		}
		if !strings.HasPrefix(extension, ".") {
			extension = "." + extension
		}
		switch extension {
		case ".com", ".exe", ".bat", ".cmd":
			extensions = append(extensions, extension)
		}
	}
	if len(extensions) == 0 {
		return []string{".com", ".exe", ".bat", ".cmd"}
	}
	return extensions
}

func normalizeDirectory(dir string) (string, error) {
	absolute, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}
	if evaluated, evalErr := filepath.EvalSymlinks(absolute); evalErr == nil {
		absolute = evaluated
	}
	return filepath.Clean(absolute), nil
}

func directoryIsBlocked(directory string, blocked []string) bool {
	for _, item := range blocked {
		if pathsEqual(directory, item) {
			return true
		}
	}
	return false
}

func pathsEqual(left, right string) bool {
	if runtime.GOOS == "windows" {
		return strings.EqualFold(left, right)
	}
	return left == right
}

func secureWorkDirectory() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	workDir := filepath.Join(configDir, "ClaudeWakeUp", "workspace")
	if err := os.MkdirAll(workDir, 0o700); err != nil {
		return "", err
	}
	return workDir, nil
}

func newClaudeCommand(ctx context.Context, claudePath string, args []string) (*exec.Cmd, error) {
	if !filepath.IsAbs(claudePath) {
		return nil, errors.New("refusing to execute Claude from a non-absolute path")
	}
	return platformClaudeCommand(ctx, claudePath, args)
}
