//go:build !darwin && !windows

package updater

import "os/exec"

// default execApp from golang
func execApp(path string, args ...string) error {
	cmd := exec.Command(path, args...)
	return cmd.Start()
}
