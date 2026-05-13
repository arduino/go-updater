// This file is part of go-updater.
//
// SPDX-FileCopyrightText: Arduino s.r.l. and/or its affiliated companies
// SPDX-License-Identifier: GPL-3.0-or-later

package updater

import (
	"net/http"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/arduino/go-updater/releaser"
)

func TestCheckForUpdates(t *testing.T) {
	t.Run("No Update", func(t *testing.T) {
		tmpExec := CreateTmpExecutable(t, "no-update-is-needed", []byte{})
		defer tmpExec.cleanup()
		client := CreateRelease(t, "2.0.0", []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06})

		err := CheckForUpdates(tmpExec.targetPath, releaser.Version("2.0.0"), client, DefaultUpgradeConfirmCb)
		require.NoError(t, err)
	})

	t.Run("status not found", func(t *testing.T) {
		tmp := CreateTmpExecutable(t, "no-update-is-needed", []byte{0xDE, 0xAD, 0xBE, 0xEF})
		defer tmp.cleanup()
		notFound := CreateReleaseWithHTTPErrorResponse(t, http.StatusNotFound)

		err := CheckForUpdates(tmp.targetPath, releaser.Version("1.0.0"), notFound, DefaultUpgradeConfirmCb)
		require.Error(t, err)
	})

	t.Run("status not found", func(t *testing.T) {
		tmp := CreateTmpExecutable(t, "a-not-found", []byte{0xAA})
		defer tmp.cleanup()

		notFound := CreateReleaseWithHTTPErrorResponse(t, http.StatusNotFound)

		err := CheckForUpdates(tmp.targetPath, releaser.Version("1.0.0"), notFound, DefaultUpgradeConfirmCb)
		require.Error(t, err)
	})

	t.Run("update ok but restart failed", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("cannot test it on windows because it opens installer executable")
		}
		tmp := CreateTmpExecutable(t, "a-bad-exec-format", []byte{0xDE, 0xAD, 0xBE, 0xEF})
		defer tmp.cleanup()
		client := CreateRelease(t, "2.0.0", []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06})

		err := CheckForUpdates(tmp.targetPath, releaser.Version("1.0.0"), client, DefaultUpgradeConfirmCb)
		require.Error(t, err)
		require.True(t, strings.Contains(err.Error(), "failed to restart application"), "Expected error about failed restart, got: %v", err)
	})
}
