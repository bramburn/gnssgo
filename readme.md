# GNSSGo

![Go](https://github.com/bramburn/gnssgo/actions/workflows/go.yml/badge.svg)

A Go implementation of RTKLIB 2.4.3 for GNSS development and research.

## Description

GNSSGo is a Go port of the popular RTKLIB C library for GNSS (Global Navigation Satellite System) data processing. It provides tools for real-time and post-processing of GNSS data, supporting various formats and protocols.

RTKLIB is an excellent tool for GNSS development and research. This project recodes RTKLIB in Go to leverage the language's modern features, concurrency model, and cross-platform capabilities.

This library can be imported into your Go applications to add RTK and other GNSS processing capabilities.

## Features

- Support for various GNSS systems (GPS, GLONASS, Galileo, QZSS, BeiDou, IRNSS)
- Real-time and post-processing positioning
- Various positioning modes (single, DGPS/DGNSS, Kinematic, Static, PPP-Kinematic, PPP-Static)
- Support for standard formats (RINEX, RTCM, BINEX, etc.)
- Serial, TCP/IP, NTRIP, and file handling
- Thread-safe serial port communication using go.bug.st/serial

## Update History

See [CHANGELOG.md](CHANGELOG.md) for a detailed history of changes.

- 2023/06/06 1.0 - Initial release
- 2023/06/15 1.1 - Updated to Go 1.21, replaced serial library with go.bug.st/serial

## Requirements

- Go 1.21 or later

## Installation

### As a Library

To use GNSSGO as a library in your Go project:

```bash
# Add the library to your project
go get github.com/bramburn/gnssgo
```

Then import it in your code:

```go
import "github.com/bramburn/gnssgo"
```

### For Development

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

### As a Library

Check out the [examples](examples/) directory for comprehensive examples of how to use GNSSGO in your applications.

### Using the Included Applications

In file 'go.work', you can choose your app by uncommenting the app path.

For configuring an app, you can define command line arguments in file 'launch.json' under the .vscode directory.

### Key Components

#### RTK Positioning

```go
// Initialize RTK control structure
var rtk gnssgo.Rtk

// Initialize processing options
var opt gnssgo.PrcOpt

// Set default processing options
opt.Mode = gnssgo.PMODE_KINEMA // Kinematic mode
opt.NavSys = gnssgo.SYS_GPS | gnssgo.SYS_GLO // Use GPS and GLONASS

// Initialize RTK control with options
rtk.InitRtk(&opt)

// Process observation data
rtk.RtkPos(obsData, numObs, &navData)

// Get solution
solution := rtk.RtkSol
```

#### Serial Port Usage

The library uses go.bug.st/serial for improved serial port communication:

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

#### RINEX File Handling

```go
// Initialize navigation data structure
var nav gnssgo.Nav

// Initialize observation data structure
var obs gnssgo.Obs

// Read a RINEX navigation file
status := gnssgo.ReadRnx("path/to/nav.rnx", 1, "", nil, &nav, nil)

// Read a RINEX observation file
status := gnssgo.ReadRnx("path/to/obs.rnx", 0, "", &obs, nil, nil)
```

## Documentation

For detailed documentation on the library's functions and types, see the [doc.go](src/doc.go) file.

## Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## License

This project is licensed under the same terms as the original RTKLIB. See [LICENSE](LICENSE) for details.

## Credits

- Original RTKLIB by T. Takasu
- Go port by Dr. Feng Xuebin, Explore Data Technology (Shenzhen) Ltd.
- Maintained by Bhavesh Ramburn
