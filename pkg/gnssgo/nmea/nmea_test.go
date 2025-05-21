package nmea

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseNMEA(t *testing.T) {
	// Test with a valid NMEA sentence
	sentence := "$GPGGA,123519,4807.038,N,01131.000,E,1,08,0.9,545.4,M,46.9,M,,*47"
	result, err := ParseNMEA(sentence)

	assert.NoError(t, err)
	assert.True(t, result.Valid)
	assert.Equal(t, sentence, result.Raw)
	assert.Equal(t, "GPGGA", result.Type)
	assert.Equal(t, "47", result.Checksum)
	assert.Equal(t, []string{"123519", "4807.038", "N", "01131.000", "E", "1", "08", "0.9", "545.4", "M", "46.9", "M", "", ""}, result.Fields)

	// Test with an invalid NMEA sentence (wrong checksum)
	sentence = "$GPGGA,123519,4807.038,N,01131.000,E,1,08,0.9,545.4,M,46.9,M,,*48"
	result, err = ParseNMEA(sentence)

	assert.Error(t, err)
	assert.False(t, result.Valid)
	assert.Contains(t, err.Error(), "checksum mismatch")

	// Test with an invalid NMEA sentence (too short)
	sentence = "$GP"
	result, err = ParseNMEA(sentence)

	assert.Error(t, err)
	assert.False(t, result.Valid)
	assert.Contains(t, err.Error(), "sentence too short")

	// Test with an invalid NMEA sentence (invalid start character)
	sentence = "GPGGA,123519,4807.038,N,01131.000,E,1,08,0.9,545.4,M,46.9,M,,*47"
	result, err = ParseNMEA(sentence)

	assert.Error(t, err)
	assert.False(t, result.Valid)
	assert.Contains(t, err.Error(), "invalid start character")
}

func TestParseGGA(t *testing.T) {
	// Test with a valid GGA sentence
	sentence := "$GPGGA,123519,4807.038,N,01131.000,E,1,08,0.9,545.4,M,46.9,M,,*47"
	result, err := ParseGGA(sentence)

	assert.NoError(t, err)
	assert.Equal(t, "123519", result.Time)
	assert.InDelta(t, 48.1173, result.Latitude, 0.0001)
	assert.Equal(t, "N", result.LatDir)
	assert.InDelta(t, 11.5167, result.Longitude, 0.0001)
	assert.Equal(t, "E", result.LonDir)
	assert.Equal(t, 1, result.Quality)
	assert.Equal(t, 8, result.NumSats)
	assert.InDelta(t, 0.9, result.HDOP, 0.0001)
	assert.InDelta(t, 545.4, result.Altitude, 0.0001)
	assert.Equal(t, "M", result.AltUnit)
	assert.InDelta(t, 46.9, result.GeoidSep, 0.0001)
	assert.Equal(t, "M", result.GeoidUnit)
	assert.InDelta(t, 0.0, result.DGPSAge, 0.0001)
	assert.Equal(t, "", result.DGPSStaID)

	// Test with an invalid GGA sentence (wrong type)
	sentence = "$GPRMC,123519,A,4807.038,N,01131.000,E,022.4,084.4,230394,003.1,W*6A"
	result, err = ParseGGA(sentence)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not a GGA sentence")
}

func TestGetFixQualityName(t *testing.T) {
	assert.Equal(t, "NONE", GetFixQualityName(0))
	assert.Equal(t, "SINGLE", GetFixQualityName(1))
	assert.Equal(t, "DGPS", GetFixQualityName(2))
	assert.Equal(t, "FIX", GetFixQualityName(4))
	assert.Equal(t, "FLOAT", GetFixQualityName(5))
	assert.Equal(t, "UNKNOWN", GetFixQualityName(3)) // Unknown value
}

func TestFindNMEASentences(t *testing.T) {
	data := "$GPGGA,123519,4807.038,N,01131.000,E,1,08,0.9,545.4,M,46.9,M,,*47\r\n$GPRMC,123519,A,4807.038,N,01131.000,E,022.4,084.4,230394,003.1,W*6A"
	sentences := FindNMEASentences(data)

	assert.Equal(t, 2, len(sentences))
	assert.Equal(t, "$GPGGA,123519,4807.038,N,01131.000,E,1,08,0.9,545.4,M,46.9,M,,*47", sentences[0])
	assert.Equal(t, "$GPRMC,123519,A,4807.038,N,01131.000,E,022.4,084.4,230394,003.1,W*6A", sentences[1])
}

func TestNMEAParser(t *testing.T) {
	parser := NewNMEAParser()
	sentence := "$GPGGA,123519,4807.038,N,01131.000,E,1,08,0.9,545.4,M,46.9,M,,*47"

	// Test Parse method
	result, err := parser.Parse(sentence)
	assert.NoError(t, err)
	assert.True(t, result.Valid)

	// Test ParseGGA method
	ggaData, err := parser.ParseGGA(sentence)
	assert.NoError(t, err)
	assert.Equal(t, "123519", ggaData.Time)

	// Test FindNMEASentences method
	data := "$GPGGA,123519,4807.038,N,01131.000,E,1,08,0.9,545.4,M,46.9,M,,*47\r\n$GPRMC,123519,A,4807.038,N,01131.000,E,022.4,084.4,230394,003.1,W*6A"
	sentences := parser.FindNMEASentences(data)
	assert.Equal(t, 2, len(sentences))
}
