// This file is part of go-updater.
//
// SPDX-FileCopyrightText: Arduino s.r.l. and/or its affiliated companies
// SPDX-License-Identifier: GPL-3.0-or-later

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

// UpdateCLI replaces the executable at targetPath with the latest release, on every OS.
// It does not restart the application. It returns true if an update was applied.
func UpdateCLI(targetPath string, current releaser.Version, client *releaser.Client, upgradeCb UpgradeConfirmCB) (bool, error) {
	currentPath := paths.New(targetPath)
	currentDir := currentPath.Parent()

	plat := releaser.NewPlatform(runtime.GOOS, runtime.GOARCH)
	slog.Info("Checking for updates", "platform", plat)
	manifest, err := client.GetLatestVersion(plat)
	if err != nil {
		return false, err
	}
	if manifest.Version == current {
		return false, nil
	}

	if upgradeCb != nil && !upgradeCb(current, manifest.Version) {
		slog.Info("Update not confirmed by user, exiting without applying update")
		return false, nil
	}

	slog.Info("Downloading update", "version", manifest.Version, "platform", plat)
	download, err := client.FetchRelease(manifest)
	if err != nil {
		return false, fmt.Errorf("could not fetch update: %w", err)
	}
	defer download.Close()

	// The temp dir is next to the executable.
	tmpDir := currentDir.Join("." + currentPath.Base() + ".new")
	if err := tmpDir.MkdirAll(); err != nil {
		return false, fmt.Errorf("could not create temp dir: %w", err)
	}
	defer func() {
		if err := tmpDir.RemoveAll(); err != nil {
			slog.Warn("Could not remove temp dir", "dir", tmpDir, "error", err)
		}
	}()

	tmpRelease := tmpDir.Join(download.FileName)
	f, err := tmpRelease.Create()
	if err != nil {
		return false, err
	}
	defer f.Close()
	sha := sha256.New()
	if _, err := io.Copy(io.MultiWriter(sha, f), download); err != nil {
		return false, err
	}
	f.Close()
	if s := sha.Sum(nil); !bytes.Equal(s, manifest.Sha256) {
		return false, fmt.Errorf("bad hash: %x (expected %x)", s, manifest.Sha256)
	}

	slog.Info("Extracting archive", "path", tmpRelease)
	extractDir := tmpDir.Join("extracted")
	archive, err := tmpRelease.Open()
	if err != nil {
		return false, fmt.Errorf("could not open archive: %w", err)
	}
	defer archive.Close()
	if err := extract.Archive(context.Background(), archive, extractDir.String(), nil); err != nil {
		return false, fmt.Errorf("extracting archive: %w", err)
	}
	var newPath *paths.Path
	err = filepath.WalkDir(extractDir.String(), func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && d.Name() == currentPath.Base() {
			newPath = paths.New(path)
			return fs.SkipAll
		}
		return nil
	})
	if err != nil {
		return false, fmt.Errorf("could not search the archive: %w", err)
	}
	if newPath == nil {
		return false, fmt.Errorf("no file named %s in the release archive", currentPath.Base())
	}

	currentStat, err := currentPath.Stat()
	if err != nil {
		return false, fmt.Errorf("could not stat current executable: %w", err)
	}
	if err := newPath.Chmod(currentStat.Mode().Perm()); err != nil {
		return false, fmt.Errorf("could not set mode of new executable: %w", err)
	}

	// Remove the leftover of a previous update.
	oldPath := currentDir.Join(currentPath.Base() + ".old")
	if err := oldPath.Remove(); err != nil && !os.IsNotExist(err) {
		return false, fmt.Errorf("could not remove old leftovers: %w", err)
	}

	slog.Info("Backup current", "from", currentPath, "to", oldPath)
	if err := currentPath.Rename(oldPath); err != nil {
		return false, fmt.Errorf("could not rename current executable: %w", err)
	}
	slog.Info("Installing update", "from", newPath, "to", currentPath)
	if err := newPath.Rename(currentPath); err != nil {
		if err := oldPath.Rename(currentPath); err != nil {
			slog.Error("Could not rollback changes after failed update", "error", err)
		}
		return false, fmt.Errorf("could not install update: %w", err)
	}

	if err := oldPath.Remove(); err != nil {
		slog.Warn("Could not remove old executable", "error", err)
		if err := hideFile(oldPath.String()); err != nil {
			slog.Warn("Could not hide old executable", "error", err)
		}
	}

	slog.Info("Update completed", "path", currentPath, "version", manifest.Version)
	return true, nil
}
