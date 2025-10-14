// This file is part of go-updater.
//
// Copyright 2025 ARDUINO SA (http://www.arduino.cc/)
//
// This software is released under the GNU General Public License version 3,
// which covers the main part of go-updater.
// The terms of this license can be found at:
// https://www.gnu.org/licenses/gpl-3.0.en.html
//
// You can be released from the requirements of the above licenses by purchasing
// a commercial license. Buying such a license is mandatory if you want to
// modify or otherwise use the software for commercial activities involving the
// Arduino software without disclosing the source code of your own applications.
// To purchase a commercial license, send an email to license@arduino.cc.

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"

	"github.com/arduino/go-updater/releaser"
)

var version = "dev"

func main() {
	var (
		isVersion bool
		outputDir string
		platform  releaser.Platform
	)

	platform = defaultPlatform()

	flag.StringVar(&outputDir, "o", "public", "Output directory for writing updates")
	flag.Var(&platform, "platform", "Target platform in the form OS-ARCH. Defaults to running os/arch or the combination of the environment variables GOOS and GOARCH if both are set.")
	flag.BoolVar(&isVersion, "version", false, "Print the version of the releaser tool")
	flag.Usage = printUsage
	flag.Parse()

	if isVersion {
		fmt.Println("Releaser version:", version)
		return
	}

	if flag.NArg() < 2 {
		flag.Usage()
		os.Exit(1)
	}

	inputPath := flag.Arg(0)
	version := flag.Arg(1)

	manifest, err := releaser.CreateRelease(inputPath, platform, releaser.Version(version), outputDir)
	if err != nil {
		log.Fatalf("could not create release: %v", err)
	}
	fmt.Println("Release created successfully!")
	jsonBytes, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		log.Fatalf("could not marshal manifest to JSON: %v", err)
	}
	fmt.Println(string(jsonBytes))
}

func defaultPlatform() releaser.Platform {
	goos := os.Getenv("GOOS")
	goarch := os.Getenv("GOARCH")
	if goos != "" && goarch != "" {
		return releaser.NewPlatform(goos, goarch)
	}
	return releaser.NewPlatform(runtime.GOOS, runtime.GOARCH)
}

func printUsage() {
	fmt.Fprintf(os.Stderr, `
Usage:
  releaser [flags] <file> <version>

Positional arguments:
  <file>    Path to the release file
  <version> Version string to embed in the update metadata

Flags:
`)
	flag.PrintDefaults()
	fmt.Fprintln(os.Stderr, `
Examples:
  releaser myapp-1.2.3.zip 1.2.3
  releaser -o public -platform linux-amd64 myapp-1.2.3.tar.gz 1.2.3`)
}
