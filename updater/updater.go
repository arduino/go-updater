// This file is part of go-updater.
//
// Copyright 2025 ARDUINO SA (http://www.arduino.cc/)
// SPDX-FileCopyrightText: Arduino s.r.l. and/or its affiliated companies
// SPDX-License-Identifier: GPL-3.0-or-later

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
	// TODO: allow to define custom "exit"
	os.Exit(0)
	return nil
}
