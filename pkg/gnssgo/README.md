# GNSSGO - Go GNSS RTK Library

GNSSGO is a Go implementation of RTKLIB, providing similar functionality for GNSS data processing and positioning. It can be imported into your Go applications to add RTK and other GNSS processing capabilities.

## Main Features

- Processing RTK (Real-Time Kinematic) GNSS data
- Calculating precise positioning from GNSS receiver data
- Support for multiple satellite systems (GPS, GLONASS, Galileo, BeiDou, QZSS, IRNSS)
- Various positioning modes (single, DGPS/DGNSS, kinematic, static, PPP-Kinematic, PPP-Static)
- RINEX file handling and conversion
- Support for standard formats (RINEX, RTCM, BINEX, etc.)
- Serial, TCP/IP, NTRIP, and file handling

## Key Types

- `Rtk`: RTK control/result type for RTK positioning
  Used to perform RTK positioning with observation data

- `Sol`: Solution type for positioning results
  Contains position, velocity, and other solution information

- `PrcOpt`: Processing options for positioning
  Configuration options for positioning algorithms

- `Nav`: Navigation data including ephemeris
  Contains satellite orbit and clock information

- `ObsD`: Observation data
  Contains GNSS observation data from receivers

- `RtkSvr`: RTK server for continuous positioning
  Provides continuous positioning with data streams

- `Stream`: Communication stream
  Handles various types of communication (serial, TCP, file, etc.)

- `RTKStatus`: RTK status information
  Contains information about the current RTK status

## Key Functions

- `Rtk.RtkPos`: Performs RTK positioning
  Process observation data to calculate precise positions

- `PntPos`: Performs single point positioning
  Calculate position from a single receiver's observations

- `Nav.SatPoss`: Calculates satellite positions
  Compute satellite positions from navigation data

- `RtkSvr.RtkSvrStart`: Starts the RTK server
  Begin continuous positioning with data streams

- `ConvRnx`: Converts receiver raw data to RINEX format
  Convert raw receiver data to standard RINEX format

- `ReadRnx`: Reads RINEX files
  Parse RINEX observation and navigation files

- `Stream.OpenStream`: Opens a communication stream
  Establish communication with receivers or other data sources

## Usage Examples

For comprehensive usage examples, see the examples directory in the repository.
Examples include basic usage, RTK positioning, file handling, and serial communication.

## RTCM Message Filtering

The library includes functionality for filtering RTCM messages:

- `DefaultRTCMFilter`: Provides a default filter that excludes unnecessary message types
- `CriticalRTCMFilter`: Provides a filter that only allows critical message types for RTK
- `FilterRTCMMessages`: Filters RTCM messages based on the provided filter

## NMEA Parsing

The library includes functionality for parsing NMEA sentences:

- `ParseNMEA`: Parses an NMEA sentence
- `ParseGGA`: Parses a GGA sentence
- `CalculateNMEAChecksum`: Calculates the checksum for an NMEA sentence

## RTK Status

The library includes functionality for tracking RTK status:

- `RTKStatus`: Contains information about the current RTK status
- `UpdateFromCovariance`: Updates the RTK status based on covariance values
- `UpdateFromNMEA`: Updates the RTK status from a NMEA GGA sentence

## Thread Safety

Most functions are thread-safe and can be used in concurrent Go routines.
The library uses Go's concurrency primitives for thread safety.

## License

This library is licensed under the MIT License. See the LICENSE file for details.
