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
	"time"
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

	appDir := filepath.Dir(executable)
	blockedExact := []string{cwd}
	blockedTrees := []string{appDir}
	if strings.EqualFold(filepath.Base(appDir), "dist") {
		blockedTrees = append(blockedTrees, filepath.Dir(appDir))
	}

	path, err := findClaudeCLI(os.Getenv("PATH"), executableExtensions(), blockedExact, blockedTrees)
	if err != nil {
		return "", err
	}
	return path, nil
}

func findClaudeCLI(pathValue string, extensions, blockedDirectories, blockedTrees []string) (string, error) {
	blockedExact := make([]string, 0, len(blockedDirectories))
	for _, dir := range blockedDirectories {
		if normalized, err := normalizeDirectory(dir); err == nil {
			blockedExact = append(blockedExact, normalized)
		}
	}
	blockedRoots := make([]string, 0, len(blockedTrees))
	for _, dir := range blockedTrees {
		if normalized, err := normalizeDirectory(dir); err == nil {
			blockedRoots = append(blockedRoots, normalized)
		}
	}

	for _, entry := range filepath.SplitList(pathValue) {
		entry = strings.TrimSpace(strings.Trim(entry, `"`))
		if entry == "" || !filepath.IsAbs(entry) {
			continue
		}
		dir, err := normalizeDirectory(entry)
		if err != nil || directoryIsBlocked(dir, blockedExact, blockedRoots) {
			continue
		}

		for _, extension := range extensions {
			candidate := filepath.Join(dir, "claude"+extension)
			info, err := os.Stat(candidate)
			if err != nil || info.IsDir() {
				continue
			}
			resolvedCandidate := candidate
			if evaluated, evalErr := filepath.EvalSymlinks(candidate); evalErr == nil {
				resolvedCandidate = evaluated
			}
			resolvedDir, err := normalizeDirectory(filepath.Dir(resolvedCandidate))
			if err != nil || directoryIsBlocked(resolvedDir, blockedExact, blockedRoots) {
				continue
			}
			absolute, err := filepath.Abs(resolvedCandidate)
			if err != nil {
				continue
			}
			return filepath.Clean(absolute), nil
		}
	}

	return "", errors.New("Claude CLI was not found in an absolute PATH directory outside the application/project tree")
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

func directoryIsBlocked(directory string, blockedExact, blockedTrees []string) bool {
	for _, item := range blockedExact {
		if pathsEqual(directory, item) {
			return true
		}
	}
	for _, root := range blockedTrees {
		if pathIsWithin(directory, root) {
			return true
		}
	}
	return false
}

func pathIsWithin(path, root string) bool {
	relative, err := filepath.Rel(root, path)
	if err != nil || filepath.IsAbs(relative) {
		return false
	}
	return relative == "." || (relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator)))
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

var requiredClaudeFlags = []string{
	"--safe-mode",
	"--no-session-persistence",
	"--disable-slash-commands",
	"--no-chrome",
	"--tools",
	"--max-turns",
	"--permission-mode",
	"--setting-sources",
	"--strict-mcp-config",
}

func validateClaudeCapabilities(claudePath, workDir string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	cmd, err := newClaudeCommand(ctx, claudePath, []string{"--help"})
	if err != nil {
		return err
	}
	cmd.Dir = workDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("cannot inspect Claude CLI capabilities: %w", err)
	}
	help := string(output)
	var missing []string
	for _, flag := range requiredClaudeFlags {
		if !strings.Contains(help, flag) {
			missing = append(missing, flag)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("Claude CLI is missing required security options: %s", strings.Join(missing, ", "))
	}
	return nil
}
