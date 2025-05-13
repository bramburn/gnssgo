// Package gnssgo provides GNSS (Global Navigation Satellite System) processing capabilities
// including RTK (Real-Time Kinematic) positioning and precise positioning calculations.
//
// This library is a Go implementation of RTKLIB, providing similar functionality
// for GNSS data processing and positioning.
//
// Main Features:
//
// - Processing RTK (Real-Time Kinematic) GNSS data
// - Calculating precise positioning from GNSS receiver data
// - Support for multiple satellite systems (GPS, GLONASS, Galileo, BeiDou, QZSS, IRNSS)
// - Various positioning modes (single, DGPS, kinematic, static, PPP)
// - RINEX file handling and conversion
//
// Key Types:
//
// - Rtk: RTK control/result type for RTK positioning
// - Sol: Solution type for positioning results
// - PrcOpt: Processing options for positioning
// - Nav: Navigation data including ephemeris
// - ObsD: Observation data
// - RtkSvr: RTK server for continuous positioning
//
// Key Functions:
//
// - Rtk.RtkPos: Performs RTK positioning
// - PntPos: Performs single point positioning
// - Nav.SatPoss: Calculates satellite positions
// - RtkSvr.RtkSvrStart: Starts the RTK server
// - ConvRnx: Converts receiver raw data to RINEX format
//
// For usage examples, see the examples directory.
package gnssgo
