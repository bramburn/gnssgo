# Changelog

All notable changes to the GNSSGO project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Comprehensive examples in the `examples` directory
- CHANGELOG.md file to track changes
- CONTRIBUTING.md file with guidelines for contributors
- Enhanced documentation in README.md

## [1.1.0] - 2023-06-15

### Changed
- Updated to Go 1.21
- Replaced serial library with go.bug.st/serial for improved serial port communication

## [1.0.0] - 2023-06-06

### Added
- Initial release of GNSSGO
- Go implementation of RTKLIB 2.4.3
- Support for various GNSS systems (GPS, GLONASS, Galileo, QZSS, BeiDou, IRNSS)
- Real-time and post-processing positioning
- Various positioning modes (single, DGPS/DGNSS, Kinematic, Static, PPP-Kinematic, PPP-Static)
- Support for standard formats (RINEX, RTCM, BINEX, etc.)
- Serial, TCP/IP, NTRIP, and file handling
