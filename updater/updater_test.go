// This file is part of go-updater.
//
// SPDX-FileCopyrightText: Arduino s.r.l. and/or its affiliated companies
// SPDX-License-Identifier: GPL-3.0-or-later

package updater

import (
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/arduino/go-updater/releaser"
)

func TestCheckForUpdates(t *testing.T) {
	t.Run("No Update", func(t *testing.T) {
		tmpExec := CreateTmpExecutable(t, "no-update-is-needed", []byte{})
		defer tmpExec.cleanup()
		client := CreateRelease(t, "2.0.0", []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06})

		err := CheckForUpdates(tmpExec.targetPath, releaser.Version("2.0.0"), client, DefaultUpgradeConfirmCb)
		require.NoError(t, err)
	})

	t.Run("status not found", func(t *testing.T) {
		tmp := CreateTmpExecutable(t, "no-update-is-needed", []byte{0xDE, 0xAD, 0xBE, 0xEF})
		defer tmp.cleanup()
		notFound := CreateReleaseWithHTTPErrorResponse(t, http.StatusNotFound)

		err := CheckForUpdates(tmp.targetPath, releaser.Version("1.0.0"), notFound, DefaultUpgradeConfirmCb)
		require.Error(t, err)
	})

	t.Run("status not found", func(t *testing.T) {
		tmp := CreateTmpExecutable(t, "a-not-found", []byte{0xAA})
		defer tmp.cleanup()

		notFound := CreateReleaseWithHTTPErrorResponse(t, http.StatusNotFound)

		err := CheckForUpdates(tmp.targetPath, releaser.Version("1.0.0"), notFound, DefaultUpgradeConfirmCb)
		require.Error(t, err)
	})

	t.Run("update ok but restart failed", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("cannot test it on windows because it opens installer executable")
		}
		tmp := CreateTmpExecutable(t, "a-bad-exec-format", []byte{0xDE, 0xAD, 0xBE, 0xEF})
		defer tmp.cleanup()
		client := CreateRelease(t, "2.0.0", []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06})

		err := CheckForUpdates(tmp.targetPath, releaser.Version("1.0.0"), client, DefaultUpgradeConfirmCb)
		require.Error(t, err)
		require.True(t, strings.Contains(err.Error(), "failed to restart application"), "Expected error about failed restart, got: %v", err)
	})
}

func TestWaitForOldApplication(t *testing.T) {
	t.Run("no-op when env var is not set", func(t *testing.T) {
		os.Unsetenv(oldPIDEnvVar)
		// Should return immediately without blocking.
		done := make(chan struct{})
		go func() { WaitForOldApplication(); close(done) }()
		select {
		case <-done:
		case <-time.After(500 * time.Millisecond):
			t.Fatal("WaitForOldApplication blocked unexpectedly")
		}
	})

	t.Run("no-op when env var is not a valid pid", func(t *testing.T) {
		t.Setenv(oldPIDEnvVar, "not-a-pid")
		done := make(chan struct{})
		go func() { WaitForOldApplication(); close(done) }()
		select {
		case <-done:
		case <-time.After(500 * time.Millisecond):
			t.Fatal("WaitForOldApplication blocked unexpectedly")
		}
	})

	t.Run("waits until old process exits", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("waitForProcess is a no-op on windows")
		}
		// Start a long-running process; we'll kill it ourselves to control timing.
		cmd := exec.Command("sleep", "60")
		require.NoError(t, cmd.Start())
		t.Setenv(oldPIDEnvVar, strconv.Itoa(cmd.Process.Pid))

		done := make(chan struct{})
		go func() { WaitForOldApplication(); close(done) }()

		// Should still be blocking after a short pause.
		select {
		case <-done:
			t.Fatal("WaitForOldApplication returned before process was killed")
		case <-time.After(200 * time.Millisecond):
		}

		// Kill and reap the child so the PID is freed and Signal(0) starts failing.
		require.NoError(t, cmd.Process.Kill())
		require.Error(t, cmd.Wait()) // killed process exits with error; reap to free the PID

		select {
		case <-done:
		case <-time.After(500 * time.Millisecond):
			t.Fatal("WaitForOldApplication did not unblock after process was killed")
		}
		require.Empty(t, os.Getenv(oldPIDEnvVar), "env var should be cleared after waiting")
	})

	t.Run("returns immediately if old process is already gone", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("waitForProcess is a no-op on windows")
		}
		cmd := exec.Command("sleep", "0")
		require.NoError(t, cmd.Start())
		require.NoError(t, cmd.Wait()) // reap it so the PID is freed before we call WaitForOldApplication
		t.Setenv(oldPIDEnvVar, strconv.Itoa(cmd.Process.Pid))

		done := make(chan struct{})
		go func() { WaitForOldApplication(); close(done) }()
		select {
		case <-done:
		case <-time.After(2 * time.Second):
			t.Fatal("WaitForOldApplication blocked on an already-dead process")
		}
	})
}
