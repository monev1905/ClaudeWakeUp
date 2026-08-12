package main

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestFindClaudeCLISkipsBlockedAndRelativeDirectories(t *testing.T) {
	blocked := t.TempDir()
	trusted := t.TempDir()
	relative := "relative-bin"

	createTestFile(t, filepath.Join(blocked, "claude.cmd"))
	createTestFile(t, filepath.Join(trusted, "claude.cmd"))

	pathValue := blocked + string(os.PathListSeparator) + relative + string(os.PathListSeparator) + trusted
	got, err := findClaudeCLI(pathValue, []string{".cmd"}, []string{blocked})
	if err != nil {
		t.Fatalf("findClaudeCLI returned an error: %v", err)
	}
	wantDir, err := normalizeDirectory(trusted)
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(wantDir, "claude.cmd")
	if got != want {
		t.Fatalf("findClaudeCLI() = %q, want %q", got, want)
	}
}

func TestFindClaudeCLIRejectsProjectCopy(t *testing.T) {
	project := t.TempDir()
	createTestFile(t, filepath.Join(project, "claude.exe"))

	_, err := findClaudeCLI(project, []string{".exe"}, []string{project})
	if err == nil {
		t.Fatal("findClaudeCLI accepted a Claude executable from the blocked project directory")
	}
}

func TestClaudeArgumentsRemainRestricted(t *testing.T) {
	want := []string{
		"-p", "wake up",
		"--tools", "",
		"--max-turns", "1",
		"--permission-mode", "plan",
		"--setting-sources", "",
		"--strict-mcp-config",
	}
	if got := claudeArguments(); !reflect.DeepEqual(got, want) {
		t.Fatalf("claudeArguments() = %#v, want %#v", got, want)
	}
}

func TestWindowsBatchCommandLineQuotesFixedArguments(t *testing.T) {
	got, err := windowsBatchCommandLine(`C:\Program Files\Claude\claude.cmd`, []string{"-p", "wake up", "--tools", ""})
	if err != nil {
		t.Fatal(err)
	}
	want := `/D /S /C ""C:\Program Files\Claude\claude.cmd" "-p" "wake up" "--tools" """`
	if got != want {
		t.Fatalf("windowsBatchCommandLine() = %q, want %q", got, want)
	}
}

func TestWindowsBatchCommandLineRejectsMetacharacterEscapes(t *testing.T) {
	if _, err := windowsBatchCommandLine("claude.cmd\rmalicious", nil); err == nil {
		t.Fatal("windowsBatchCommandLine accepted a carriage return")
	}
	if _, err := windowsBatchCommandLine("claude.cmd", []string{`unsafe"argument`}); err == nil {
		t.Fatal("windowsBatchCommandLine accepted a quote in an argument")
	}
}

func createTestFile(t *testing.T, path string) {
	t.Helper()
	if err := os.WriteFile(path, []byte("test"), 0o600); err != nil {
		t.Fatalf("cannot create test file: %v", err)
	}
}
