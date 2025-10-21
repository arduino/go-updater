# go-updater

A cross-platform updater library for Go applications.
On Linux and macOS, it replaces the executable (or .app for macOS) and launches the new version, while on Windows, it runs the installer with the required privileges.

## Installation

```bash
go get github.com/arduino/go-updater
```

## Quick Start

### 1. Basic Example

Here's a simple example of how to integrate auto-updates into your Go application.

Add a `version` variable in the main package. Note that this variable will be overwritten during the build process through ldflags flag.

```go
package main

import (
    "fmt"
    "log"
    "os"
    
    "github.com/arduino/go-updater/updater"
    "github.com/arduino/go-updater/releaser"
)

var version = "0.0.0" 

func main() {
    // Your current application version
    currentVersion := releaser.Version(version)
    
    // Create HTTP client for your update server
    // The base URL should point to the parent directory containing platform-specific folders
    // The client will automatically discover the correct manifest.json based on the current platform
    // e.g., running on Linux will look for: https://releases.example.com/path/to/release/linux-amd64.json
    client := releaser.NewClient("https://releases.example.com/", "path/to/release/")
    
    fmt.Println("Checking for updates...")
    executablePath, err := os.Executable()
	if err != nil {
		panic("could not get executable path")
	}
    confirmUpdate := func(current, target releaser.Version) bool {
       return true 
    }
    err := updater.CheckForUpdates(
        executablePath,                    // Path to current executable
        currentVersion,                    // Current version
        client,                            // HTTP releaser client
        confirmUpdate,                     // Auto-confirm updates
    )
    if err != nil {
        panic(err)
    }

    // Note: If an update was found and applied, the application will be restarted
    // with the new version and the code below will never be executed.
    // Your application logic should be placed here only for cases where:
    // - No update was available (err == nil)
    // - Update check failed (e.g., error fetching the release)
    // - User declined the update (err == nil)
}
```

### 2. Build Your Application

Build your application with a specific version by following these requirements:

- Use the LDFLAGS `-X` flag to set the version at build time
- Include the version in the output filename (this is mandatory for the releaser tool to function correctly)

```bash
GOOS=linux GOARCH=amd64 go build -o myapp-linux-amd64-1.0.0  -ldflags="-X 'main.version=1.0.0'" ./cmd/myapp
```

### 3. Create Release Manifest

```bash
go run github.com/arduino/go-updater/cmd/releaser ./myapp-linux-amd64-1.0.0  1.0.0  -platform linux-amd64   -o ./releases/

Release created successfully!
{
  "name": "myapp-linux-amd64-1.0.0",
  "version": "1.0.0",
  "sha256": "6238F9cnMS8ete3kfDnD9Yk7iDFMWLBX31HXHmii734="
}
```

where:

- `name` is the name of the executable/archive
- `version` is the version of the release
- `sha256` is the sha256 of the executable/archive

### 4. Server Setup

Set up an HTTP server to serve your releases. The updater expects this structure where each platform has its own json file:

```
https://releases.example.com/                    <- Base URL used in NewClient()
├── /path/to/release/                    
   ├── myapp-linux-amd64-1.0.0.tar.gz            <- Actual executable/archive
   ├── myapp-windows-1.0.0-installer.exe
   ├── myapp-darwin-1.0.0.zip
   |
   └── darwin-amd64.json
   └── linux-amd64.json                          <- Platform manifest 
   └── windows-amd64.json
```

## License

This software is released under the GNU General Public License version 3. See the [LICENSE](LICENSE) file for complete details.

For commercial licensing options, please contact license@arduino.cc.

---

Made with ❤️ by [Arduino](https://www.arduino.cc/)
