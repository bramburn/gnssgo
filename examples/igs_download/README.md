# IGS Product Download Example

This example demonstrates how to use the GNSSGO IGS package to download precise ephemeris (SP3) and precise clock (CLK) files from the International GNSS Service (IGS) and its analysis centers.

## Running the Example

```bash
# Download SP3 and CLK files for today from IGS
go run main.go

# Download only SP3 file from JPL for a specific date
go run main.go -sp3 -clk=false -ac jpl -date 2023-05-15

# Download and decompress files
go run main.go -decompress

# Specify a custom output directory
go run main.go -out ./my-data
```

## Command-Line Options

- `-sp3`: Download SP3 file (default: true)
- `-clk`: Download CLK file (default: true)
- `-ac`: Analysis center (igs, cod, emr, esa, gfz, jpl) (default: igs)
- `-date`: Date in YYYY-MM-DD format (default: today)
- `-out`: Output directory (default: ./data)
- `-decompress`: Decompress downloaded files (default: false)

## Available Analysis Centers

- `igs` - International GNSS Service
- `cod` - Center for Orbit Determination in Europe
- `emr` - Natural Resources Canada
- `esa` - European Space Agency
- `gfz` - GeoForschungsZentrum Potsdam
- `jpl` - Jet Propulsion Laboratory

## Notes

- The downloaded files are compressed with Unix compress (.Z extension)
- The decompression feature requires the `uncompress` command to be available on the system
- Files are organized by GPS week in the output directory
