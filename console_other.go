//go:build !windows

package main

func attachConsole() {
	// No-op on non-Windows platforms
}

func hideConsole() {
	// No-op on non-Windows platforms
}
