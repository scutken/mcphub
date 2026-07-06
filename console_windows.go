//go:build windows

package main

import (
	"os"

	"golang.org/x/sys/windows"
)

var (
	kernel32           = windows.NewLazySystemDLL("kernel32.dll")
	procAttachConsole  = kernel32.NewProc("AttachConsole")
	procGetStdHandle   = kernel32.NewProc("GetStdHandle")
)

// attachConsole attaches to the parent console for CLI mode on Windows.
// Wails builds GUI-subsystem executables by default, which have no console.
func attachConsole() {
	const ATTACH_PARENT_PROCESS = ^uintptr(0) // (DWORD)-1

	procAttachConsole.Call(ATTACH_PARENT_PROCESS)

	// Reopen stdout/stderr to the attached console
	const STD_OUTPUT_HANDLE = ^uintptr(10) // -11 as uintptr
	const STD_ERROR_HANDLE  = ^uintptr(11) // -12 as uintptr

	hOut, _, _ := procGetStdHandle.Call(STD_OUTPUT_HANDLE)
	if hOut != 0 && hOut != uintptr(windows.InvalidHandle) {
		os.Stdout = os.NewFile(hOut, "/dev/stdout")
	}
	hErr, _, _ := procGetStdHandle.Call(STD_ERROR_HANDLE)
	if hErr != 0 && hErr != uintptr(windows.InvalidHandle) {
		os.Stderr = os.NewFile(hErr, "/dev/stderr")
	}
}
