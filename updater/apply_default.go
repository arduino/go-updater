//go:build !darwin

package updater

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"

	"github.com/arduino/go-updater/releaser"

	"github.com/arduino/go-paths-helper"
	"github.com/codeclysm/extract/v4"
)

func apply(targetPath string, current releaser.Version, client *releaser.Client, upgradeConfirmCb UpgradeConfirmCB) (string, error) {
	currentPath := paths.New(targetPath)
	currentDir := currentPath.Parent()

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
	download, err := client.FetchZip(manifest.Version, plat)
	if err != nil {
		return "", fmt.Errorf("could not fetch update: %w", err)
	}
	defer download.Close()

	// Download the zip
	tmpZip := currentDir.Join("update.zip")
	defer func() {
		if err := tmpZip.Remove(); err != nil {
			slog.Warn("Could not remove temp zip", "zip", tmpZip.String(), "error", err)
		}
	}()

	tmpZipFile, err := tmpZip.Create()
	if err != nil {
		return "", err
	}
	defer tmpZipFile.Close()

	sha := sha256.New()
	if _, err := io.Copy(io.MultiWriter(sha, tmpZipFile), download); err != nil {
		return "", err
	}
	tmpZipFile.Close()

	// Check the hash
	if s := sha.Sum(nil); !bytes.Equal(s, manifest.Sha256) {
		return "", fmt.Errorf("bad hash: %x (expected %x)", s, manifest.Sha256)
	}

	// Unzip the update
	newDir := currentDir.Join(currentPath.Base() + ".new")
	slog.Info("Unzipping update", "tmpDir", newDir)
	if err := newDir.MkdirAll(); err != nil {
		return "", fmt.Errorf("could not create tmp dir to unzip update: %w", err)
	}
	defer func() {
		if err := newDir.RemoveAll(); err != nil {
			slog.Warn("Could not remove temp dir", "dir", newDir.String(), "error", err)
		}
	}()

	tmpZipFile, err = tmpZip.Open()
	if err != nil {
		return "", fmt.Errorf("could not open archive for unzip: %w", err)
	}
	defer tmpZipFile.Close()

	if err := extract.Archive(context.Background(), tmpZipFile, newDir.String(), nil); err != nil {
		return "", fmt.Errorf("extracting archive: %w", err)
	}

	// Find the binary inside the unzipped folder
	binaryPath := ""
	err = filepath.Walk(newDir.String(), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && (info.Mode()&0111 != 0) {
			binaryPath = path
			return io.EOF // stop walking after finding the first executable
		}
		return nil
	})
	if err != nil && err != io.EOF {
		return "", fmt.Errorf("error walking for binary: %w", err)
	}
	if binaryPath == "" {
		return "", fmt.Errorf("no executable binary found in %s", newDir.String())
	}

	// Remove old path leftovers
	oldPath := currentPath.Parent().Join(currentPath.Base() + ".old")

	slog.Info("Deleting old leftovers", "old", oldPath)
	err = oldPath.Remove()
	if err != nil && !os.IsNotExist(err) {
		return "", fmt.Errorf("could not remove old leftovers: %w", err)
	}

	// Rename current app as .old
	slog.Info("Backup current", "from", currentPath, "to", oldPath)
	if err := currentPath.Rename(oldPath); err != nil {
		return "", fmt.Errorf("could not rename old folder: %w", err)
	}

	// Move the new executable in place of the current one
	slog.Info("Installing update", "from", binaryPath, "to", currentPath.String())
	if err := os.Rename(binaryPath, currentPath.String()); err != nil {
		// Try rollback changes
		err = oldPath.Rename(currentPath)
		if err != nil {
			slog.Error("Could not rollback changes after failed update", "error", err)
		}
		return "", fmt.Errorf("could not install app: %w", err)
	}

	// Cleanup
	slog.Info("Cleanup", "path", oldPath)
	err = oldPath.Remove()
	if err != nil {
		slog.Warn("Could not remove old app", "error", err)
		// WINDOWS: the folder cannot be removed, so we try to hide it
		err := hideFile(oldPath.String())
		if err != nil {
			slog.Warn("Could not hide old app", "error", err)
		}
	}

	slog.Info("Updated completed", "path", currentPath)

	return currentPath.String(), nil
}
