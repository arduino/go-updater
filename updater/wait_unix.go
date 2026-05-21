// This file is part of go-updater.
//
// SPDX-FileCopyrightText: Arduino s.r.l. and/or its affiliated companies
// SPDX-License-Identifier: GPL-3.0-or-later

//go:build !windows

package updater

import (
	"errors"
	"os"
	"syscall"
	"time"
)

// waitForProcess polls until the process with the given PID is no longer running.
func waitForProcess(pid int) {
	for {
		p, err := os.FindProcess(pid)
		if err != nil {
			return
		}
		err = p.Signal(syscall.Signal(0))
		if err == nil || errors.Is(err, syscall.EPERM) {
			time.Sleep(100 * time.Millisecond)
			continue
		}
		// ESRCH or os.ErrProcessDone → process is gone
		return
	}
}
