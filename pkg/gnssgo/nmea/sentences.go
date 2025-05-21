package nmea

import (
	"errors"
	"strconv"
	"strings"
	"time"
)

// RMCData represents parsed RMC sentence data (Recommended Minimum Navigation Information)
type RMCData struct {
	Time      string    // UTC time (hhmmss.sss)
	Status    string    // Status (A=active, V=void)
	Latitude  float64   // Latitude in degrees
	LatDir    string    // Latitude direction (N/S)
	Longitude float64   // Longitude in degrees
	LonDir    string    // Longitude direction (E/W)
	Speed     float64   // Speed over ground in knots
	Course    float64   // Course over ground in degrees
	Date      string    // Date (ddmmyy)
	MagVar    float64   // Magnetic variation in degrees
	MagVarDir string    // Magnetic variation direction (E/W)
	Mode      string    // Mode indicator (A=autonomous, D=differential, E=estimated)
	DateTime  time.Time // Combined date and time
}

// VTGData represents parsed VTG sentence data (Track Made Good and Ground Speed)
type VTGData struct {
	TrackTrue     float64 // Track made good (degrees true)
	TrackMagnetic float64 // Track made good (degrees magnetic)
	SpeedKnots    float64 // Speed over ground in knots
	SpeedKmh      float64 // Speed over ground in km/h
	Mode          string  // Mode indicator (A=autonomous, D=differential, E=estimated)
}

// GSAData represents parsed GSA sentence data (GNSS DOP and Active Satellites)
type GSAData struct {
	Mode1  string   // Mode (M=manual, A=automatic)
	Mode2  int      // Fix type (1=no fix, 2=2D fix, 3=3D fix)
	SatIDs []string // List of satellite IDs used in position fix
	PDOP   float64  // Position dilution of precision
	HDOP   float64  // Horizontal dilution of precision
	VDOP   float64  // Vertical dilution of precision
}

// GSVData represents parsed GSV sentence data (GNSS Satellites in View)
type GSVData struct {
	TotalMessages int      // Total number of GSV messages
	MessageNumber int      // Message number
	SatsInView    int      // Number of satellites in view
	Satellites    []GSVSat // Satellite information
}

// GSVSat represents satellite information in a GSV sentence
type GSVSat struct {
	ID        string // Satellite ID
	Elevation int    // Elevation in degrees (0-90)
	Azimuth   int    // Azimuth in degrees (0-359)
	SNR       int    // Signal-to-noise ratio (0-99), -1 if not tracking
}

// ParseRMC parses an RMC sentence
func ParseRMC(sentence string) (RMCData, error) {
	var data RMCData

	// Parse the sentence first
	parsed, err := ParseNMEA(sentence)
	if err != nil {
		return data, err
	}

	if !parsed.Valid {
		return data, errors.New("invalid NMEA sentence")
	}

	// Check if it's an RMC sentence
	if !strings.HasSuffix(parsed.Type, "RMC") {
		return data, errors.New("not an RMC sentence")
	}

	// Check if we have enough fields
	if len(parsed.Fields) < 11 {
		return data, errors.New("not enough fields in RMC sentence")
	}

	// Parse time
	data.Time = parsed.Fields[0]

	// Parse status
	data.Status = parsed.Fields[1]

	// Parse latitude
	if parsed.Fields[2] != "" {
		lat, err := strconv.ParseFloat(parsed.Fields[2], 64)
		if err == nil {
			// Convert NMEA format (DDMM.MMMM) to decimal degrees
			latDeg := float64(int(lat / 100))
			latMin := lat - latDeg*100
			data.Latitude = latDeg + latMin/60

			// Apply direction
			if parsed.Fields[3] == "S" {
				data.Latitude = -data.Latitude
			}
		}
	}
	data.LatDir = parsed.Fields[3]

	// Parse longitude
	if parsed.Fields[4] != "" {
		lon, err := strconv.ParseFloat(parsed.Fields[4], 64)
		if err == nil {
			// Convert NMEA format (DDDMM.MMMM) to decimal degrees
			lonDeg := float64(int(lon / 100))
			lonMin := lon - lonDeg*100
			data.Longitude = lonDeg + lonMin/60

			// Apply direction
			if parsed.Fields[5] == "W" {
				data.Longitude = -data.Longitude
			}
		}
	}
	data.LonDir = parsed.Fields[5]

	// Parse speed
	if parsed.Fields[6] != "" {
		data.Speed, _ = strconv.ParseFloat(parsed.Fields[6], 64)
	}

	// Parse course
	if parsed.Fields[7] != "" {
		data.Course, _ = strconv.ParseFloat(parsed.Fields[7], 64)
	}

	// Parse date
	data.Date = parsed.Fields[8]

	// Parse magnetic variation
	if parsed.Fields[9] != "" {
		data.MagVar, _ = strconv.ParseFloat(parsed.Fields[9], 64)
		if parsed.Fields[10] == "W" {
			data.MagVar = -data.MagVar
		}
	}
	data.MagVarDir = parsed.Fields[10]

	// Parse mode indicator if available
	if len(parsed.Fields) > 11 {
		data.Mode = parsed.Fields[11]
	}

	// Parse combined date and time
	if data.Date != "" && data.Time != "" {
		// Date format: DDMMYY
		day, _ := strconv.Atoi(data.Date[0:2])
		month, _ := strconv.Atoi(data.Date[2:4])
		year, _ := strconv.Atoi(data.Date[4:6])
		year += 2000 // Adjust for century

		// Time format: HHMMSS.SSS
		hour, _ := strconv.Atoi(data.Time[0:2])
		minute, _ := strconv.Atoi(data.Time[2:4])
		second, _ := strconv.Atoi(data.Time[4:6])

		// Create time.Time object
		data.DateTime = time.Date(year, time.Month(month), day, hour, minute, second, 0, time.UTC)
	}

	return data, nil
}

// ParseVTG parses a VTG sentence
func ParseVTG(sentence string) (VTGData, error) {
	var data VTGData

	// Parse the sentence first
	parsed, err := ParseNMEA(sentence)
	if err != nil {
		return data, err
	}

	if !parsed.Valid {
		return data, errors.New("invalid NMEA sentence")
	}

	// Check if it's a VTG sentence
	if !strings.HasSuffix(parsed.Type, "VTG") {
		return data, errors.New("not a VTG sentence")
	}

	// Check if we have enough fields
	if len(parsed.Fields) < 8 {
		return data, errors.New("not enough fields in VTG sentence")
	}

	// Parse track true
	if parsed.Fields[0] != "" {
		data.TrackTrue, _ = strconv.ParseFloat(parsed.Fields[0], 64)
	}

	// Parse track magnetic
	if parsed.Fields[2] != "" {
		data.TrackMagnetic, _ = strconv.ParseFloat(parsed.Fields[2], 64)
	}

	// Parse speed in knots
	if parsed.Fields[4] != "" {
		data.SpeedKnots, _ = strconv.ParseFloat(parsed.Fields[4], 64)
	}

	// Parse speed in km/h
	if parsed.Fields[6] != "" {
		data.SpeedKmh, _ = strconv.ParseFloat(parsed.Fields[6], 64)
	}

	// Parse mode indicator if available
	if len(parsed.Fields) > 8 {
		data.Mode = parsed.Fields[8]
	}

	return data, nil
}
