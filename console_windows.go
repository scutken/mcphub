//go:build windows

package main

import (
	"os"

	"golang.org/x/sys/windows"
)

// attachConsole attaches to the parent console for CLI mode on Windows.
// Wails builds GUI-subsystem executables by default, which have no console.
func attachConsole() {
	const ATTACH_PARENT_PROCESS = ^uint32(0)

	windows.AttachConsole(ATTACH_PARENT_PROCESS)

	// Reopen stdout to the attached console
	stdout, _ := windows.GetStdHandle(windows.STD_OUTPUT_HANDLE)
	if stdout != windows.InvalidHandle {
		os.Stdout = os.NewFile(uintptr(stdout), "/dev/stdout")
	}
	stderr, _ := windows.GetStdHandle(windows.STD_ERROR_HANDLE)
	if stderr != windows.InvalidHandle {
		os.Stderr = os.NewFile(uintptr(stderr), "/dev/stderr")
	}
}
