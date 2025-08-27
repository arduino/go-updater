package releaser

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type Manifest struct {
	Name    string  `json:"name"`
	Version Version `json:"version"`
	Sha256  []byte  `json:"sha256"`
}

type Version string

func (v Version) String() string {
	return string(v)
}

func (v Version) Equals(other Version) bool {
	return v.String() == other.String()
}

// CreateRelease creates a release manifest and copies the input file to the output directory.
func CreateRelease(inputPath string, platform Platform, version Version, outputDir string) (Manifest, error) {
	fi, err := os.Stat(inputPath)
	if err != nil {
		return Manifest{}, fmt.Errorf("could not stat input: %w", err)
	}
	if fi.IsDir() {
		return Manifest{}, fmt.Errorf("input path must be a file, not a directory: %s", inputPath)
	}

	fileName := filepath.Base(inputPath)

	if !strings.Contains(fileName, version.String()) {
		return Manifest{}, fmt.Errorf("input file %s must contain the version string %s", inputPath, version.String())
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return Manifest{}, fmt.Errorf("could not create output dir: %s. %w", outputDir, err)
	}

	// Prepare output paths
	jsonPath := filepath.Join(outputDir, platform.String()+".json")
	targetPath := filepath.Join(outputDir, fileName)

	// Now calculate the SHA256 of the file itself
	data, err := os.Open(inputPath)
	if err != nil {
		return Manifest{}, fmt.Errorf("could not open file for hashing: %w", err)
	}
	defer data.Close()
	sha := sha256.New()
	if _, err := io.Copy(sha, data); err != nil {
		return Manifest{}, fmt.Errorf("could not hash file: %w", err)
	}

	if err := copyFile(inputPath, targetPath); err != nil {
		return Manifest{}, err
	}

	c := Manifest{
		Name:    filepath.Base(inputPath),
		Version: version,
		Sha256:  sha.Sum(nil),
	}

	jsonFile, err := os.Create(jsonPath)
	if err != nil {
		return Manifest{}, fmt.Errorf("could not create json file: %w", err)
	}
	defer jsonFile.Close()
	enc := json.NewEncoder(jsonFile)
	enc.SetIndent("", "    ")
	err = enc.Encode(c)
	if err != nil {
		return Manifest{}, fmt.Errorf("could not marshal json: %w", err)
	}

	return c, nil
}

func copyFile(src, dest string) error {
	fDest, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("could not create target file: %w", err)
	}
	defer fDest.Close()

	fSrc, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("could not open input file: %w", err)
	}
	defer fSrc.Close()

	if _, err := io.Copy(fDest, fSrc); err != nil {
		return fmt.Errorf("could not copy input file to target: %w", err)
	}
	return nil
}
