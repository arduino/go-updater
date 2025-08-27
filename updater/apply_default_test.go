package updater

import (
	"archive/zip"
	"bytes"
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

func TestApply(t *testing.T) {
	tmpExec := CreateTmpExecutable(t, "successfulUpdate", []byte{0xDE, 0xAD, 0xBE, 0xEF})
	defer tmpExec.cleanup()
	client := CreateRelease(t, "2.0.0", []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06})

	restartPath, err := apply(tmpExec.targetPath, releaser.Version("1.0.0"), client, DefaultUpgradeConfirmCb)
	require.NoError(t, err)
	require.NotEmpty(t, restartPath)

	data, err := os.ReadFile(restartPath)
	require.NoError(t, err)
	require.Equal(t, []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06}, data, "Updated binary content does not match expected content")

	// Check that the target directory contains only the updated file
	targetDir := filepath.Dir(tmpExec.targetPath)
	entries, err := os.ReadDir(targetDir)
	require.NoError(t, err)
	require.Len(t, entries, 1)
	require.Equal(t, filepath.Base(tmpExec.targetPath), entries[0].Name())
	require.False(t, entries[0].IsDir())
}

func TestApplyWithNoUpdate(t *testing.T) {
	tmpExec := CreateTmpExecutable(t, "noUpdate", []byte{0xDE, 0xAD, 0xBE, 0xEF})
	defer tmpExec.cleanup()
	client := CreateRelease(t, "1.0.0", []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06})

	result, err := apply(tmpExec.targetPath, releaser.Version("1.0.0"), client, DefaultUpgradeConfirmCb)
	require.NoError(t, err)
	require.Equal(t, "", result)

	// Check that the target directory contains only the updated file
	targetDir := filepath.Dir(tmpExec.targetPath)
	entries, err := os.ReadDir(targetDir)
	require.NoError(t, err)
	require.Len(t, entries, 1)
	require.Equal(t, filepath.Base(tmpExec.targetPath), entries[0].Name())
	require.False(t, entries[0].IsDir())
}

type TmpExecutable struct {
	targetPath string // Path to the executable to be replaced
	content    []byte // Content of executable file

	cleanup func() // Cleanup function to clean up files and dirs
}

func CreateTmpExecutable(t *testing.T, binaryName string, content []byte) TmpExecutable {
	t.Helper()

	// prefix the binary name with "test-" to put the folders in the .gitignore
	tmpDir := filepath.Join(".", "test-"+binaryName)
	err := os.MkdirAll(tmpDir, 0755)
	require.NoError(t, err)

	if runtime.GOOS == "windows" {
		binaryName += ".exe"
	}
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

func CreateRelease(t *testing.T, version releaser.Version, content []byte) *releaser.Client {
	t.Helper()

	tmpDir := t.TempDir()

	inputDir := filepath.Join(tmpDir, "input")
	require.NoError(t, os.Mkdir(inputDir, 0700))

	binName := "new-bin-"
	if runtime.GOOS == "windows" {
		binName += version.String() + ".exe"
		binPath := filepath.Join(inputDir, binName)
		require.NoError(t, os.WriteFile(binPath, content, 0600))
	} else {
		zipName := binName + version.String() + ".zip"
		zipFilePath := filepath.Join(inputDir, zipName)
		zipFile, err := os.Create(zipFilePath)
		require.NoError(t, err)
		zipW := zip.NewWriter(zipFile)
		header := &zip.FileHeader{
			Name:   binName,
			Method: zip.Deflate,
		}
		header.SetMode(0700)
		// Create the file in the ZIP archive
		binF, err := zipW.CreateHeader(header)
		require.NoError(t, err)
		_, err = binF.Write(content)
		require.NoError(t, err)
		require.NoError(t, zipW.Close())
		require.NoError(t, zipFile.Close())
		binName = zipName
	}

	binPath := filepath.Join(inputDir, binName)
	outputDir := filepath.Join(tmpDir, "output")

	_, err := releaser.CreateRelease(binPath, releaser.NewPlatform(runtime.GOOS, runtime.GOARCH), version, outputDir)
	require.NoError(t, err)

	jsonPath := filepath.Join(outputDir, runtime.GOOS+"-"+runtime.GOARCH+".json")
	rawManifest, err := os.ReadFile(jsonPath)
	require.NoError(t, err)
	manifestResp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewReader(rawManifest)),
	}

	outBinPath := filepath.Join(outputDir, binName)
	outBinBytes, err := os.ReadFile(outBinPath)
	require.NoError(t, err)
	// require.Equal(t, content, outBinBytes)
	outBinResp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewReader(outBinBytes)),
	}

	client := &releaser.Client{
		BaseURL: &url.URL{Scheme: "http", Host: "example.com"},
		CmdName: "testcmd",
		HTTPClient: &mockHTTPClient{doFunc: func(req *http.Request) (*http.Response, error) {
			if req.URL.Path == "/testcmd/"+runtime.GOOS+"-"+runtime.GOARCH+".json" && req.Method == http.MethodGet {
				return manifestResp, nil
			}
			if req.URL.Path == "/testcmd/"+binName && req.Method == http.MethodGet {
				return outBinResp, nil
			}
			panic("unreachable request")
		}},
	}

	return client
}

func CreateReleaseWithHTTPErrorResponse(t *testing.T, statusCode int) *releaser.Client {
	t.Helper()
	return &releaser.Client{
		BaseURL: &url.URL{Scheme: "http", Host: "example.com"},
		CmdName: "testcmd",
		HTTPClient: &mockHTTPClient{doFunc: func(req *http.Request) (*http.Response, error) {
			resp := &http.Response{
				StatusCode: statusCode,
				Body:       io.NopCloser(bytes.NewBufferString("")),
			}
			return resp, nil
		}},
	}
}

// mockHTTPClient implements releaser.HTTPDoer for testing.
type mockHTTPClient struct {
	doFunc func(req *http.Request) (*http.Response, error)
}

func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return m.doFunc(req)
}
