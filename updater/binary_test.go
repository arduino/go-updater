// This file is part of go-updater.
//
// SPDX-FileCopyrightText: Arduino s.r.l. and/or its affiliated companies
// SPDX-License-Identifier: GPL-3.0-or-later

package updater

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/arduino/go-updater/releaser"
)

func TestUpdateCLI(t *testing.T) {
	newContent := []byte{0x01, 0x02, 0x03}

	tests := []struct {
		name     string
		fileName string
		build    func(t *testing.T, binName string) []byte
	}{
		{
			name:     "zip",
			fileName: "mycli-2.0.0.zip",
			build: func(t *testing.T, binName string) []byte {
				var buf bytes.Buffer
				zw := zip.NewWriter(&buf)
				w, err := zw.Create("mycli/" + binName)
				require.NoError(t, err)
				_, err = w.Write(newContent)
				require.NoError(t, err)
				require.NoError(t, zw.Close())
				return buf.Bytes()
			},
		},
		{
			name:     "tar.gz with an unrelated file name",
			fileName: "mycli-2.0.0.bin",
			build: func(t *testing.T, binName string) []byte {
				var buf bytes.Buffer
				gz := gzip.NewWriter(&buf)
				tw := tar.NewWriter(gz)
				require.NoError(t, tw.WriteHeader(&tar.Header{Name: "LICENSE.txt", Mode: 0644, Size: 3}))
				_, err := tw.Write([]byte("GPL"))
				require.NoError(t, err)
				require.NoError(t, tw.WriteHeader(&tar.Header{Name: binName, Mode: 0755, Size: int64(len(newContent))}))
				_, err = tw.Write(newContent)
				require.NoError(t, err)
				require.NoError(t, tw.Close())
				require.NoError(t, gz.Close())
				return buf.Bytes()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmp := CreateTmpExecutable(t, "update-binary", []byte{0xDE, 0xAD, 0xBE, 0xEF})
			defer tmp.cleanup()
			before, err := os.Stat(tmp.targetPath)
			require.NoError(t, err)

			release := tt.build(t, filepath.Base(tmp.targetPath))
			client := newBinaryRelease(t, "2.0.0", tt.fileName, release)

			updated, err := UpdateCLI(tmp.targetPath, releaser.Version("1.0.0"), client, DefaultUpgradeConfirmCb)
			require.NoError(t, err)
			require.True(t, updated)

			data, err := os.ReadFile(tmp.targetPath)
			require.NoError(t, err)
			require.Equal(t, newContent, data)

			after, err := os.Stat(tmp.targetPath)
			require.NoError(t, err)
			require.Equal(t, before.Mode().Perm(), after.Mode().Perm())

			entries, err := os.ReadDir(filepath.Dir(tmp.targetPath))
			require.NoError(t, err)
			require.Len(t, entries, 1)
		})
	}

	t.Run("no update", func(t *testing.T) {
		tmp := CreateTmpExecutable(t, "update-binary-none", []byte{0xDE, 0xAD, 0xBE, 0xEF})
		defer tmp.cleanup()
		client := newBinaryRelease(t, "1.0.0", "mycli-1.0.0.zip", tests[0].build(t, "mycli"))

		updated, err := UpdateCLI(tmp.targetPath, releaser.Version("1.0.0"), client, DefaultUpgradeConfirmCb)
		require.NoError(t, err)
		require.False(t, updated)
	})

	t.Run("archive without the executable", func(t *testing.T) {
		tmp := CreateTmpExecutable(t, "update-binary-missing", []byte{0xDE, 0xAD, 0xBE, 0xEF})
		defer tmp.cleanup()
		release := tests[0].build(t, "other-name")
		client := newBinaryRelease(t, "2.0.0", "mycli-2.0.0.zip", release)

		updated, err := UpdateCLI(tmp.targetPath, releaser.Version("1.0.0"), client, DefaultUpgradeConfirmCb)
		require.ErrorContains(t, err, "no file named")
		require.False(t, updated)

		data, err := os.ReadFile(tmp.targetPath)
		require.NoError(t, err)
		require.Equal(t, []byte{0xDE, 0xAD, 0xBE, 0xEF}, data)
	})
}

// newBinaryRelease returns a client that serves release as fileName for the running platform.
func newBinaryRelease(t *testing.T, version releaser.Version, fileName string, release []byte) *releaser.Client {
	t.Helper()

	inputPath := filepath.Join(t.TempDir(), fileName)
	require.NoError(t, os.WriteFile(inputPath, release, 0600))
	outputDir := t.TempDir()
	_, err := releaser.CreateRelease(inputPath, releaser.NewPlatform(runtime.GOOS, runtime.GOARCH), version, outputDir)
	require.NoError(t, err)

	return &releaser.Client{
		BaseURL: &url.URL{Scheme: "http", Host: "example.com"},
		CmdName: "testcmd",
		HTTPClient: &mockHTTPClient{doFunc: func(req *http.Request) (*http.Response, error) {
			var file string
			switch req.URL.Path {
			case "/testcmd/" + runtime.GOOS + "-" + runtime.GOARCH + ".json":
				file = runtime.GOOS + "-" + runtime.GOARCH + ".json"
			case "/testcmd/" + fileName:
				file = fileName
			default:
				panic("unreachable request: " + req.URL.Path)
			}
			body, err := os.ReadFile(filepath.Join(outputDir, file))
			require.NoError(t, err)
			return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewReader(body))}, nil
		}},
	}
}
