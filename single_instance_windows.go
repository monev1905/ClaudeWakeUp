//go:build windows

package main

import (
	"fmt"
	"syscall"
	"unsafe"
)

const errorAlreadyExists syscall.Errno = 183

var (
	kernel32     = syscall.NewLazyDLL("kernel32.dll")
	createMutexW = kernel32.NewProc("CreateMutexW")
	closeHandle  = kernel32.NewProc("CloseHandle")
)

func acquireSingleInstance() (release func(), alreadyRunning bool, err error) {
	name, err := syscall.UTF16PtrFromString(`Local\ClaudeWakeUp_8C5401B1`)
	if err != nil {
		return nil, false, err
	}
	handle, _, callErr := createMutexW.Call(0, 0, uintptr(unsafe.Pointer(name)))
	if handle == 0 {
		return nil, false, fmt.Errorf("CreateMutexW failed: %w", callErr)
	}
	if callErr == errorAlreadyExists {
		_, _, _ = closeHandle.Call(handle)
		return func() {}, true, nil
	}
	return func() { _, _, _ = closeHandle.Call(handle) }, false, nil
}
