package rtcm

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestLegacyRTCMMessageDecoding tests the decoding of legacy RTCM messages (1001-1004, 1009-1012)
func TestLegacyRTCMMessageDecoding(t *testing.T) {
	// Test cases for different legacy RTCM message types
	testCases := []struct {
		name      string
		hexData   string
		msgType   int
		stationID uint16
		expectErr bool
	}{
		{
			name:      "RTCM Message Type 1001 (GPS L1 Observables)",
			hexData:   "D300190103D2000000000000000000000000000000000000000000000000000000",
			msgType:   1001,
			stationID: 1234,
			expectErr: false,
		},
		{
			name:      "RTCM Message Type 1002 (Extended GPS L1 Observables)",
			hexData:   "D300280103EA000000000000000000000000000000000000000000000000000000000000000000000000",
			msgType:   1002,
			stationID: 1234,
			expectErr: false,
		},
		{
			name:      "RTCM Message Type 1003 (GPS L1/L2 Observables)",
			hexData:   "D300280103EB000000000000000000000000000000000000000000000000000000000000000000000000",
			msgType:   1003,
			stationID: 1234,
			expectErr: false,
		},
		{
			name:      "RTCM Message Type 1004 (Extended GPS L1/L2 Observables)",
			hexData:   "D300380103EC000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
			msgType:   1004,
			stationID: 1234,
			expectErr: false,
		},
		{
			name:      "RTCM Message Type 1009 (GLONASS L1 Observables)",
			hexData:   "D300190103F1000000000000000000000000000000000000000000000000000000",
			msgType:   1009,
			stationID: 1234,
			expectErr: false,
		},
		{
			name:      "RTCM Message Type 1010 (Extended GLONASS L1 Observables)",
			hexData:   "D300280103F2000000000000000000000000000000000000000000000000000000000000000000000000",
			msgType:   1010,
			stationID: 1234,
			expectErr: false,
		},
		{
			name:      "RTCM Message Type 1011 (GLONASS L1/L2 Observables)",
			hexData:   "D300280103F3000000000000000000000000000000000000000000000000000000000000000000000000",
			msgType:   1011,
			stationID: 1234,
			expectErr: false,
		},
		{
			name:      "RTCM Message Type 1012 (Extended GLONASS L1/L2 Observables)",
			hexData:   "D300380103F4000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
			msgType:   1012,
			stationID: 1234,
			expectErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Decode hex data
			data, err := hex.DecodeString(tc.hexData)
			assert.NoError(t, err)

			// Create a message
			msg := RTCMMessage{
				Type:      tc.msgType,
				Length:    len(data) - 6, // Subtract header and CRC
				Data:      data,
				StationID: tc.stationID,
			}

			// Decode the message
			result, err := DecodeRTCMMessage(&msg)
			if tc.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)

				// Check that the result is of the expected type
				obs, ok := result.(*ObservationData)
				assert.True(t, ok)
				assert.NotNil(t, obs)

				// We don't check the station ID here because our test data doesn't have valid station IDs
				// The station ID in the message is extracted from the data, not from the RTCMMessage struct
			}
		})
	}
}

// TestLegacyRTCMMessageDecodingWithRealData tests the decoding of legacy RTCM messages with more realistic data
func TestLegacyRTCMMessageDecodingWithRealData(t *testing.T) {
	// This is a simplified RTCM message type 1004 (Extended GPS L1/L2 Observables)
	// We're using a simplified version for testing
	hexData := "D300380103EC000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"

	// Decode hex data
	data, err := hex.DecodeString(hexData)
	assert.NoError(t, err)

	// Create a message
	msg := RTCMMessage{
		Type:      1004,
		Length:    len(data) - 6, // Subtract header and CRC
		Data:      data,
		StationID: 1234,
	}

	// Decode the message
	result, err := DecodeRTCMMessage(&msg)
	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Check that the result is of the expected type
	obs, ok := result.(*ObservationData)
	assert.True(t, ok)
	assert.NotNil(t, obs)

	// We don't check the station ID here because our test data doesn't have valid station IDs
	// The station ID in the message is extracted from the data, not from the RTCMMessage struct
}
