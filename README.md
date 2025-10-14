# go-updater

A cross-platform auto-updater library for Go applications that enables seamless automatic updates with secure verification and restart capabilities.

[![Go Reference](https://pkg.go.dev/badge/github.com/arduino/go-updater.svg)](https://pkg.go.dev/github.com/arduino/go-updater)
[![License: GPL v3](https://img.shields.io/badge/License-GPLv3-blue.svg)](https://www.gnu.org/licenses/gpl-3.0)

## Features

- 🚀 **Automatic Updates**: Check for and apply updates automatically
- 🔒 **Secure Verification**: SHA256 checksum validation for all downloads  
- 🌍 **Cross-Platform**: Support for Windows, macOS, and Linux
- 📦 **Archive Support**: Platform-specific archive extraction (see supported formats below)
- 🔄 **Automatic Restart**: Seamlessly restart applications after updates
- 🎯 **Platform Detection**: Automatic OS and architecture detection

## Supported Archive Formats

Archive support varies by platform:

- **Linux**: `.zip`, `.gz`, `.tgz` (tar.gz)
- **macOS**: `.zip` 
- **Windows**: Windows installers created with NSIS

## Installation

```bash
go get github.com/arduino/go-updater
```

## Quick Start

### Basic Example

Here's a simple example of how to integrate auto-updates into your Go application:

```go
package main

import (
    "fmt"
    "log"
    "os"
    
    "github.com/arduino/go-updater/updater"
    "github.com/arduino/go-updater/releaser"
)

func main() {
    // Your current application version
    currentVersion := releaser.Version("1.0.0")
    
    // Create HTTP client for your update server
    client := releaser.NewClient("https://releases.example.com/", "path/to/manifest")
    
    // Check for updates
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
        client,                           // HTTP client
        confirmUpdate,                   // Auto-confirm updates
    )
    
    if err != nil {
        log.Printf("Update check failed: %v", err)
        // Continue with normal application startup
    }

    // Note: If an update was found and applied, the application will be restarted
    // with the new version and the code below will never be executed.
    // Your application logic should be placed here only for cases where:
    // - No update was available
    // - Update check failed
    // - User declined the update
}
```

## Creating Releases

Use the included releaser tool to create releases for your application:

### 1. Build Your Application

```bash
# Build for multiple platforms
GOOS=linux GOARCH=amd64 go build -o myapp-linux-amd64 ./cmd/myapp
```

### 2. Create Release Manifest

```bash
# Create releases
go run github.com/arduino/go-updater/cmd/releaser \\
    -input ./myapp-linux-amd64 \\
    -version 1.2.0 \\
    -platform linux-amd64 \\
    -output ./releases/
```

### 3. Server Setup

Set up an HTTP server to serve your releases. The updater expects this structure:

```
https://releases.example.com/
├── linux-amd64/
│   ├── myapp-linux-amd64
|   ├──linux-amd64.json
├── windows-amd64/
│   ├── myapp-windows-amd64.exe
│   ├── windows-amd64.json
└── darwin-amd64/
    ├── myapp-darwin-amd64
    ├── darwin-amd64.json
```

Example `linux-amd64.json`:

```json
{
    "name": "myapp-linux-amd64",
    "version": "1.2.0",
    "sha256": "abc123def456..."
}
```

## Security Considerations

- **HTTPS Only**: Always use HTTPS for your release server
- **Checksum Verification**: All downloads are automatically verified using SHA256
- **Path Validation**: Archive extraction includes path traversal protection

## License

This software is released under the GNU General Public License version 3. See [LICENSE](LICENSE) for details.

For commercial licensing options, please contact license@arduino.cc.

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)  
5. Open a Pull Request

---

Made with ❤️ by [Arduino](https://www.arduino.cc/)