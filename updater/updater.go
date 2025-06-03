package updater

import (
	"fmt"
	"os"

	"github.com/arduino/go-updater/releaser"
)

type Version string

func (v Version) String() string {
	return string(v)
}

// Start checks if an update has to be downloaded and if so returns the path to the
// binary to be executed to perform the update.
// If the update is not available, it returns an empty string and no error.
// If the update is available, it returns the path to the binary to be executed.
// If there is an error, it returns the error.
func CheckForUpdates(targetPath string, current Version, client *releaser.Client) (string, error) {
	return checkForUpdates(targetPath, current, client)
}

func Restart(executable string) error {
	err := execApp(executable)
	if err != nil {
		return fmt.Errorf("could not exec app: %w", err)
	}
	// TODO: allow to define custom "exit" code to be used in the wail app runtime.quit()
	os.Exit(0)
	return nil
}
