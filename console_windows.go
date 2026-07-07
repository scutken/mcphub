//go:build windows

package main

import (
	"os"

	"golang.org/x/sys/windows"
)

var (
	kernel32             = windows.NewLazySystemDLL("kernel32.dll")
	procAttachConsole    = kernel32.NewProc("AttachConsole")
	procGetConsoleWindow = kernel32.NewProc("GetConsoleWindow")
	procFreeConsole      = kernel32.NewProc("FreeConsole")
	procGetFileType      = kernel32.NewProc("GetFileType")
	user32               = windows.NewLazySystemDLL("user32.dll")
	procShowWindow       = user32.NewProc("ShowWindow")
)

// Windows 句柄类型常量（GetFileType 返回值）
const (
	fileTypeUnknown = 0 // 句柄无效或不可用
	fileTypeDisk    = 1 // 文件
	fileTypeChar    = 2 // 控制台/字符设备
	fileTypePipe    = 3 // 管道
)

// handleValid 判断继承的标准句柄是否可用。
// 父 shell 重定向时传入的是管道/文件句柄，GetFileType 返回非 UNKNOWN；
// 无控制台且未重定向时句柄无效，返回 UNKNOWN。
func handleValid(fd uintptr) bool {
	t, _, _ := procGetFileType.Call(fd)
	return t != fileTypeUnknown
}

// attachConsole 在 CLI 模式下确保标准句柄可用。
// 已有控制台窗口时进程天然继承父控制台，直接返回。
// 无控制台窗口时，父 shell 通常仍通过 STARTUPINFO 传入重定向句柄（管道/文件），
// 此时必须保留这些句柄——否则用 CONOUT$ 覆盖会让输出错写到控制台屏幕缓冲区，
// 调用方读管道得到空。仅当继承的句柄无效时才 AttachConsole 并用 CONOUT$/CONIN$ 重建。
func attachConsole() {
	const ATTACH_PARENT_PROCESS = ^uintptr(0) // (DWORD)-1

	// 已有控制台窗口则无需 attach
	hwnd, _, _ := procGetConsoleWindow.Call()
	if hwnd != 0 {
		return
	}

	// 检查继承的标准句柄是否有效（管道/文件/控制台）
	stdoutValid := handleValid(os.Stdout.Fd())
	stderrValid := handleValid(os.Stderr.Fd())
	stdinValid := handleValid(os.Stdin.Fd())

	// 全部有效则保留继承句柄，无需 attach
	if stdoutValid && stderrValid && stdinValid {
		return
	}

	// 至少一个无效：attach 到父进程控制台，为无效句柄重建
	procAttachConsole.Call(ATTACH_PARENT_PROCESS)

	if !stdoutValid {
		if h, err := windows.CreateFile(
			windows.StringToUTF16Ptr("CONOUT$"),
			windows.GENERIC_READ|windows.GENERIC_WRITE,
			windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE,
			nil,
			windows.OPEN_EXISTING,
			0,
			0,
		); err == nil {
			os.Stdout = os.NewFile(uintptr(h), "/dev/stdout")
		}
	}
	if !stderrValid {
		if h, err := windows.CreateFile(
			windows.StringToUTF16Ptr("CONOUT$"),
			windows.GENERIC_READ|windows.GENERIC_WRITE,
			windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE,
			nil,
			windows.OPEN_EXISTING,
			0,
			0,
		); err == nil {
			os.Stderr = os.NewFile(uintptr(h), "/dev/stderr")
		}
	}
	if !stdinValid {
		if h, err := windows.CreateFile(
			windows.StringToUTF16Ptr("CONIN$"),
			windows.GENERIC_READ|windows.GENERIC_WRITE,
			windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE,
			nil,
			windows.OPEN_EXISTING,
			0,
			0,
		); err == nil {
			os.Stdin = os.NewFile(uintptr(h), "/dev/stdin")
		}
	}
}

// hideConsole 隐藏控制台窗口并释放控制台，供 GUI 模式使用。
// console 子系统构建会让双击运行时弹出控制台窗口，GUI 模式应立即隐藏以避免视觉干扰。
func hideConsole() {
	hwnd, _, _ := procGetConsoleWindow.Call()
	if hwnd == 0 {
		return
	}
	const SW_HIDE = 0
	procShowWindow.Call(hwnd, SW_HIDE)
	procFreeConsole.Call()
}