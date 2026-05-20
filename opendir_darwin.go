//go:build darwin

package main

import "os/exec"

func openDirectory(path string) error {
	return exec.Command("open", path).Start()
}
