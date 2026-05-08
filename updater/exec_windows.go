// This file is part of go-updater.
//
// Copyright 2025 ARDUINO SA (http://www.arduino.cc/)
// SPDX-FileCopyrightText: Arduino s.r.l. and/or its affiliated companies
// SPDX-License-Identifier: GPL-3.0-or-later

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
