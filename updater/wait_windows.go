// This file is part of go-updater.
//
// SPDX-FileCopyrightText: Arduino s.r.l. and/or its affiliated companies
// SPDX-License-Identifier: GPL-3.0-or-later

//go:build windows

package updater

// waitForProcess is a no-op on Windows: updates are handled by the installer
// via runas.RunElevated, which runs synchronously and manages the process
// transition itself.
func waitForProcess(pid int) {}
