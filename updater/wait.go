// This file is part of go-updater.
//
// SPDX-FileCopyrightText: Arduino s.r.l. and/or its affiliated companies
// SPDX-License-Identifier: GPL-3.0-or-later

package updater

import (
	"os"
	"strconv"
)

const oldPIDEnvVar = "GO_UPDATER_OLD_PID"

// WaitForOldApplication blocks until the process that launched this update exits.
// It reads the GO_UPDATER_OLD_PID environment variable set by CheckForUpdates in the
// old process, and waits for that PID to terminate before returning.
//
// Call this once at the very beginning of main(), before opening files, ports,
// or any other exclusive resources, so the new version only proceeds once the
// old one has fully released them.
//
// If the app was not started by go-updater, this is a no-op.
func WaitForOldApplication() {
	pidStr := os.Getenv(oldPIDEnvVar)
	if pidStr == "" {
		return
	}
	pid, err := strconv.Atoi(pidStr)
	if err != nil {
		return
	}
	waitForProcess(pid)
	os.Unsetenv(oldPIDEnvVar)
}
