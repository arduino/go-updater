// This file is part of go-updater.
//
// SPDX-FileCopyrightText: Arduino s.r.l. and/or its affiliated companies
// SPDX-License-Identifier: GPL-3.0-or-later

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
