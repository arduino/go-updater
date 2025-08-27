package releaser

import (
	"encoding/json"
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

	version := Version("1.2.3")
	dummyFile := filepath.Join(inputDir, "dummy-1.2.3.txt")
	content := []byte("hello world")
	require.NoError(t, os.WriteFile(dummyFile, content, 0600))

	manifest, err := CreateRelease(dummyFile, NewPlatform("linux", "amd64"), version, outputDir)
	require.NoError(t, err)

	// Check manifest fields
	require.Equal(t, version, manifest.Version)
	require.Equal(t, "dummy-1.2.3.txt", manifest.Name)
	require.Len(t, manifest.Sha256, 32)

	// Check that the file exists and contains the file
	// Expected folder structure in outputDir:
	// outputDir/
	//      dummy.txt
	//   linux-amd64.json
	//
	outPath := filepath.Join(outputDir, manifest.Name)
	outFile, err := os.ReadFile(outPath)
	require.NoError(t, err)
	require.Equal(t, content, outFile)

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
	require.Equal(t, "dummy-1.2.3.txt", m.Name)
}

func TestVersion(t *testing.T) {
	tests := []struct {
		name     string
		v1       Version
		v2       Version
		expected bool
	}{
		{
			name:     "Equal versions",
			v1:       Version("1.0.0"),
			v2:       Version("1.0.0"),
			expected: true,
		},
		{
			name:     "Different versions",
			v1:       Version("1.0.0"),
			v2:       Version("2.0.0"),
			expected: false,
		},
		{
			name:     "Empty versions",
			v1:       Version(""),
			v2:       Version(""),
			expected: true,
		},
		{
			name:     "One empty version",
			v1:       Version("1.0.0"),
			v2:       Version(""),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.v1 == tt.v2
			if result != tt.expected {
				t.Errorf("Version.Equals() = %v, expected %v", result, tt.expected)
			}
		})
	}
}
