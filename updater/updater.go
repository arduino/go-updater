// This file is part of go-updater.
//
// Copyright 2025 ARDUINO SA (http://www.arduino.cc/)
//
// This software is released under the GNU General Public License version 3,
// which covers the main part of go-updater.
// The terms of this license can be found at:
// https://www.gnu.org/licenses/gpl-3.0.en.html
//
// You can be released from the requirements of the above licenses by purchasing
// a commercial license. Buying such a license is mandatory if you want to
// modify or otherwise use the software for commercial activities involving the
// Arduino software without disclosing the source code of your own applications.
// To purchase a commercial license, send an email to license@arduino.cc.

package updater

import (
	"fmt"
	"os"

	"github.com/arduino/go-updater/releaser"
)

// UpgradeConfirmCB is a function that is called when an update is ready to be applied.
type UpgradeConfirmCB func(current, target releaser.Version) bool

var DefaultUpgradeConfirmCb = func(current, target releaser.Version) bool { return true }

// CheckForUpdates checks for updates and applies it if available.
// If the upgradeCb is not nil, it will prompt the user for confirmation before applying the update.
// If an update is applied, it will restart the application with the new version.
// If no update is available, it will return nil.
// If an error occurs during the update process, it will return the error.
func CheckForUpdates(targetPath string, current releaser.Version, client *releaser.Client, upgradeCb UpgradeConfirmCB) error {
	restartPath, err := apply(targetPath, current, client, upgradeCb)
	if err != nil {
		return err
	}

	if restartPath == "" {
		return nil // No update available
	}

	if err := execApp(restartPath); err != nil {
		return fmt.Errorf("update applied, but failed to restart application: %w", err)
	}
	// TODO: allow to define custom "exit" function
	os.Exit(0)
	return nil
}
