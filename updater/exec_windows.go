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
	"fmt"

	runas "github.com/arduino/go-windows-runas"
)

func execApp(path string) error {
	exitCode, err := runas.RunElevated(path, "", []string{}, false, false)
	if err != nil {
		return fmt.Errorf("could not run installer: %w", err)
	}
	if exitCode != 0 {
		return fmt.Errorf("installer exited with code %d", exitCode)
	}
	return nil
}
