// This file is part of go-updater.
//
// Copyright 2025 ARDUINO SA (http://www.arduino.cc/)
// SPDX-FileCopyrightText: Arduino s.r.l. and/or its affiliated companies
// SPDX-License-Identifier: GPL-3.0-or-later

package updater

import (
	"fmt"
	"syscall"

	"log/slog"
)

func hideFile(filePath string) error {
	// Windows specific: Set the hidden attribute
	// FILE_ATTRIBUTE_HIDDEN (0x02)
	// See: https://learn.microsoft.com/en-us/windows/win32/fileio/file-attribute-constants
	ptr, err := syscall.UTF16PtrFromString(filePath)
	if err != nil {
		return fmt.Errorf("failed to convert path to UTF16 for Windows: %w", err)
	}
	err = syscall.SetFileAttributes(ptr, syscall.FILE_ATTRIBUTE_HIDDEN)
	if err != nil {
		return fmt.Errorf("failed to set hidden attribute on Windows for '%s': %w", filePath, err)
	}
	slog.Info("Hide file", "path", filePath)
	return nil
}
