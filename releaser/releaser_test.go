package releaser

import (
	"archive/zip"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCreateReleaseFile(t *testing.T) {
	tmpDir := t.TempDir()

	outputDir := filepath.Join(tmpDir, "output")
	inputDir := filepath.Join(tmpDir, "input")
	require.NoError(t, os.Mkdir(inputDir, 0700))

	dummyFile := filepath.Join(inputDir, "dummy.txt")
	content := []byte("hello world")
	require.NoError(t, os.WriteFile(dummyFile, content, 0600))

	version := "1.2.3"
	manifest, err := CreateRelease(dummyFile, NewPlatform("linux", "amd64"), version, outputDir)
	require.NoError(t, err)

	// Check manifest fields
	require.Equal(t, version, manifest.Version)
	require.Len(t, manifest.Sha256, 32)

	// Check that the zip file exists and contains the file
	// Expected folder structure in outputDir:
	// outputDir/
	//  1.2.3/
	//      linux-amd64.zip
	//   linux-amd64.json
	//
	zipPath := filepath.Join(outputDir, version, "linux-amd64.zip")
	zf, err := zip.OpenReader(zipPath)
	require.NoError(t, err, "could not open zip file %s", zipPath)
	defer zf.Close()

	found := false
	for _, f := range zf.File {
		if f.Name == "dummy.txt" {
			found = true
			rc, err := f.Open()
			if err != nil {
				t.Fatalf("could not open file in zip: %v", err)
			}
			defer rc.Close()
			data := make([]byte, len(content))
			_, err = rc.Read(data)
			require.NoError(t, err, "could not read file from zip")
			if string(data) != string(content) {
				t.Errorf("zip content mismatch: got %q, want %q", string(data), string(content))
			}
		}
	}
	if !found {
		t.Errorf("dummy.txt not found in zip")
	}

	// Check that the manifest JSON file exists and is valid
	jsonPath := filepath.Join(outputDir, "linux-amd64.json")
	b, err := os.ReadFile(jsonPath)
	require.NoError(t, err, "could not read manifest json file %s", jsonPath)

	// Unmarshal the manifest JSON
	var m Manifest
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatalf("could not unmarshal manifest json: %v", err)
	}
	require.Equal(t, version, m.Version)
	require.Len(t, m.Sha256, 32)
}

func TestCreateReleaseFolder(t *testing.T) {
	tmpDir := t.TempDir()

	outputDir := filepath.Join(tmpDir, "output")
	inputDir := filepath.Join(tmpDir, "input")
	require.NoError(t, os.Mkdir(inputDir, 0700))

	// Create multiple files and a subdirectory
	// Folder structure:
	// input/
	//   a.txt
	//   subdir/
	//     b.txt
	subDir := filepath.Join(inputDir, "subdir")
	require.NoError(t, os.Mkdir(subDir, 0700))

	fileA := filepath.Join(inputDir, "a.txt")
	fileB := filepath.Join(subDir, "b.txt")
	contentA := []byte("file A")
	contentB := []byte("file B")
	require.NoError(t, os.WriteFile(fileA, contentA, 0600))
	require.NoError(t, os.WriteFile(fileB, contentB, 0600))

	version := "2.0.0"
	manifest, err := CreateRelease(inputDir, NewPlatform("linux", "amd64"), version, outputDir)
	require.NoError(t, err)

	// Check manifest fields
	require.Equal(t, version, manifest.Version)
	require.Len(t, manifest.Sha256, 32)

	// Check that the zip file exists and contains all files
	zipPath := filepath.Join(outputDir, version, "linux-amd64.zip")
	zf, err := zip.OpenReader(zipPath)
	require.NoError(t, err, "could not open zip file %s", zipPath)
	defer zf.Close()

	foundA := false
	foundB := false
	for _, f := range zf.File {
		if f.Name == "input/a.txt" {
			foundA = true
			rc, err := f.Open()
			require.NoError(t, err)
			data, err := io.ReadAll(rc)
			require.NoError(t, err)
			require.Equal(t, contentA, data)
			rc.Close()
		}
		if f.Name == "input/subdir/b.txt" {
			foundB = true
			rc, err := f.Open()
			require.NoError(t, err)
			data, err := io.ReadAll(rc)
			require.NoError(t, err)
			require.Equal(t, contentB, data)
			rc.Close()
		}
	}
	require.True(t, foundA, "a.txt not found in zip")
	require.True(t, foundB, "subdir/b.txt not found in zip")

	// Check that the manifest JSON file exists and is valid
	jsonPath := filepath.Join(outputDir, "linux-amd64.json")
	b, err := os.ReadFile(jsonPath)
	require.NoError(t, err, "could not read manifest json file %s", jsonPath)

	var m Manifest
	require.NoError(t, json.Unmarshal(b, &m))
	require.Equal(t, version, m.Version)
	require.Len(t, m.Sha256, 32)
}
