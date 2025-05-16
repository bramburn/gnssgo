/*
* format.go : Format detection and handling for GNSS binary files
*
* This file provides functions for detecting and handling different GNSS binary formats.
*/

package converter

import (
	"strings"

	"github.com/bramburn/gnssgo/pkg/gnssgo"
)

// DetectFormat detects the format of a GNSS binary file based on its extension
// Returns the format code (STRFMT_???) or -1 if unknown
func DetectFormat(file string) int {
	var idx int
	
	// Find the last dot in the filename
	if idx = strings.LastIndex(file, "."); idx < 0 {
		return -1
	}
	
	// Get the extension
	ext := file[idx:]
	
	// Match extension to format
	switch {
	case ext == ".rtcm2":
		return gnssgo.STRFMT_RTCM2
	case ext == ".rtcm3":
		return gnssgo.STRFMT_RTCM3
	case ext == ".gps":
		return gnssgo.STRFMT_OEM4
	case ext == ".ubx":
		return gnssgo.STRFMT_UBX
	case ext == ".log":
		return gnssgo.STRFMT_SS2
	case ext == ".bin":
		return gnssgo.STRFMT_CRES
	case ext == ".stq":
		return gnssgo.STRFMT_STQ
	case ext == ".jps":
		return gnssgo.STRFMT_JAVAD
	case ext == ".bnx":
		return gnssgo.STRFMT_BINEX
	case ext == ".binex":
		return gnssgo.STRFMT_BINEX
	case ext == ".rt17":
		return gnssgo.STRFMT_RT17
	case ext == ".sbf":
		return gnssgo.STRFMT_SEPT
	case ext == ".obs":
		return gnssgo.STRFMT_RINEX
	case ext[len(ext)-1:] == "o":
		return gnssgo.STRFMT_RINEX
	case ext[len(ext)-1:] == "O":
		return gnssgo.STRFMT_RINEX
	case ext == ".rnx":
		return gnssgo.STRFMT_RINEX
	case ext == ".nav":
		return gnssgo.STRFMT_RINEX
	case ext[len(ext)-1:] == "n":
		return gnssgo.STRFMT_RINEX
	case ext[len(ext)-1:] == "N":
		return gnssgo.STRFMT_RINEX
	}
	
	return -1
}

// FormatFromString converts a format string to its corresponding format code
// Returns the format code (STRFMT_???) or -1 if unknown
func FormatFromString(fmts string) int {
	switch fmts {
	case "rtcm2":
		return gnssgo.STRFMT_RTCM2
	case "rtcm3":
		return gnssgo.STRFMT_RTCM3
	case "nov":
		return gnssgo.STRFMT_OEM4
	case "oem3":
		return gnssgo.STRFMT_OEM3
	case "ubx":
		return gnssgo.STRFMT_UBX
	case "ss2":
		return gnssgo.STRFMT_SS2
	case "hemis":
		return gnssgo.STRFMT_CRES
	case "stq":
		return gnssgo.STRFMT_STQ
	case "javad":
		return gnssgo.STRFMT_JAVAD
	case "nvs":
		return gnssgo.STRFMT_NVS
	case "binex":
		return gnssgo.STRFMT_BINEX
	case "rt17":
		return gnssgo.STRFMT_RT17
	case "sbf":
		return gnssgo.STRFMT_SEPT
	case "rinex":
		return gnssgo.STRFMT_RINEX
	}
	return -1
}
