// This file is part of go-updater.
//
// Copyright 2025 ARDUINO SA (http://www.arduino.cc/)
// SPDX-FileCopyrightText: Arduino s.r.l. and/or its affiliated companies
// SPDX-License-Identifier: GPL-3.0-or-later

//go:build !darwin && !windows

package updater

import "os/exec"

// default execApp from golang
func execApp(path string, args ...string) error {
	cmd := exec.Command(path, args...)
	return cmd.Start()
}
