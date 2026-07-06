//go:build windows

package main

import (
	"golang.org/x/sys/windows"
)

var (
	kernel32           = windows.NewLazySystemDLL("kernel32.dll")
	procAttachConsole  = kernel32.NewProc("AttachConsole")
	procGetConsoleWindow = kernel32.NewProc("GetConsoleWindow")
)

// attachConsole attaches to the parent console for CLI mode on Windows.
// Wails builds GUI-subsystem executables by default, which have no console.
// 只在当前进程没有控制台窗口时才 attach，避免覆盖已有的 stdout/stderr。
func attachConsole() {
	const ATTACH_PARENT_PROCESS = ^uintptr(0) // (DWORD)-1

	// 检查是否已有控制台窗口
	hwnd, _, _ := procGetConsoleWindow.Call()
	if hwnd != 0 {
		return // 已有控制台，不需要 attach
	}

	// 无控制台（GUI 模式启动），attach 到父进程控制台
	procAttachConsole.Call(ATTACH_PARENT_PROCESS)
}
