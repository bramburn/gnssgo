package gnssgo

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// NMEA sentence types
const (
	NMEA_GGA = "GGA" // Global Positioning System Fix Data
	NMEA_RMC = "RMC" // Recommended Minimum Navigation Information
	NMEA_GSA = "GSA" // GPS DOP and Active Satellites
	NMEA_GSV = "GSV" // Satellites in View
	NMEA_GLL = "GLL" // Geographic Position - Latitude/Longitude
	NMEA_VTG = "VTG" // Track Made Good and Ground Speed
	NMEA_ZDA = "ZDA" // Time & Date
)

// NMEASentence represents a parsed NMEA sentence
type NMEASentence struct {
	Type     string   // Sentence type (GGA, RMC, etc.)
	Fields   []string // Fields in the sentence
	Valid    bool     // Whether the sentence is valid
	Checksum string   // Checksum of the sentence
}

// GGAData represents parsed GGA sentence data
type GGAData struct {
	Time       string  // UTC time (hhmmss.sss)
	Latitude   float64 // Latitude in degrees
	LatDir     string  // Latitude direction (N/S)
	Longitude  float64 // Longitude in degrees
	LonDir     string  // Longitude direction (E/W)
	Quality    int     // Fix quality (0=invalid, 1=GPS fix, 2=DGPS fix, 4=RTK fix, 5=Float RTK)
	NumSats    int     // Number of satellites
	HDOP       float64 // Horizontal dilution of precision
	Altitude   float64 // Altitude above mean sea level
	AltUnit    string  // Altitude unit (M=meters)
	GeoidSep   float64 // Geoid separation
	GeoidUnit  string  // Geoid separation unit (M=meters)
	DGPSAge    float64 // Age of differential corrections (seconds)
	DGPSStaID  string  // DGPS station ID
}

// ParseNMEA parses an NMEA sentence
func ParseNMEA(sentence string) (NMEASentence, error) {
	result := NMEASentence{
		Valid: false,
	}

	// Check for minimum length
	if len(sentence) < 6 {
		return result, errors.New("sentence too short")
	}

	// Check for valid start character
	if sentence[0] != '$' {
		return result, errors.New("invalid start character")
	}

	// Extract checksum if present
	checksumPos := strings.LastIndex(sentence, "*")
	var data string
	if checksumPos != -1 && checksumPos < len(sentence)-2 {
		data = sentence[:checksumPos]
		result.Checksum = sentence[checksumPos+1:]
		
		// Verify checksum
		calcChecksum := CalculateNMEAChecksum(data[1:])
		if strings.ToUpper(result.Checksum) != strings.ToUpper(calcChecksum) {
			return result, fmt.Errorf("checksum mismatch: got %s, expected %s", result.Checksum, calcChecksum)
		}
	} else {
		data = sentence
	}

	// Split into fields
	fields := strings.Split(data, ",")
	if len(fields) < 2 {
		return result, errors.New("not enough fields")
	}

	// Extract sentence type
	typeField := strings.TrimPrefix(fields[0], "$")
	if len(typeField) < 3 {
		return result, errors.New("invalid sentence type")
	}
	
	// Extract the actual type (last 3 characters)
	result.Type = typeField[len(typeField)-3:]
	result.Fields = fields[1:]
	result.Valid = true

	return result, nil
}

// CalculateNMEAChecksum calculates the checksum for an NMEA sentence
func CalculateNMEAChecksum(data string) string {
	var checksum uint8
	for i := 0; i < len(data); i++ {
		checksum ^= data[i]
	}
	return fmt.Sprintf("%02X", checksum)
}

// ParseGGA parses a GGA sentence
func ParseGGA(sentence NMEASentence) (GGAData, error) {
	var data GGAData
	
	if !sentence.Valid || sentence.Type != NMEA_GGA {
		return data, errors.New("not a valid GGA sentence")
	}
	
	if len(sentence.Fields) < 14 {
		return data, errors.New("not enough fields in GGA sentence")
	}
	
	// Parse time
	data.Time = sentence.Fields[0]
	
	// Parse latitude
	if lat, err := parseLatLon(sentence.Fields[1]); err == nil {
		data.Latitude = lat
	}
	data.LatDir = sentence.Fields[2]
	
	// Parse longitude
	if lon, err := parseLatLon(sentence.Fields[3]); err == nil {
		data.Longitude = lon
	}
	data.LonDir = sentence.Fields[4]
	
	// Parse fix quality
	if quality, err := strconv.Atoi(sentence.Fields[5]); err == nil {
		data.Quality = quality
	}
	
	// Parse number of satellites
	if numSats, err := strconv.Atoi(sentence.Fields[6]); err == nil {
		data.NumSats = numSats
	}
	
	// Parse HDOP
	if hdop, err := strconv.ParseFloat(sentence.Fields[7], 64); err == nil {
		data.HDOP = hdop
	}
	
	// Parse altitude
	if alt, err := strconv.ParseFloat(sentence.Fields[8], 64); err == nil {
		data.Altitude = alt
	}
	data.AltUnit = sentence.Fields[9]
	
	// Parse geoid separation
	if geoid, err := strconv.ParseFloat(sentence.Fields[10], 64); err == nil {
		data.GeoidSep = geoid
	}
	data.GeoidUnit = sentence.Fields[11]
	
	// Parse DGPS age
	if age, err := strconv.ParseFloat(sentence.Fields[12], 64); err == nil {
		data.DGPSAge = age
	}
	
	// Parse DGPS station ID
	data.DGPSStaID = sentence.Fields[13]
	
	return data, nil
}

// parseLatLon parses a latitude or longitude string from NMEA format to decimal degrees
func parseLatLon(coord string) (float64, error) {
	if coord == "" {
		return 0, errors.New("empty coordinate")
	}
	
	// NMEA format: ddmm.mmmmm (latitude) or dddmm.mmmmm (longitude)
	var degrees, minutes float64
	var err error
	
	if len(coord) >= 3 {
		// Extract degrees
		if len(coord) >= 5 { // Longitude (3 digits for degrees)
			degrees, err = strconv.ParseFloat(coord[:3], 64)
			if err != nil {
				return 0, err
			}
			minutes, err = strconv.ParseFloat(coord[3:], 64)
		} else { // Latitude (2 digits for degrees)
			degrees, err = strconv.ParseFloat(coord[:2], 64)
			if err != nil {
				return 0, err
			}
			minutes, err = strconv.ParseFloat(coord[2:], 64)
		}
		
		if err != nil {
			return 0, err
		}
		
		// Convert to decimal degrees
		return degrees + minutes/60.0, nil
	}
	
	return 0, errors.New("invalid coordinate format")
}
