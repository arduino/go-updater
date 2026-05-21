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
// If an update is applied, it starts the new version and returns nil — the caller is responsible
// for exiting the current process (e.g. os.Exit(0) or wails runtime.Quit()).
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

	// Pass our PID to the new process so it can wait for us to exit before
	// acquiring exclusive resources (files, ports, etc.).
	os.Setenv(oldPIDEnvVar, strconv.Itoa(os.Getpid()))
	if err := execApp(restartPath); err != nil {
		os.Unsetenv(oldPIDEnvVar)
		return fmt.Errorf("update applied, but failed to restart application: %w", err)
	}
	os.Unsetenv(oldPIDEnvVar)

	// The new version is running. The caller must now exit the current process.
	return nil
}
