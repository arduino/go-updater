// This file is part of go-updater.
//
// SPDX-FileCopyrightText: Arduino s.r.l. and/or its affiliated companies
// SPDX-License-Identifier: GPL-3.0-or-later

package updater

import (
	"fmt"
	"os"
	"strconv"

	"github.com/arduino/go-updater/releaser"
)

// UpgradeConfirmCB is a function that is called when an update is ready to be applied.
type UpgradeConfirmCB func(current, target releaser.Version) bool

var DefaultUpgradeConfirmCb = func(current, target releaser.Version) bool { return true }

// CheckForUpdates checks for updates and applies it if available.
// If the upgradeCb is not nil, it will prompt the user for confirmation before applying the update.
// Returns (true, nil) if an update was applied and the new version is running — the caller must
// then exit the current process (e.g. os.Exit(0) or wails runtime.Quit()).
// Returns (false, nil) if no update is available.
// Returns (false, err) if an error occurs during the update process.
func CheckForUpdates(targetPath string, current releaser.Version, client *releaser.Client, upgradeCb UpgradeConfirmCB) (bool, error) {
	restartPath, err := apply(targetPath, current, client, upgradeCb)
	if err != nil {
		return false, err
	}

	if restartPath == "" {
		return false, nil // No update available
	}

	// Pass the current PID to the new process so it can wait for it to exit before launching the new version.
	os.Setenv(oldPIDEnvVar, strconv.Itoa(os.Getpid()))
	if err := execApp(restartPath); err != nil {
		os.Unsetenv(oldPIDEnvVar)
		return false, fmt.Errorf("update applied, but failed to restart application: %w", err)
	}
	os.Unsetenv(oldPIDEnvVar)

	return true, nil
}
