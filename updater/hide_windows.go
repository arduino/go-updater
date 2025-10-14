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
