// Package nmea provides functionality for parsing NMEA sentences
package nmea

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
)

// NMEASentence represents a parsed NMEA sentence
type NMEASentence struct {
	Raw      string   // Raw NMEA sentence
	Type     string   // Sentence type (e.g., GGA, RMC)
	Fields   []string // Fields in the sentence
	Valid    bool     // Whether the sentence is valid
	Checksum string   // Checksum of the sentence
}

// GGAData represents parsed GGA sentence data
type GGAData struct {
	Time      string  // UTC time (hhmmss.sss)
	Latitude  float64 // Latitude in degrees
	LatDir    string  // Latitude direction (N/S)
	Longitude float64 // Longitude in degrees
	LonDir    string  // Longitude direction (E/W)
	Quality   int     // Fix quality (0=invalid, 1=GPS fix, 2=DGPS fix, 4=RTK fix, 5=Float RTK)
	NumSats   int     // Number of satellites
	HDOP      float64 // Horizontal dilution of precision
	Altitude  float64 // Altitude above mean sea level
	AltUnit   string  // Altitude unit (M=meters)
	GeoidSep  float64 // Geoid separation
	GeoidUnit string  // Geoid separation unit (M=meters)
	DGPSAge   float64 // Age of differential corrections (seconds)
	DGPSStaID string  // DGPS station ID
}

// ParseNMEA parses an NMEA sentence
func ParseNMEA(sentence string) (NMEASentence, error) {
	result := NMEASentence{
		Raw:   sentence,
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

	// Extract the actual type
	result.Type = typeField
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
func ParseGGA(sentence string) (GGAData, error) {
	var data GGAData

	// Parse the sentence first
	parsed, err := ParseNMEA(sentence)
	if err != nil {
		return data, err
	}

	if !parsed.Valid {
		return data, errors.New("invalid NMEA sentence")
	}

	// Check if it's a GGA sentence
	if !strings.HasSuffix(parsed.Type, "GGA") {
		return data, errors.New("not a GGA sentence")
	}

	// Check if we have enough fields
	if len(parsed.Fields) < 14 {
		return data, errors.New("not enough fields in GGA sentence")
	}

	// Parse time
	data.Time = parsed.Fields[0]

	// Parse latitude
	if parsed.Fields[1] != "" {
		lat, err := strconv.ParseFloat(parsed.Fields[1], 64)
		if err == nil {
			// Convert NMEA format (DDMM.MMMM) to decimal degrees
			latDeg := math.Floor(lat / 100.0)
			latMin := lat - latDeg*100.0
			data.Latitude = latDeg + latMin/60.0

			// Apply direction
			if parsed.Fields[2] == "S" {
				data.Latitude = -data.Latitude
			}
		}
	}
	data.LatDir = parsed.Fields[2]

	// Parse longitude
	if parsed.Fields[3] != "" {
		lon, err := strconv.ParseFloat(parsed.Fields[3], 64)
		if err == nil {
			// Convert NMEA format (DDDMM.MMMM) to decimal degrees
			lonDeg := math.Floor(lon / 100.0)
			lonMin := lon - lonDeg*100.0
			data.Longitude = lonDeg + lonMin/60.0

			// Apply direction
			if parsed.Fields[4] == "W" {
				data.Longitude = -data.Longitude
			}
		}
	}
	data.LonDir = parsed.Fields[4]

	// Parse fix quality
	if parsed.Fields[5] != "" {
		quality, err := strconv.Atoi(parsed.Fields[5])
		if err == nil {
			data.Quality = quality
		}
	}

	// Parse number of satellites
	if parsed.Fields[6] != "" {
		sats, err := strconv.Atoi(parsed.Fields[6])
		if err == nil {
			data.NumSats = sats
		}
	}

	// Parse HDOP
	if parsed.Fields[7] != "" {
		hdop, err := strconv.ParseFloat(parsed.Fields[7], 64)
		if err == nil {
			data.HDOP = hdop
		}
	}

	// Parse altitude
	if parsed.Fields[8] != "" {
		alt, err := strconv.ParseFloat(parsed.Fields[8], 64)
		if err == nil {
			data.Altitude = alt
		}
	}
	data.AltUnit = parsed.Fields[9]

	// Parse geoid separation
	if parsed.Fields[10] != "" {
		geoid, err := strconv.ParseFloat(parsed.Fields[10], 64)
		if err == nil {
			data.GeoidSep = geoid
		}
	}
	data.GeoidUnit = parsed.Fields[11]

	// Parse age of differential
	if parsed.Fields[12] != "" {
		age, err := strconv.ParseFloat(parsed.Fields[12], 64)
		if err == nil {
			data.DGPSAge = age
		}
	}

	// Parse DGPS station ID
	data.DGPSStaID = parsed.Fields[13]

	return data, nil
}

// FindNMEASentences finds all NMEA sentences in a string
func FindNMEASentences(data string) []string {
	var sentences []string
	lines := strings.Split(data, "\r\n")

	for _, line := range lines {
		if strings.HasPrefix(line, "$") {
			sentences = append(sentences, line)
		}
	}

	return sentences
}

// GetFixQualityName returns a string representation of the fix quality
func GetFixQualityName(quality int) string {
	switch quality {
	case 0:
		return "NONE"
	case 1:
		return "SINGLE"
	case 2:
		return "DGPS"
	case 4:
		return "FIX"
	case 5:
		return "FLOAT"
	default:
		return "UNKNOWN"
	}
}
