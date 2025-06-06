package releaser

import (
	"archive/zip"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type Manifest struct {
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

func CreateRelease(inputPath string, platform Platform, version Version, outputDir string) (Manifest, error) {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return Manifest{}, fmt.Errorf("could not create output dir: %s. %w", outputDir, err)
	}
	// Prepare output paths
	jsonPath := filepath.Join(outputDir, platform.String()+".json")
	versionDir := filepath.Join(outputDir, version.String())
	zipPath := filepath.Join(versionDir, platform.String()+".zip")

	if err := os.MkdirAll(versionDir, 0755); err != nil {
		return Manifest{}, fmt.Errorf("could not create version dir: %w", err)
	}

	// Create the zip file
	zipFile, err := os.Create(zipPath)
	if err != nil {
		return Manifest{}, fmt.Errorf("could not create zip file: %w", err)
	}
	zipWriter := zip.NewWriter(zipFile)

	fi, err := os.Stat(inputPath)
	if err != nil {
		return Manifest{}, fmt.Errorf("could not stat input: %w", err)
	}

	if fi.IsDir() {
		// Recursively add directory contents to the zip
		err = filepath.Walk(inputPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			relPath, err := filepath.Rel(filepath.Dir(inputPath), path)
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			return addFileToZip(zipWriter, path, relPath)
		})
		if err != nil {
			return Manifest{}, fmt.Errorf("could not zip directory: %w", err)
		}
	} else {
		// Single file: make it executable before zipping
		err = addFileToZip(zipWriter, inputPath, filepath.Base(inputPath))
		if err != nil {
			return Manifest{}, fmt.Errorf("could not zip file: %w", err)
		}
	}

	zipWriter.Close()
	zipFile.Close()

	// Now calculate the SHA256 of the zip file itself
	zipData, err := os.Open(zipPath)
	if err != nil {
		return Manifest{}, fmt.Errorf("could not open zip for hashing: %w", err)
	}
	defer zipData.Close()
	sha := sha256.New()
	if _, err := io.Copy(sha, zipData); err != nil {
		return Manifest{}, fmt.Errorf("could not hash zip: %w", err)
	}

	c := Manifest{
		Version: version,
		Sha256:  sha.Sum(nil),
	}
	b, err := json.MarshalIndent(c, "", "    ")
	if err != nil {
		return Manifest{}, fmt.Errorf("could not marshal json: %w", err)
	}
	if err := os.WriteFile(jsonPath, b, 0600); err != nil {
		return Manifest{}, fmt.Errorf("could not write json file: %w", err)
	}

	return c, nil
}

// addFileToZip adds a file to the zip archive.
func addFileToZip(zipWriter *zip.Writer, filePath, zipPath string) error {
	info, err := os.Stat(filePath)
	if err != nil {
		return err
	}
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}
	header.Name = zipPath
	// Set the executable bit in the zip header (for Unix systems)
	header.SetMode(info.Mode() | 0111)

	w, err := zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}

	_, err = io.Copy(w, file)
	return err
}
