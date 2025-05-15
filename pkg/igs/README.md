# IGS Product Downloader

This package provides functionality to download precise ephemeris (SP3) and precise clock (CLK) files from the International GNSS Service (IGS) and its analysis centers.

## Features

- Download SP3 (precise orbit) and CLK (precise clock) files
- Support for multiple IGS analysis centers (IGS, COD, EMR, ESA, GFZ, JPL)
- Automatic GPS week and day calculation
- File decompression support

## Usage

### As a Library

```go
package main

import (
	"fmt"
	"time"

	"github.com/bramburn/gnssgo/pkg/igs"
)

func main() {
	// Create a new client with download directory
	client := igs.NewClient("./data")

	// Download SP3 file for today from IGS
	today := time.Now().UTC()
	filePath, err := client.DownloadSP3(today, igs.AnalysisCenterIGS)
	if err != nil {
		fmt.Printf("Error downloading SP3 file: %v\n", err)
		return
	}
	fmt.Printf("Downloaded SP3 file to: %s\n", filePath)

	// Download CLK file for today from IGS
	filePath, err = client.DownloadCLK(today, igs.AnalysisCenterIGS)
	if err != nil {
		fmt.Printf("Error downloading CLK file: %v\n", err)
		return
	}
	fmt.Printf("Downloaded CLK file to: %s\n", filePath)

	// Decompress the file
	decompressedPath, err := igs.DecompressFile(filePath)
	if err != nil {
		fmt.Printf("Error decompressing file: %v\n", err)
		return
	}
	fmt.Printf("Decompressed to: %s\n", decompressedPath)
}
```

### Command Line Tool

The package includes a command-line tool for downloading IGS products:

```bash
# Download SP3 file for today from IGS
go run pkg/igs/cmd/main.go -type sp3 -ac igs -out ./data

# Download CLK file for a specific date from JPL
go run pkg/igs/cmd/main.go -type clk -ac jpl -date 2023-05-15 -out ./data -decompress

# List available analysis centers
go run pkg/igs/cmd/main.go -list-centers

# List available product types
go run pkg/igs/cmd/main.go -list-products
```

## Available Analysis Centers

- `igs` - International GNSS Service
- `cod` - Center for Orbit Determination in Europe
- `emr` - Natural Resources Canada
- `esa` - European Space Agency
- `gfz` - GeoForschungsZentrum Potsdam
- `jpl` - Jet Propulsion Laboratory

## Available Product Types

- `sp3` - Precise orbit files (.sp3)
- `clk` - Precise clock files (.clk)

## Notes

- The downloaded files are compressed with Unix compress (.Z extension)
- The `DecompressFile` function requires the `uncompress` command to be available on the system
- Files are organized by GPS week in the download directory
