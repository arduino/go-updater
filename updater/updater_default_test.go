package updater

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/arduino/go-updater/releaser"
)

func TestPerformUpdate(t *testing.T) {
	tmpExec := CreateTmpExecutable(t, "successfulUpdate", []byte{0xDE, 0xAD, 0xBE, 0xEF})
	defer tmpExec.cleanup()
	client := CreateRelease(t, "2.0.0", []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06})

	restartPath, err := checkForUpdates(tmpExec.targetPath, Version("1.0.0"), client)
	require.NoError(t, err)
	require.NotEmpty(t, restartPath)

	data, err := os.ReadFile(restartPath)
	require.NoError(t, err)
	require.Equal(t, []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06}, data, "Updated binary content does not match expected content")
}

func TestNoUpdateRequired(t *testing.T) {
	tmpExec := CreateTmpExecutable(t, "noUpdate", []byte{0xDE, 0xAD, 0xBE, 0xEF})
	defer tmpExec.cleanup()
	client := CreateRelease(t, "1.0.0", []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06})

	result, err := checkForUpdates(tmpExec.targetPath, Version("1.0.0"), client)
	require.NoError(t, err)
	require.Equal(t, "", result)
}

func TestCleanUpOldFiles(t *testing.T) {
	tmpExec := CreateTmpExecutable(t, "cleanUp", []byte{0xDE, 0xAD, 0xBE, 0xEF})
	client := CreateRelease(t, "3.0.0", []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06})

	result, err := checkForUpdates(tmpExec.targetPath, Version("1.0.0"), client)
	require.NoError(t, err)
	require.Equal(t, tmpExec.targetPath, result)

	// Check that the target directory contains only the updated file
	targetDir := filepath.Dir(tmpExec.targetPath)
	entries, err := os.ReadDir(targetDir)
	require.NoError(t, err)
	require.Len(t, entries, 1)
	require.Equal(t, filepath.Base(tmpExec.targetPath), entries[0].Name())
	require.False(t, entries[0].IsDir())

	tmpExec.cleanup()
}

type TmpExecutable struct {
	targetPath string // Path to the executable to be replaced
	content    []byte // Content of executable file

	cleanup func() // Cleanup function to clean up files and dirs
}

func CreateTmpExecutable(t *testing.T, binaryName string, content []byte) TmpExecutable {
	tmpDir := filepath.Join(".", "test-"+binaryName)
	err := os.MkdirAll(tmpDir, 0755)
	require.NoError(t, err)

	binaryPath := filepath.Join(tmpDir, binaryName)
	require.NoError(t, os.WriteFile(binaryPath, content, 0600))

	return TmpExecutable{
		targetPath: binaryPath,
		content:    content,
		cleanup: func() {
			// Remove the temporary directory and its contents
			if err := os.RemoveAll(tmpDir); err != nil {
				t.Logf("Failed to clean up temporary directory %s: %v", tmpDir, err)
			}
		},
	}
}

func CreateRelease(t *testing.T, version Version, content []byte) *releaser.Client {
	tmpDir := t.TempDir()

	inputDir := filepath.Join(tmpDir, "input")

	require.NoError(t, os.Mkdir(inputDir, 0700))
	fileA := filepath.Join(inputDir, "new-bin")
	require.NoError(t, os.WriteFile(fileA, content, 0600))

	outputDir := filepath.Join(tmpDir, "output")

	_, err := releaser.CreateRelease(inputDir, releaser.NewPlatform("linux", "amd64"), version.String(), outputDir)
	require.NoError(t, err)

	// check zip file exists and json manifest is created
	zipPath := filepath.Join(outputDir, version.String(), "linux-amd64.zip")
	_, err = os.Stat(zipPath)
	require.NoError(t, err, "zip file does not exist")

	require.NoError(t, err)
	jsonPath := filepath.Join(outputDir, "linux-amd64.json")
	_, err = os.Stat(jsonPath)
	require.NoError(t, err, "manifest JSON file does not exist")

	rawManifest, err := os.ReadFile(jsonPath)
	require.NoError(t, err)
	manifestResp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewReader(rawManifest)),
	}

	zipBytes, err := os.ReadFile(zipPath)
	require.NoError(t, err)
	zipResp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewReader(zipBytes)),
	}

	client := &releaser.Client{
		BaseURL: &url.URL{Scheme: "http", Host: "example.com"},
		CmdName: "testcmd",
		HTTPClient: &mockHTTPClient{doFunc: func(req *http.Request) (*http.Response, error) {
			if req.URL.Path == "/testcmd/linux-amd64.json" && req.Method == http.MethodGet {
				return manifestResp, nil
			}
			if req.URL.Path == fmt.Sprintf("/testcmd/%s/linux-amd64.zip", version) && req.Method == http.MethodGet {
				return zipResp, nil
			}
			panic("unreachable request")
		}},
	}

	return client
}

// mockHTTPClient implements releaser.HTTPDoer for testing.
type mockHTTPClient struct {
	doFunc func(req *http.Request) (*http.Response, error)
}

func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return m.doFunc(req)
}
