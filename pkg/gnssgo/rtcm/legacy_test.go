package rtcm

import (
	"encoding/hex"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestDecodeType1001 tests decoding of RTCM message type 1001 (L1-only GPS RTK observables)
func TestDecodeType1001(t *testing.T) {
	// Sample RTCM message type 1001 (hex encoded)
	// This is a synthetic example with the following content:
	// - Message type: 1001
	// - Station ID: 1234
	// - GPS Time of Week: 345678.000 seconds
	// - Synchronous GNSS Flag: 1
	// - No. of GPS Satellite Signals Processed: 4
	// - GPS Satellite ID: 5, 12, 19, 23
	hexData := "D300190103D2000000000000000000000000000000000000000000000000000000"
	data, err := hex.DecodeString(hexData)
	assert.NoError(t, err)

	// Create a message
	msg := RTCMMessage{
		Type:      1001,
		Length:    len(data) - 6, // Subtract header and CRC
		Data:      data,
		Timestamp: time.Now(),
		StationID: 1234,
	}

	// Decode the message
	result, err := DecodeRTCMMessage(&msg)
	assert.NoError(t, err)

	// Verify the result
	obs, ok := result.(*ObservationData)
	assert.True(t, ok)
	assert.NotNil(t, obs)
	// Add more specific assertions based on the expected content
}

// TestDecodeType1002 tests decoding of RTCM message type 1002 (Extended L1-only GPS RTK observables)
func TestDecodeType1002(t *testing.T) {
	// Sample RTCM message type 1002 (hex encoded)
	hexData := "D300280103EA000000000000000000000000000000000000000000000000000000000000000000000000"
	data, err := hex.DecodeString(hexData)
	assert.NoError(t, err)

	// Create a message
	msg := RTCMMessage{
		Type:      1002,
		Length:    len(data) - 6, // Subtract header and CRC
		Data:      data,
		Timestamp: time.Now(),
		StationID: 1234,
	}

	// Decode the message
	result, err := DecodeRTCMMessage(&msg)
	assert.NoError(t, err)

	// Verify the result
	obs, ok := result.(*ObservationData)
	assert.True(t, ok)
	assert.NotNil(t, obs)
	// Add more specific assertions based on the expected content
}

// TestDecodeType1003 tests decoding of RTCM message type 1003 (L1&L2 GPS RTK observables)
func TestDecodeType1003(t *testing.T) {
	// Sample RTCM message type 1003 (hex encoded)
	hexData := "D300190103EB000000000000000000000000000000000000000000000000000000"
	data, err := hex.DecodeString(hexData)
	assert.NoError(t, err)

	// Create a message
	msg := RTCMMessage{
		Type:      1003,
		Length:    len(data) - 6, // Subtract header and CRC
		Data:      data,
		Timestamp: time.Now(),
		StationID: 1234,
	}

	// Decode the message
	result, err := DecodeRTCMMessage(&msg)
	assert.NoError(t, err)

	// Verify the result
	obs, ok := result.(*ObservationData)
	assert.True(t, ok)
	assert.NotNil(t, obs)
	// Add more specific assertions based on the expected content
}

// TestDecodeType1004 tests decoding of RTCM message type 1004 (Extended L1&L2 GPS RTK observables)
func TestDecodeType1004(t *testing.T) {
	// Sample RTCM message type 1004 (hex encoded)
	hexData := "D300380103EC000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"
	data, err := hex.DecodeString(hexData)
	assert.NoError(t, err)

	// Create a message
	msg := RTCMMessage{
		Type:      1004,
		Length:    len(data) - 6, // Subtract header and CRC
		Data:      data,
		Timestamp: time.Now(),
		StationID: 1234,
	}

	// Decode the message
	result, err := DecodeRTCMMessage(&msg)
	assert.NoError(t, err)

	// Verify the result
	obs, ok := result.(*ObservationData)
	assert.True(t, ok)
	assert.NotNil(t, obs)
	// Add more specific assertions based on the expected content
}

// TestDecodeType1009 tests decoding of RTCM message type 1009 (L1-only GLONASS RTK observables)
func TestDecodeType1009(t *testing.T) {
	// Sample RTCM message type 1009 (hex encoded)
	hexData := "D300190103F1000000000000000000000000000000000000000000000000000000"
	data, err := hex.DecodeString(hexData)
	assert.NoError(t, err)

	// Create a message
	msg := RTCMMessage{
		Type:      1009,
		Length:    len(data) - 6, // Subtract header and CRC
		Data:      data,
		Timestamp: time.Now(),
		StationID: 1234,
	}

	// Decode the message
	result, err := DecodeRTCMMessage(&msg)
	assert.NoError(t, err)

	// Verify the result
	obs, ok := result.(*ObservationData)
	assert.True(t, ok)
	assert.NotNil(t, obs)
	// Add more specific assertions based on the expected content
}

// Add more tests for message types 1010, 1011, 1012
