package nmea

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

// NMEATime converts NMEA time string (HHMMSS.SSS) to time.Time
func NMEATime(timeStr string, dateStr string) (time.Time, error) {
	if timeStr == "" {
		return time.Time{}, fmt.Errorf("empty time string")
	}

	// Parse time components
	var hour, minute, second, millisecond int
	var err error

	// Handle time format: HHMMSS.SSS
	if len(timeStr) >= 6 {
		hour, err = strconv.Atoi(timeStr[0:2])
		if err != nil {
			return time.Time{}, fmt.Errorf("invalid hour: %s", timeStr[0:2])
		}

		minute, err = strconv.Atoi(timeStr[2:4])
		if err != nil {
			return time.Time{}, fmt.Errorf("invalid minute: %s", timeStr[2:4])
		}

		second, err = strconv.Atoi(timeStr[4:6])
		if err != nil {
			return time.Time{}, fmt.Errorf("invalid second: %s", timeStr[4:6])
		}

		// Parse milliseconds if available
		if len(timeStr) > 7 {
			msStr := timeStr[7:]
			// Pad with zeros to ensure 3 digits
			for len(msStr) < 3 {
				msStr += "0"
			}
			// Truncate to 3 digits if longer
			if len(msStr) > 3 {
				msStr = msStr[:3]
			}
			millisecond, err = strconv.Atoi(msStr)
			if err != nil {
				millisecond = 0
			}
		}
	} else {
		return time.Time{}, fmt.Errorf("invalid time format: %s", timeStr)
	}

	// If date string is provided, parse it
	if dateStr != "" && len(dateStr) >= 6 {
		// Date format: DDMMYY
		day, err := strconv.Atoi(dateStr[0:2])
		if err != nil {
			return time.Time{}, fmt.Errorf("invalid day: %s", dateStr[0:2])
		}

		month, err := strconv.Atoi(dateStr[2:4])
		if err != nil {
			return time.Time{}, fmt.Errorf("invalid month: %s", dateStr[2:4])
		}

		year, err := strconv.Atoi(dateStr[4:6])
		if err != nil {
			return time.Time{}, fmt.Errorf("invalid year: %s", dateStr[4:6])
		}

		// Adjust for century (assuming 20xx for years 00-99)
		year += 2000
		if year > time.Now().Year()+20 {
			year -= 100 // Adjust for 19xx if the result is too far in the future
		}

		return time.Date(year, time.Month(month), day, hour, minute, second, millisecond*1000000, time.UTC), nil
	}

	// If no date provided, use current date
	now := time.Now().UTC()
	return time.Date(now.Year(), now.Month(), now.Day(), hour, minute, second, millisecond*1000000, time.UTC), nil
}

// ParseLatLon parses latitude or longitude from NMEA format to decimal degrees
func ParseLatLon(value string, direction string) (float64, error) {
	if value == "" {
		return 0, fmt.Errorf("empty coordinate value")
	}

	// Parse the value
	coord, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid coordinate value: %s", value)
	}

	// Convert NMEA format (DDMM.MMMM or DDDMM.MMMM) to decimal degrees
	degrees := math.Floor(coord / 100.0)
	minutes := coord - degrees*100.0
	result := degrees + minutes/60.0

	// Apply direction
	if direction == "S" || direction == "W" {
		result = -result
	}

	return result, nil
}

// FormatLatLon formats a decimal degree coordinate to NMEA format
func FormatLatLon(value float64, isLat bool) (string, string) {
	// Determine direction
	var direction string
	if isLat {
		if value >= 0 {
			direction = "N"
		} else {
			direction = "S"
			value = -value
		}
	} else {
		if value >= 0 {
			direction = "E"
		} else {
			direction = "W"
			value = -value
		}
	}

	// Convert to NMEA format (DDMM.MMMM or DDDMM.MMMM)
	degrees := math.Floor(value)
	minutes := (value - degrees) * 60.0

	// Format the string
	var result string
	if isLat {
		result = fmt.Sprintf("%02.0f%09.6f", degrees, minutes)
	} else {
		result = fmt.Sprintf("%03.0f%09.6f", degrees, minutes)
	}

	return result, direction
}

// ValidateChecksum validates the checksum of an NMEA sentence
func ValidateChecksum(sentence string) bool {
	// Find the checksum separator
	checksumPos := strings.LastIndex(sentence, "*")
	if checksumPos == -1 || checksumPos >= len(sentence)-2 {
		return false
	}

	// Extract the checksum
	checksum := sentence[checksumPos+1:]
	data := sentence[1:checksumPos] // Skip the $ at the beginning

	// Calculate the checksum
	calculatedChecksum := CalculateNMEAChecksum(data)

	// Compare checksums (case-insensitive)
	return strings.EqualFold(checksum, calculatedChecksum)
}

// GenerateNMEASentence generates an NMEA sentence with proper checksum
func GenerateNMEASentence(sentenceType string, fields []string) string {
	// Create the sentence without checksum
	parts := []string{"$" + sentenceType}
	parts = append(parts, fields...)
	sentence := strings.Join(parts, ",")

	// Calculate checksum
	checksum := CalculateNMEAChecksum(sentence[1:])

	// Add checksum
	return sentence + "*" + checksum
}
