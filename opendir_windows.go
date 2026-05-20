//go:build windows

package main

import "os/exec"

func openDirectory(path string) error {
	return exec.Command("explorer", path).Start()
}
