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

//go:build windows

package updater

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"
	"log/slog"
	"os"
	"runtime"

	"github.com/arduino/go-updater/releaser"
)

// NOTE: On windows, the 'targetpath' is ignored (see the first argument) because is not replaced with the new version but
// the update is based on an installer executable. So, instead of replacing the binary in place, the installer is downloaded and run (see the exec_windows.go)
func apply(_ string, current releaser.Version, client *releaser.Client, upgradeConfirmCb UpgradeConfirmCB) (string, error) {
	plat := releaser.NewPlatform(runtime.GOOS, runtime.GOARCH)
	slog.Info("Checking for updates", "platform", plat)
	manifest, err := client.GetLatestVersion(plat)
	if err != nil {
		return "", err
	}
	if manifest.Version == current {
		// No updates available, bye bye
		return "", nil
	}

	if upgradeConfirmCb != nil && !upgradeConfirmCb(current, manifest.Version) {
		slog.Info("Update not confirmed by user, exiting without applying update")
		return "", nil
	}

	// Download the update
	slog.Info("Downloading update", "version", manifest.Version, "platform", plat)
	download, err := client.FetchRelease(manifest)
	if err != nil {
		return "", fmt.Errorf("could not fetch update: %w", err)
	}
	defer download.Close()

	tmpRelease, err := os.CreateTemp("", "update-*.exe")
	if err != nil {
		return "", fmt.Errorf("could not create temp file: %w", err)
	}

	sha := sha256.New()
	if _, err := io.Copy(io.MultiWriter(sha, tmpRelease), download); err != nil {
		return "", err
	}
	tmpRelease.Close()

	// Check the hash
	if s := sha.Sum(nil); !bytes.Equal(s, manifest.Sha256) {
		return "", fmt.Errorf("bad hash: %x (expected %x)", s, manifest.Sha256)
	}

	return tmpRelease.Name(), nil
}
