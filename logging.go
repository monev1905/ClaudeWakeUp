package main

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
)

const maxLogSize = 1 << 20 // 1 MiB

type cappedLogFile struct {
	mu   sync.Mutex
	file *os.File
	size int64
}

func newLogger() (*log.Logger, string, func(), error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil, "", nil, err
	}
	logDir := filepath.Join(configDir, "ClaudeWakeUp")
	if err := os.MkdirAll(logDir, 0o700); err != nil {
		return nil, "", nil, err
	}
	logPath := filepath.Join(logDir, "ClaudeWakeUp.log")
	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return nil, "", nil, err
	}
	info, err := file.Stat()
	if err != nil {
		_ = file.Close()
		return nil, "", nil, err
	}
	capped := &cappedLogFile{file: file, size: info.Size()}
	writer := io.MultiWriter(os.Stdout, capped)
	return log.New(writer, "", log.Ldate|log.Ltime), logPath, func() { _ = capped.Close() }, nil
}

func (writer *cappedLogFile) Write(data []byte) (int, error) {
	writer.mu.Lock()
	defer writer.mu.Unlock()

	if writer.size+int64(len(data)) > maxLogSize {
		if err := writer.file.Truncate(0); err != nil {
			return 0, err
		}
		if _, err := writer.file.Seek(0, 0); err != nil {
			return 0, err
		}
		writer.size = 0
	}
	written, err := writer.file.Write(data)
	writer.size += int64(written)
	return written, err
}

func (writer *cappedLogFile) Close() error {
	writer.mu.Lock()
	defer writer.mu.Unlock()
	return writer.file.Close()
}
