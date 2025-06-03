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

func main() {
	var (
		outputDir string
		platform  releaser.Platform
	)

	platform = defaultPlatform()

	flag.StringVar(&outputDir, "o", "public", "Output directory for writing updates")
	flag.Var(&platform, "platform", "Target platform in the form OS-ARCH. Defaults to running os/arch or the combination of the environment variables GOOS and GOARCH if both are set.")
	flag.Usage = printUsage
	flag.Parse()

	if flag.NArg() < 2 {
		flag.Usage()
		os.Exit(1)
	}

	inputPath := flag.Arg(0)
	version := flag.Arg(1)

	manifest, err := releaser.CreateRelease(inputPath, platform, version, outputDir)
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
  go-selfupdate [flags] <binary-or-dir> <version>

Positional arguments:
  <binary-or-dir>   Path to the binary file or directory containing binaries
  <version>         Version string to embed in the update metadata

Flags:
`)
	flag.PrintDefaults()
	fmt.Fprintln(os.Stderr, `
Examples:
  go-selfupdate myapp 1.2.3
  go-selfupdate -o public -platform linux-amd64 myapp 1.2.3
  go-selfupdate /tmp/mybinares/ 1.2.3`)
}
