package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestCappedLogFileTruncatesBeforeLimit(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.log")
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		t.Fatal(err)
	}
	writer := &cappedLogFile{file: file, size: maxLogSize - 2}
	if _, err := writer.Write([]byte("new entry")); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(content, []byte("new entry")) {
		t.Fatalf("log content = %q, want %q", content, "new entry")
	}
}
