# GNSSGo

![Go](https://github.com/bramburn/gnssgo/actions/workflows/go.yml/badge.svg)

A Go implementation of RTKLIB 2.4.3 for GNSS development and research.

## Description

GNSSGo is a Go port of the popular RTKLIB C library for GNSS (Global Navigation Satellite System) data processing. It provides tools for real-time and post-processing of GNSS data, supporting various formats and protocols.

RTKLIB is an excellent tool for GNSS development and research. This project recodes RTKLIB in Go to leverage the language's modern features, concurrency model, and cross-platform capabilities.

## Features

- Support for various GNSS systems (GPS, GLONASS, Galileo, QZSS, BeiDou, IRNSS)
- Real-time and post-processing positioning
- Various positioning modes (single, DGPS/DGNSS, Kinematic, Static, PPP-Kinematic, PPP-Static)
- Support for standard formats (RINEX, RTCM, BINEX, etc.)
- Serial, TCP/IP, NTRIP, and file handling
- Thread-safe serial port communication using go.bug.st/serial

## Update History

- 2023/06/06 1.0 - Initial release
- 2023/06/15 1.1 - Updated to Go 1.21, replaced serial library with go.bug.st/serial

## Requirements

- Go 1.21 or later

## Installation

```bash
# Clone the repository
git clone https://github.com/bramburn/gnssgo.git
cd gnssgo

# Build the project
go build ./...

# Run tests
go test ./...
```

## Usage

In file 'go.work', you can choose your app by uncommenting the app path.

For configuring an app, you can define command line arguments in file 'launch.json' under the .vscode directory.

### Serial Port Usage

The library now uses go.bug.st/serial for improved serial port communication. Example:

```go
import (
    "github.com/bramburn/gnssgo"
)

func main() {
    // Open a serial port stream
    // Format: port[:brate[:bsize[:parity[:stopb[:fctr[#port]]]]]]
    // Example: COM1:115200:8:N:1:off
    var stream gnssgo.Stream
    stream.OpenStream(gnssgo.STR_SERIAL, gnssgo.STR_MODE_RW, "COM1:115200:8:N:1")
    
    // Read data
    buff := make([]byte, 1024)
    n := stream.StreamRead(buff, 1024)
    
    // Close when done
    stream.StreamClose()
}
```

## Credits

- Original RTKLIB by T. Takasu
- Go port by Dr. Feng Xuebin, Explore Data Technology (Shenzhen) Ltd.
- Maintained by Bhavesh Ramburn
