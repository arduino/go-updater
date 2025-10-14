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
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"runtime"

	"github.com/arduino/go-updater/releaser"

	"github.com/arduino/go-paths-helper"
	"github.com/codeclysm/extract/v4"
)

func apply(targetPath string, current releaser.Version, client *releaser.Client, upgradeConfirmCb UpgradeConfirmCB) (string, error) {
	currentAppPath := paths.New(targetPath).Parent().Parent().Parent()
	if currentAppPath.Ext() != ".app" {
		return "", fmt.Errorf("could not find app root in %s", targetPath)
	}
	currentFolderAppName := currentAppPath.Base()
	oldAppPath := currentAppPath.Parent().Join(currentFolderAppName + ".old.app")
	if oldAppPath.Exist() {
		return "", fmt.Errorf("temp app already exists: %s, cannot update", oldAppPath)
	}

	// Fetch information about updates
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

	// Download the update.
	slog.Info("Downloading update", "version", manifest.Version, "platform", plat)
	download, err := client.FetchRelease(manifest)
	if err != nil {
		return "", err
	}
	defer download.Close()

	tmp := paths.TempDir().Join(currentFolderAppName)
	if err := tmp.MkdirAll(); err != nil {
		return "", err
	}
	tmpRelease := tmp.Join(download.FileName)
	tmpAppPath := tmp.Join(fmt.Sprintf(".%s.new", currentFolderAppName))

	defer func() {
		_ = tmp.RemoveAll()
	}()

	f, err := tmpRelease.Create()
	if err != nil {
		return "", err
	}
	defer f.Close()

	sha := sha256.New()
	if _, err := io.Copy(io.MultiWriter(sha, f), download); err != nil {
		return "", err
	}
	f.Close()

	// Check the hash
	if s := sha.Sum(nil); !bytes.Equal(s, manifest.Sha256) {
		return "", fmt.Errorf("bad hash: %x (expected %x)", s, manifest.Sha256)
	}

	if tmpRelease.Ext() != ".zip" {
		// TODO: add .dmg support
		return "", fmt.Errorf("expected a .zip release file, got %s", tmpRelease)
	}

	// Unzip the update
	slog.Info("Unzipping update", "tmpDir", tmpAppPath)
	if err := tmpAppPath.MkdirAll(); err != nil {
		return "", fmt.Errorf("could not create tmp dir to unzip update: %w", err)
	}

	f, err = tmpRelease.Open()
	if err != nil {
		return "", fmt.Errorf("could not open archive for unzip: %w", err)
	}
	defer f.Close()
	if err := extract.Archive(context.Background(), f, tmpAppPath.String(), nil); err != nil {
		return "", fmt.Errorf("extracting archive: %w", err)
	}

	// search the .app in the unzipped directory
	apps, err := fs.Glob(os.DirFS(tmpAppPath.String()), "*.app")
	if err != nil || len(apps) != 1 {
		return "", fmt.Errorf("could not find .app in tmp dir %q: %w", tmpAppPath, err)
	}
	tmpAppPath = tmpAppPath.Join(apps[0])

	slog.Info("Renaming old app", "from", currentAppPath, "to", oldAppPath)
	if err := currentAppPath.Rename(oldAppPath); err != nil {
		return "", fmt.Errorf("could not rename old app as .old: %w", err)
	}

	// Install new app
	slog.Info("Copying updated app", "from", tmpAppPath, "to", currentAppPath)
	if err := tmpAppPath.CopyDirTo(currentAppPath); err != nil {
		// Try rollback changes
		_ = currentAppPath.RemoveAll()
		_ = oldAppPath.Rename(currentAppPath)
		return "", fmt.Errorf("could not install app: %w", err)
	}

	// Remove old app
	slog.Info("Removing old app", "to", oldAppPath)
	_ = oldAppPath.RemoveAll()

	slog.Info("Returning updated app", "path", currentAppPath)

	return currentAppPath.String(), nil
}
