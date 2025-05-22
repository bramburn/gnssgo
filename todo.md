# GNSSGO Project TODO List

This document outlines the remaining tasks to complete the project restructuring and development, as well as tracking placeholders and incomplete code that need to be addressed in future development.

## Project Structure

- [x] Set up monorepo structure with `/pkg/gnssgo` for core library
- [x] Create `/gui` directory for Wails application
- [x] Update go.work file to include new directories
- [x] Update import paths in tests
- [ ] Complete the Wails GUI application development
- [ ] Add more comprehensive documentation for the GUI application

## Stream Implementation

### Serial Communication (`pkg/gnssgo/stream/serial.go`)
- [x] Basic implementation completed
- [ ] Add support for hardware flow control
- [ ] Implement better error handling and recovery
- [ ] Add support for more serial port settings
- [ ] Add comprehensive logging for debugging

### File I/O (`pkg/gnssgo/stream/file.go`)
- [x] Basic implementation completed
- [x] Complete the time-tagged file implementation
  - [x] Implement proper file swapping based on time
  - [x] Complete the `readfiletime` function for reading data at specific times
  - [x] Implement proper file path keyword replacement in `reppath`
- [x] Add support for compressed files
- [x] Add better error handling and recovery
- [x] Add comprehensive tests for file stream functionality

### TCP Communication (`pkg/gnssgo/stream/tcp.go`)
- [x] Basic implementation completed
- [ ] Improve connection handling and recovery
- [ ] Add support for TLS/SSL
- [ ] Implement better timeout and retry mechanisms
- [ ] Add comprehensive logging for debugging

### UDP Communication (`pkg/gnssgo/stream/udp.go`)
- [x] Basic implementation completed
- [ ] Add support for multicast
- [ ] Implement better error handling and recovery
- [ ] Add comprehensive logging for debugging

### NTRIP Client (`pkg/gnssgo/stream/ntrip.go`)
- [x] Basic implementation completed
- [x] Enhanced implementation with error handling and retry logic
- [x] Complete the `ReadNtrip` function for reading data from the NTRIP connection
- [x] Complete the `WriteNtrip` function for writing data to the NTRIP connection
- [x] Implement better handling of RTCM messages
- [x] Add comprehensive logging for debugging
- [x] Consolidate NTRIP client implementations to use the enhanced version
- [x] Add comprehensive tests for the NTRIP implementation
- [ ] Add support for NTRIP 2.0

## GUI Application

- [ ] Fix Wails initialization issues
- [ ] Implement core functionality in the GUI
- [ ] Create proper UI components for GNSS data visualization
- [ ] Add settings and configuration screens
- [ ] Implement data import/export functionality
- [ ] Add real-time data processing capabilities

## Testing

- [x] Fix failing tests that were skipped during restructuring
- [x] Update test data paths to work with new structure
- [ ] Add tests for new GUI functionality
- [ ] Implement integration tests between core library and GUI

## Documentation

- [ ] Create comprehensive API documentation
- [ ] Add usage examples for the GUI application
- [ ] Update installation instructions for the new structure
- [ ] Create developer guide for contributing to the project

## Build and Deployment

- [ ] Set up CI/CD pipeline for automated testing
- [ ] Create build scripts for different platforms (Windows, macOS, Linux)
- [ ] Package the application for distribution
- [ ] Set up release process for versioning

## Performance Improvements

- [ ] Profile the application to identify bottlenecks
- [ ] Optimize critical algorithms
- [ ] Implement parallel processing where applicable
- [ ] Reduce memory usage for large datasets

## Future Features

- [ ] Add support for additional GNSS constellations
- [ ] Implement real-time data streaming
- [ ] Add advanced data analysis tools
- [ ] Create visualization components for satellite positions and signal quality

## RTCM Message Handling

### RTCM Parser (`pkg/gnssgo/rtcm/rtcm.go`)
- [x] Implement a comprehensive RTCM 3.x parser
- [x] Add support for all RTCM 3.3 message types (1074-1077, 1084-1087, 1094-1097, 1124-1127, 1057-1068)
- [x] Implement MSM7 parsing
- [x] Add proper error handling and validation
- [x] Add comprehensive logging for debugging
- [ ] Complete implementation of SSR message decoders
- [ ] Add more comprehensive tests with real RTCM data
- [ ] Optimize performance for high-throughput applications

## RTK Positioning

### RTK Algorithms (`pkg/gnssgo/rtk/rtk.go`)
- [ ] Port the RTK positioning algorithms from C to Go
- [ ] Implement the RTK filter
- [ ] Add support for multiple GNSS constellations
- [ ] Implement proper error handling and validation
- [ ] Add comprehensive logging for debugging

## Hardware Support

### TOP708 Receiver (`pkg/gnssgo/hardware/topgnss/top708/top708.go`)
- [ ] Implement the TOP708 receiver interface
- [ ] Add support for configuration commands
- [ ] Implement proper error handling and recovery
- [ ] Add comprehensive logging for debugging
