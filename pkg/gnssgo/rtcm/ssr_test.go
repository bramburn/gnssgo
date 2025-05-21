package rtcm

import (
	"testing"
	"time"

	"github.com/bramburn/gnssgo/pkg/gnssgo"
)

// TestDecodeSSRHeader tests the decoding of SSR headers
func TestDecodeSSRHeader(t *testing.T) {
	// Create a sample SSR message with a valid header
	msg := &RTCMMessage{
		Type:      1057, // GPS orbit correction
		Length:    100,
		Data:      make([]byte, 100),
		Timestamp: time.Now(),
		StationID: 1234,
	}

	// Set header fields in the message data
	// Message type (24 bits) and station ID (12 bits) are already set
	pos := 36 // Start after message type and station ID

	// Set epoch time (20 bits)
	gnssgo.SetBitU(msg.Data, pos, 20, 500000)
	pos += 20

	// Set update interval (4 bits)
	gnssgo.SetBitU(msg.Data, pos, 4, 2) // 2 = 5 seconds
	pos += 4

	// Set multiple message flag (1 bit)
	gnssgo.SetBitU(msg.Data, pos, 1, 0)
	pos += 1

	// Set satellite reference datum flag (1 bit)
	gnssgo.SetBitU(msg.Data, pos, 1, 0)
	pos += 1

	// Set IOD SSR indicator (4 bits)
	gnssgo.SetBitU(msg.Data, pos, 4, 3)
	pos += 4

	// Set SSR provider ID (16 bits)
	gnssgo.SetBitU(msg.Data, pos, 16, 5)
	pos += 16

	// Set SSR solution ID (4 bits)
	gnssgo.SetBitU(msg.Data, pos, 4, 1)
	pos += 4

	// Set number of satellites (6 bits)
	gnssgo.SetBitU(msg.Data, pos, 6, 2)
	pos += 6

	// Set satellite IDs (6 bits each)
	gnssgo.SetBitU(msg.Data, pos, 6, 5) // PRN 5
	pos += 6
	gnssgo.SetBitU(msg.Data, pos, 6, 12) // PRN 12
	pos += 6

	// Decode the header
	header, newPos, err := decodeSSRHeader(msg)
	if err != nil {
		t.Fatalf("Failed to decode SSR header: %v", err)
	}

	// Check the decoded values
	if header.MessageType != 1057 {
		t.Errorf("Expected message type 1057, got %d", header.MessageType)
	}
	if header.GNSSID != 0 {
		t.Errorf("Expected GNSS ID 0 (GPS), got %d", header.GNSSID)
	}
	if header.Epoch != 500000 {
		t.Errorf("Expected epoch 500000, got %d", header.Epoch)
	}
	if header.UpdateInterval != 2 {
		t.Errorf("Expected update interval 2, got %d", header.UpdateInterval)
	}
	if header.MultipleMessage {
		t.Errorf("Expected multiple message flag false, got true")
	}
	if header.SatelliteReferenceDatum {
		t.Errorf("Expected satellite reference datum flag false, got true")
	}
	if header.IODSSRIndicator != 3 {
		t.Errorf("Expected IOD SSR indicator 3, got %d", header.IODSSRIndicator)
	}
	if header.SSRProviderID != 5 {
		t.Errorf("Expected SSR provider ID 5, got %d", header.SSRProviderID)
	}
	if header.SolutionID != 1 {
		t.Errorf("Expected SSR solution ID 1, got %d", header.SolutionID)
	}
	if header.NumSatellites != 2 {
		t.Errorf("Expected number of satellites 2, got %d", header.NumSatellites)
	}
	if newPos != pos {
		t.Errorf("Expected position %d, got %d", pos, newPos)
	}
}

// TestDecodeSSRPhaseBias tests the decoding of SSR phase bias messages
func TestDecodeSSRPhaseBias(t *testing.T) {
	// Create a sample SSR phase bias message
	msg := &RTCMMessage{
		Type:      1265, // GPS phase bias
		Length:    150,
		Data:      make([]byte, 150),
		Timestamp: time.Now(),
		StationID: 1234,
	}

	// Set header fields in the message data
	// Message type (24 bits) and station ID (12 bits) are already set
	pos := 36 // Start after message type and station ID

	// Set epoch time (20 bits)
	gnssgo.SetBitU(msg.Data, pos, 20, 500000)
	pos += 20

	// Set update interval (4 bits)
	gnssgo.SetBitU(msg.Data, pos, 4, 2) // 2 = 5 seconds
	pos += 4

	// Set multiple message flag (1 bit)
	gnssgo.SetBitU(msg.Data, pos, 1, 0)
	pos += 1

	// Set satellite reference datum flag (1 bit)
	gnssgo.SetBitU(msg.Data, pos, 1, 0)
	pos += 1

	// Set IOD SSR indicator (4 bits)
	gnssgo.SetBitU(msg.Data, pos, 4, 3)
	pos += 4

	// Set SSR provider ID (16 bits)
	gnssgo.SetBitU(msg.Data, pos, 16, 5)
	pos += 16

	// Set SSR solution ID (4 bits)
	gnssgo.SetBitU(msg.Data, pos, 4, 1)
	pos += 4

	// Set number of satellites (6 bits)
	gnssgo.SetBitU(msg.Data, pos, 6, 1)
	pos += 6

	// Set satellite ID (6 bits)
	gnssgo.SetBitU(msg.Data, pos, 6, 5) // PRN 5
	pos += 6

	// Set satellite data
	// Satellite ID (6 bits)
	gnssgo.SetBitU(msg.Data, pos, 6, 5) // PRN 5
	pos += 6

	// Number of biases (5 bits)
	gnssgo.SetBitU(msg.Data, pos, 5, 2) // 2 biases
	pos += 5

	// Yaw angle (9 bits)
	gnssgo.SetBitU(msg.Data, pos, 9, 45) // 45 degrees
	pos += 9

	// Yaw rate (8 bits)
	gnssgo.SetBits(msg.Data, pos, 8, 10) // 1.0 degrees/s
	pos += 8

	// Bias 1
	// Signal ID (5 bits)
	gnssgo.SetBitU(msg.Data, pos, 5, 1) // L1C
	pos += 5

	// Integer indicator (1 bit)
	gnssgo.SetBitU(msg.Data, pos, 1, 1) // Integer
	pos += 1

	// Wide-lane integer indicator (2 bits)
	gnssgo.SetBitU(msg.Data, pos, 2, 1) // Wide-lane integer
	pos += 2

	// Discontinuity counter (4 bits)
	gnssgo.SetBitU(msg.Data, pos, 4, 3) // Counter = 3
	pos += 4

	// Phase bias (20 bits)
	gnssgo.SetBits(msg.Data, pos, 20, 1000) // 0.1 m
	pos += 20

	// Bias 2
	// Signal ID (5 bits)
	gnssgo.SetBitU(msg.Data, pos, 5, 2) // L2C
	pos += 5

	// Integer indicator (1 bit)
	gnssgo.SetBitU(msg.Data, pos, 1, 0) // Not integer
	pos += 1

	// Wide-lane integer indicator (2 bits)
	gnssgo.SetBitU(msg.Data, pos, 2, 0) // Not wide-lane integer
	pos += 2

	// Discontinuity counter (4 bits)
	gnssgo.SetBitU(msg.Data, pos, 4, 5) // Counter = 5
	pos += 4

	// Phase bias (20 bits)
	gnssgo.SetBits(msg.Data, pos, 20, -500) // -0.05 m
	pos += 20

	// Decode the phase bias message
	phaseBias, err := decodeSSRPhaseBias(msg)
	if err != nil {
		t.Fatalf("Failed to decode SSR phase bias: %v", err)
	}

	// Check the decoded values
	if phaseBias.Header.MessageType != 1265 {
		t.Errorf("Expected message type 1265, got %d", phaseBias.Header.MessageType)
	}
	if phaseBias.Header.NumSatellites != 1 {
		t.Errorf("Expected number of satellites 1, got %d", phaseBias.Header.NumSatellites)
	}
	if len(phaseBias.PhaseBiases) != 1 {
		t.Fatalf("Expected 1 satellite phase bias, got %d", len(phaseBias.PhaseBiases))
	}

	// Check satellite data
	satBias := phaseBias.PhaseBiases[0]
	if satBias.SatID != 5 {
		t.Errorf("Expected satellite ID 5, got %d", satBias.SatID)
	}
	if satBias.NumBiases != 2 {
		t.Errorf("Expected number of biases 2, got %d", satBias.NumBiases)
	}
	if satBias.YawAngle != 45.0*gnssgo.D2R {
		t.Errorf("Expected yaw angle 45 degrees, got %.2f degrees", satBias.YawAngle*gnssgo.R2D)
	}
	if satBias.YawRate != 1.0*gnssgo.D2R {
		t.Errorf("Expected yaw rate 1.0 degrees/s, got %.2f degrees/s", satBias.YawRate*gnssgo.R2D)
	}

	// Check bias 1
	if satBias.SignalIDs[0] != 1 {
		t.Errorf("Expected signal ID 1, got %d", satBias.SignalIDs[0])
	}
	if !satBias.IntegerIndicators[0] {
		t.Errorf("Expected integer indicator true, got false")
	}
	if !satBias.WideLaneIntegerIndicators[0] {
		t.Errorf("Expected wide-lane integer indicator true, got false")
	}
	if satBias.DiscontinuityCounters[0] != 3 {
		t.Errorf("Expected discontinuity counter 3, got %d", satBias.DiscontinuityCounters[0])
	}
	if satBias.PhaseBiases[0] != 0.1 {
		t.Errorf("Expected phase bias 0.1 m, got %.3f m", satBias.PhaseBiases[0])
	}

	// Check bias 2
	if satBias.SignalIDs[1] != 2 {
		t.Errorf("Expected signal ID 2, got %d", satBias.SignalIDs[1])
	}
	if satBias.IntegerIndicators[1] {
		t.Errorf("Expected integer indicator false, got true")
	}
	if satBias.WideLaneIntegerIndicators[1] {
		t.Errorf("Expected wide-lane integer indicator false, got true")
	}
	if satBias.DiscontinuityCounters[1] != 5 {
		t.Errorf("Expected discontinuity counter 5, got %d", satBias.DiscontinuityCounters[1])
	}
	if satBias.PhaseBiases[1] != -0.05 {
		t.Errorf("Expected phase bias -0.05 m, got %.3f m", satBias.PhaseBiases[1])
	}
}
