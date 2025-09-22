//go:build windows

package updater

import (
	"fmt"

	runas "github.com/arduino/go-windows-runas"
)

func execApp(path string) error {
	exitCode, err := runas.RunElevated(path, "", []string{}, false)
	if err != nil {
		return fmt.Errorf("could not run installer: %w", err)
	}
	if exitCode != 0 {
		return fmt.Errorf("installer exited with code %d", exitCode)
	}
	return nil
}
