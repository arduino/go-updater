//go:build !darwin

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
	download, err := client.FetchRelease(manifest)
	if err != nil {
		return "", fmt.Errorf("could not fetch update: %w", err)
	}
	defer download.Close()

	// Download the release
	tmpRelease := currentDir.Join(download.FileName)
	defer func() {
		if err := tmpRelease.Remove(); err != nil {
			slog.Warn("Could not remove temp file", "file", tmpRelease.String(), "error", err)
		}
	}()

	tmpFile, err := tmpRelease.Create()
	if err != nil {
		return "", err
	}
	defer tmpFile.Close()

	sha := sha256.New()
	if _, err := io.Copy(io.MultiWriter(sha, tmpFile), download); err != nil {
		return "", err
	}
	tmpFile.Close()

	// Check the hash
	if s := sha.Sum(nil); !bytes.Equal(s, manifest.Sha256) {
		return "", fmt.Errorf("bad hash: %x (expected %x)", s, manifest.Sha256)
	}

	tmpFile, err = tmpRelease.Open()
	if err != nil {
		return "", fmt.Errorf("could not open release file: %w", err)
	}
	defer tmpFile.Close()

	var binaryPath string
	switch tmpRelease.Ext() {
	case ".zip", ".gz", ".tgz":
		slog.Info("Extracting archive", "path", tmpRelease.String())
		newDir := currentDir.Join(currentPath.Base() + ".new")
		if err := newDir.MkdirAll(); err != nil {
			return "", fmt.Errorf("could not create tmp dir to extract archive: %w", err)
		}
		defer func() {
			if err := newDir.RemoveAll(); err != nil {
				slog.Warn("Could not remove temp dir", "dir", newDir.String(), "error", err)
			}
		}()
		if err := extract.Archive(context.Background(), tmpFile, newDir.String(), nil); err != nil {
			return "", fmt.Errorf("extracting archive: %w", err)
		}
		err = filepath.Walk(newDir.String(), func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() && (isExecutable(info)) {
				binaryPath = path
				return io.EOF // stop walking after finding the first executable
			}
			return nil
		})
		if err != nil && err != io.EOF {
			return "", fmt.Errorf("error walking for binary: %w", err)
		}
	default:
		stat, err := tmpFile.Stat()
		if err != nil {
			return "", fmt.Errorf("could not stat release file: %w", err)
		}
		if !isExecutable(stat) {
			return "", fmt.Errorf("expected an executable binary, got %s", tmpRelease.String())
		}
		binaryPath = tmpRelease.String()
	}
	_ = tmpFile.Close()

	if binaryPath == "" {
		return "", fmt.Errorf("no executable binary found in release")
	}

	// Remove old path leftovers
	oldPath := currentDir.Join(currentPath.Base() + ".old")

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
		if err := oldPath.Rename(currentPath); err != nil {
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
		if err := hideFile(oldPath.String()); err != nil {
			slog.Warn("Could not hide old app", "error", err)
		}
	}

	slog.Info("Updated completed", "path", currentPath)

	return currentPath.String(), nil
}

// isExecutable checks if a file is executable.
func isExecutable(info fs.FileInfo) bool {
	if runtime.GOOS == "windows" {
		return filepath.Ext(info.Name()) == ".exe"
	} else {
		return info.Mode()&0111 != 0
	}
}
