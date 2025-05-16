package top708

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestNMEAParserParse tests the Parse method of NMEAParser
func TestNMEAParserParse(t *testing.T) {
	// Create a new NMEA parser
	parser := NewNMEAParser()

	// Test with a valid NMEA sentence
	sentence := "$GPGGA,123519,4807.038,N,01131.000,E,1,08,0.9,545.4,M,46.9,M,,*47"
	result := parser.Parse(sentence)

	// Verify the result
	assert.True(t, result.Valid)
	assert.Equal(t, sentence, result.Raw)
	assert.Equal(t, "GPGGA", result.Type)
	assert.Equal(t, "47", result.Checksum)
	assert.Equal(t, []string{"123519", "4807.038", "N", "01131.000", "E", "1", "08", "0.9", "545.4", "M", "46.9", "M", "", ""}, result.Fields)

	// Test with an invalid NMEA sentence (no $ prefix)
	sentence = "GPGGA,123519,4807.038,N,01131.000,E,1,08,0.9,545.4,M,46.9,M,,*47"
	result = parser.Parse(sentence)

	// Verify the result
	assert.False(t, result.Valid)
	assert.Equal(t, sentence, result.Raw)

	// Test with an invalid NMEA sentence (no checksum)
	sentence = "$GPGGA,123519,4807.038,N,01131.000,E,1,08,0.9,545.4,M,46.9,M,,"
	result = parser.Parse(sentence)

	// Verify the result
	assert.False(t, result.Valid)
	assert.Equal(t, sentence, result.Raw)

	// Test with an invalid NMEA sentence (invalid checksum)
	sentence = "$GPGGA,123519,4807.038,N,01131.000,E,1,08,0.9,545.4,M,46.9,M,,*48"
	result = parser.Parse(sentence)

	// Verify the result
	assert.False(t, result.Valid)
	assert.Equal(t, sentence, result.Raw)
}

// TestNMEAParserCalculateChecksum tests the calculateChecksum method of NMEAParser
func TestNMEAParserCalculateChecksum(t *testing.T) {
	// Create a new NMEA parser
	parser := NewNMEAParser()

	// Test with a valid NMEA sentence
	data := "GPGGA,123519,4807.038,N,01131.000,E,1,08,0.9,545.4,M,46.9,M,,"
	checksum := parser.calculateChecksum(data)

	// Verify the result
	assert.Equal(t, "47", checksum)
}

// TestRTCMParserParse tests the Parse method of RTCMParser
func TestRTCMParserParse(t *testing.T) {
	// Create a new RTCM parser
	parser := NewRTCMParser()

	// Test with a valid RTCM message
	// RTCM message with preamble 0xD3, length 10, message ID 1005
	data := []byte{0xD3, 0x00, 0x0A, 0xF8, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	result := parser.Parse(data)

	// Verify the result
	assert.True(t, result.Valid)
	assert.Equal(t, data, result.Raw)
	assert.Equal(t, 10, result.Length)
	assert.Equal(t, 62, result.MessageID) // 0xF8 >> 2 = 62

	// Test with an invalid RTCM message (no preamble)
	data = []byte{0x00, 0x00, 0x0A, 0xF8, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	result = parser.Parse(data)

	// Verify the result
	assert.False(t, result.Valid)
	assert.Equal(t, data, result.Raw)

	// Test with an invalid RTCM message (too short)
	data = []byte{0xD3, 0x00}
	result = parser.Parse(data)

	// Verify the result
	assert.False(t, result.Valid)
	assert.Equal(t, data, result.Raw)

	// Test with an invalid RTCM message (incomplete)
	data = []byte{0xD3, 0x00, 0x0A, 0xF8}
	result = parser.Parse(data)

	// Verify the result
	assert.False(t, result.Valid)
	assert.Equal(t, data, result.Raw)
}

// TestUBXParserParse tests the Parse method of UBXParser
func TestUBXParserParse(t *testing.T) {
	// Create a new UBX parser
	parser := NewUBXParser()

	// Test with a valid UBX message
	// UBX message with header 0xB5 0x62, class 0x01, ID 0x02, length 0x1C, payload and checksum
	data := []byte{0xB5, 0x62, 0x01, 0x02, 0x1C, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x1D, 0x1D}
	result := parser.Parse(data)

	// Verify the result
	assert.Equal(t, data, result.Raw)
	assert.Equal(t, byte(0x01), result.Class)
	assert.Equal(t, byte(0x02), result.ID)
	assert.Equal(t, 28, result.Length) // 0x1C = 28

	// Test with an invalid UBX message (no header)
	data = []byte{0x00, 0x00, 0x01, 0x02, 0x1C, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x1D, 0x1D}
	result = parser.Parse(data)

	// Verify the result
	assert.False(t, result.Valid)
	assert.Equal(t, data, result.Raw)

	// Test with an invalid UBX message (too short)
	data = []byte{0xB5, 0x62, 0x01, 0x02, 0x1C, 0x00, 0x00}
	result = parser.Parse(data)

	// Verify the result
	assert.False(t, result.Valid)
	assert.Equal(t, data, result.Raw)
}

// TestUBXParserCalculateChecksum tests the calculateChecksum method of UBXParser
func TestUBXParserCalculateChecksum(t *testing.T) {
	// Create a new UBX parser
	parser := NewUBXParser()

	// Test with a simple UBX message for easier verification
	data := []byte{0x01, 0x02, 0x00, 0x00}
	checksum := parser.calculateChecksum(data)

	// Manually calculate the checksum
	var ck_a, ck_b byte
	for _, b := range data {
		ck_a = ck_a + b
		ck_b = ck_b + ck_a
	}
	expected := uint16(ck_a) | (uint16(ck_b) << 8)

	// Verify the result
	assert.Equal(t, expected, checksum)
}
