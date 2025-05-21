package rtcm_test

import (
	"testing"
	"time"

	"github.com/bramburn/gnssgo/pkg/gnssgo/rtcm"
)

// TestRTCMPreambleDetection tests the detection of RTCM preamble
func TestRTCMPreambleDetection(t *testing.T) {
	// Create a test RTCM message with valid preamble
	data := []byte{
		0xD3, 0x00, 0x13, // Header (preamble + length)
		0x3E, 0xD7, 0xD3, 0x02, 0x02, 0x98, 0x0E, 0xDE, 0xEF, 0x34, 0xB4, 0xBD, 0x62, 0xAC, 0x09, 0x41, 0x98, 0x6F, 0x33, // Data
		0x36, 0x0B, 0x98, // CRC
	}

	// Create a parser
	parser := rtcm.NewRTCMParser()

	// Parse the message
	messages, remaining, err := parser.ParseRTCMMessage(data)
	if err != nil {
		t.Fatalf("Failed to parse RTCM message: %v", err)
	}

	// Check the results
	if len(messages) != 1 {
		t.Fatalf("Expected 1 message, got %d", len(messages))
	}
	if len(remaining) != 0 {
		t.Errorf("Expected 0 remaining bytes, got %d", len(remaining))
	}
	if messages[0].Type != 1005 {
		t.Errorf("Expected message type 1005, got %d", messages[0].Type)
	}
}

// TestRTCMInvalidPreamble tests handling of invalid RTCM preamble
func TestRTCMInvalidPreamble(t *testing.T) {
	// Create a test RTCM message with invalid preamble
	data := []byte{
		0xD4, 0x00, 0x13, // Invalid preamble
		0x3E, 0xD7, 0xD3, 0x02, 0x02, 0x98, 0x0E, 0xDE, 0xEF, 0x34, 0xB4, 0xBD, 0x62, 0xAC, 0x09, 0x41, 0x98, 0x6F, 0x33, // Data
		0x36, 0x0B, 0x98, // CRC
	}

	// Create a parser
	parser := rtcm.NewRTCMParser()

	// Parse the message
	messages, _, err := parser.ParseRTCMMessage(data)
	if err == nil {
		t.Fatalf("Expected error for invalid preamble, got nil")
	}
	if len(messages) != 0 {
		t.Errorf("Expected 0 messages, got %d", len(messages))
	}
}

// TestRTCMIncompleteMessage tests handling of incomplete RTCM message
func TestRTCMIncompleteMessage(t *testing.T) {
	// Create a test RTCM message that's too short
	data := []byte{
		0xD3, 0x00, 0x13, // Header (preamble + length)
		0x3E, 0xD7, // Incomplete data
	}

	// Create a parser
	parser := rtcm.NewRTCMParser()

	// Parse the message
	messages, remaining, err := parser.ParseRTCMMessage(data)
	if err != nil {
		t.Fatalf("Expected no error for incomplete message, got %v", err)
	}
	if len(messages) != 0 {
		t.Errorf("Expected 0 messages, got %d", len(messages))
	}
	if len(remaining) != len(data) {
		t.Errorf("Expected %d remaining bytes, got %d", len(data), len(remaining))
	}
}

// TestRTCMMultipleMessages tests parsing of multiple RTCM messages
func TestRTCMMultipleMessages(t *testing.T) {
	// Create two test RTCM messages
	data := []byte{
		// First message
		0xD3, 0x00, 0x13, // Header (preamble + length)
		0x3E, 0xD7, 0xD3, 0x02, 0x02, 0x98, 0x0E, 0xDE, 0xEF, 0x34, 0xB4, 0xBD, 0x62, 0xAC, 0x09, 0x41, 0x98, 0x6F, 0x33, // Data
		0x36, 0x0B, 0x98, // CRC
		// Second message
		0xD3, 0x00, 0x13, // Header (preamble + length)
		0x4E, 0xD7, 0xD3, 0x02, 0x02, 0x98, 0x0E, 0xDE, 0xEF, 0x34, 0xB4, 0xBD, 0x62, 0xAC, 0x09, 0x41, 0x98, 0x6F, 0x33, // Data
		0x36, 0x0B, 0x98, // CRC
	}

	// Create a parser
	parser := rtcm.NewRTCMParser()

	// Parse the messages
	messages, remaining, err := parser.ParseRTCMMessage(data)
	if err != nil {
		t.Fatalf("Failed to parse RTCM messages: %v", err)
	}

	// Check the results
	if len(messages) != 2 {
		t.Fatalf("Expected 2 messages, got %d", len(messages))
	}
	if len(remaining) != 0 {
		t.Errorf("Expected 0 remaining bytes, got %d", len(remaining))
	}
}

// TestRTCMMessageStats tests the message statistics functionality
func TestRTCMMessageStats(t *testing.T) {
	// Create a test RTCM message
	data := []byte{
		0xD3, 0x00, 0x13, // Header (preamble + length)
		0x3E, 0xD7, 0xD3, 0x02, 0x02, 0x98, 0x0E, 0xDE, 0xEF, 0x34, 0xB4, 0xBD, 0x62, 0xAC, 0x09, 0x41, 0x98, 0x6F, 0x33, // Data
		0x36, 0x0B, 0x98, // CRC
	}

	// Create a parser
	parser := rtcm.NewRTCMParser()

	// Parse the message
	_, _, err := parser.ParseRTCMMessage(data)
	if err != nil {
		t.Fatalf("Failed to parse RTCM message: %v", err)
	}

	// Check the statistics
	stats := parser.GetStats()
	if len(stats) != 1 {
		t.Fatalf("Expected 1 message type in stats, got %d", len(stats))
	}

	// Check the message type
	if _, ok := stats[1005]; !ok {
		t.Errorf("Expected stats for message type 1005")
	}

	// Check the count
	if stats[1005].Count != 1 {
		t.Errorf("Expected count 1, got %d", stats[1005].Count)
	}
}

// TestRTCMMessageTypeDescription tests the message type description functionality
func TestRTCMMessageTypeDescription(t *testing.T) {
	// Test some message type descriptions
	testCases := []struct {
		msgType     int
		description string
	}{
		{1005, "Station Coordinates XYZ"},
		{1006, "Station Coordinates XYZ with Height"},
		{1019, "GPS Ephemeris"},
		{1074, "GPS MSM4"},
		{1084, "GLONASS MSM4"},
		{1094, "Galileo MSM4"},
		{1124, "BeiDou MSM4"},
		{9999, "Unknown (9999)"},
	}

	for _, tc := range testCases {
		desc := rtcm.GetMessageTypeDescription(tc.msgType)
		if desc != tc.description {
			t.Errorf("For message type %d, expected description '%s', got '%s'", tc.msgType, tc.description, desc)
		}
	}
}

// TestValidateCRC tests the CRC validation functionality
func TestValidateCRC(t *testing.T) {
	// Create a test RTCM message with valid CRC
	data := []byte{
		0xD3, 0x00, 0x13, // Header (preamble + length)
		0x3E, 0xD7, 0xD3, 0x02, 0x02, 0x98, 0x0E, 0xDE, 0xEF, 0x34, 0xB4, 0xBD, 0x62, 0xAC, 0x09, 0x41, 0x98, 0x6F, 0x33, // Data
		0x36, 0x0B, 0x98, // CRC
	}

	// Create a message
	msg := rtcm.RTCMMessage{
		Type:      1005,
		Length:    19,
		Data:      data,
		Timestamp: time.Now(),
	}

	// Validate the CRC
	if !rtcm.ValidateCRC(&msg) {
		t.Errorf("CRC validation failed for valid message")
	}

	// Corrupt the CRC
	data[21] = 0x99
	msg.Data = data

	// Validate the CRC again
	if rtcm.ValidateCRC(&msg) {
		t.Errorf("CRC validation passed for invalid message")
	}
}

// TestDecodeRTCMMessage tests the message decoding functionality
func TestDecodeRTCMMessage(t *testing.T) {
	// This is a placeholder test - actual implementation would test specific message types
	// Create a test RTCM message
	data := []byte{
		0xD3, 0x00, 0x13, // Header (preamble + length)
		0x3E, 0xD7, 0xD3, 0x02, 0x02, 0x98, 0x0E, 0xDE, 0xEF, 0x34, 0xB4, 0xBD, 0x62, 0xAC, 0x09, 0x41, 0x98, 0x6F, 0x33, // Data
		0x36, 0x0B, 0x98, // CRC
	}

	// Create a message
	msg := rtcm.RTCMMessage{
		Type:      1005,
		Length:    19,
		Data:      data,
		Timestamp: time.Now(),
	}

	// Decode the message
	_, err := rtcm.DecodeRTCMMessage(&msg)
	if err != nil {
		t.Fatalf("Failed to decode RTCM message: %v", err)
	}
}
