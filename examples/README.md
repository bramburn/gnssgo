# GNSSGO Examples

This directory contains examples demonstrating how to use the GNSSGO library for various GNSS-related tasks.

## Directory Structure

- **basic**: Basic usage of the GNSSGO library (version, time handling, coordinate conversion)
- **rtk**: Real-Time Kinematic (RTK) positioning examples
- **file_handling**: Examples for reading and writing RINEX files
- **serial**: Serial communication with GNSS receivers

## Running the Examples

Each example is a standalone Go program. To run an example:

1. Make sure you have Go installed (version 1.21 or later)
2. Navigate to the example directory
3. Run the example with `go run main.go`

For example:

```bash
cd basic
go run main.go
```

## Using Examples in Your Project

These examples are designed to be educational. To use them in your project:

1. Study the example that most closely matches your needs
2. Copy the relevant code to your project
3. Modify the code as needed for your specific requirements

## Dependencies

All examples depend on the GNSSGO library. If you're running the examples from within the GNSSGO repository, the dependency will be resolved automatically.

If you're using these examples in a separate project, you'll need to import the GNSSGO library:

```go
import "github.com/bramburn/gnssgo"
```

## Notes

- Some examples contain commented-out code that would interact with actual hardware or files. These sections are marked with comments and should be uncommented and modified when used with real data.
- Replace placeholder paths and device names with actual values when using these examples in a real application.
