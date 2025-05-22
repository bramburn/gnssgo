package rtcm

import (
	"testing"
	"time"

	"github.com/bramburn/gnssgo/pkg/gnssgo"
)

// TestDecodeMSMHeader tests the decoding of MSM headers
func TestDecodeMSMHeader(t *testing.T) {
	// Create a sample MSM message
	msg := &RTCMMessage{
		Type:      MSM_GPS_RANGE_START + MSM7 - 1, // GPS MSM7
		Length:    100,
		Data:      make([]byte, 100),
		Timestamp: time.Now(),
		StationID: 1234,
	}

	// Set header fields in the message data
	// Message type (24 bits) and station ID (12 bits) are already set
	pos := 36 // Start after message type and station ID

	// Set epoch time (30 bits for GPS)
	gnssgo.SetBitU(msg.Data, pos, 30, 500000)
	pos += 30

	// Set multiple message flag (1 bit)
	gnssgo.SetBitU(msg.Data, pos, 1, 0)
	pos += 1

	// Set issue of data station (3 bits)
	gnssgo.SetBitU(msg.Data, pos, 3, 5)
	pos += 3

	// Set clock steering indicator (2 bits)
	gnssgo.SetBitU(msg.Data, pos, 2, 2)
	pos += 2

	// Set external clock indicator (2 bits)
	gnssgo.SetBitU(msg.Data, pos, 2, 1)
	pos += 2

	// Set smoothing indicator (1 bit)
	gnssgo.SetBitU(msg.Data, pos, 1, 1)
	pos += 1

	// Set smoothing interval (3 bits)
	gnssgo.SetBitU(msg.Data, pos, 3, 3)
	pos += 3

	// Set satellite mask (64 bits)
	// Set bits for PRN 1, 5, and 10
	gnssgo.SetBitU(msg.Data, pos, 32, 0x00000421) // Bits 0, 5, 10 set in first 32 bits
	pos += 32
	gnssgo.SetBitU(msg.Data, pos, 32, 0x00000000) // No bits set in second 32 bits
	pos += 32

	// Set signal mask (32 bits)
	// Set bits for L1 C/A, L2P, and L5I
	gnssgo.SetBitU(msg.Data, pos, 32, 0x00000205) // Bits 0, 2, 9 set
	pos += 32

	// Set cell mask (3 satellites * 3 signals = 9 bits)
	// Set all combinations except PRN10-L5I
	gnssgo.SetBitU(msg.Data, pos, 9, 0x000001FD) // All bits set except bit 8
	pos += 9

	// Decode the header
	header, newPos, err := decodeMSMHeader(msg, gnssgo.SYS_GPS)
	if err != nil {
		t.Fatalf("Failed to decode MSM header: %v", err)
	}

	// Check the decoded values
	if header.MessageType != MSM_GPS_RANGE_START+MSM7-1 {
		t.Errorf("Expected message type %d, got %d", MSM_GPS_RANGE_START+MSM7-1, header.MessageType)
	}
	if header.StationID != 1234 {
		t.Errorf("Expected station ID 1234, got %d", header.StationID)
	}
	if header.GNSSID != 0 {
		t.Errorf("Expected GNSS ID 0 (GPS), got %d", header.GNSSID)
	}
	if header.Epoch != 500000 {
		t.Errorf("Expected epoch 500000, got %d", header.Epoch)
	}
	if header.MultipleMessage {
		t.Errorf("Expected multiple message flag false, got true")
	}
	if header.IssueOfDataStation != 5 {
		t.Errorf("Expected IODS 5, got %d", header.IssueOfDataStation)
	}
	if header.ClockSteeringIndicator != 2 {
		t.Errorf("Expected clock steering indicator 2, got %d", header.ClockSteeringIndicator)
	}
	if header.ExternalClockIndicator != 1 {
		t.Errorf("Expected external clock indicator 1, got %d", header.ExternalClockIndicator)
	}
	if !header.SmoothingIndicator {
		t.Errorf("Expected smoothing indicator true, got false")
	}
	if header.SmoothingInterval != 3 {
		t.Errorf("Expected smoothing interval 3, got %d", header.SmoothingInterval)
	}
	if header.NumSatellites != 3 {
		t.Errorf("Expected 3 satellites, got %d", header.NumSatellites)
	}
	if header.NumSignals != 3 {
		t.Errorf("Expected 3 signals, got %d", header.NumSignals)
	}
	if header.NumCells != 8 {
		t.Errorf("Expected 8 cells, got %d", header.NumCells)
	}
	if newPos != pos {
		t.Errorf("Expected position %d, got %d", pos, newPos)
	}
}

// TestDecodeMSMMessage tests the decoding of MSM messages
func TestDecodeMSMMessage(t *testing.T) {
	// Create a sample MSM message (MSM4 - GPS)
	msg := &RTCMMessage{
		Type:      MSM_GPS_RANGE_START + MSM4 - 1, // GPS MSM4
		Length:    150,
		Data:      make([]byte, 150),
		Timestamp: time.Now(),
		StationID: 1234,
	}

	// Set header fields in the message data
	// Message type (24 bits) and station ID (12 bits) are already set
	pos := 36 // Start after message type and station ID

	// Set epoch time (30 bits for GPS)
	gnssgo.SetBitU(msg.Data, pos, 30, 500000)
	pos += 30

	// Set multiple message flag (1 bit)
	gnssgo.SetBitU(msg.Data, pos, 1, 0)
	pos += 1

	// Set issue of data station (3 bits)
	gnssgo.SetBitU(msg.Data, pos, 3, 5)
	pos += 3

	// Set clock steering indicator (2 bits)
	gnssgo.SetBitU(msg.Data, pos, 2, 2)
	pos += 2

	// Set external clock indicator (2 bits)
	gnssgo.SetBitU(msg.Data, pos, 2, 1)
	pos += 2

	// Set smoothing indicator (1 bit)
	gnssgo.SetBitU(msg.Data, pos, 1, 1)
	pos += 1

	// Set smoothing interval (3 bits)
	gnssgo.SetBitU(msg.Data, pos, 3, 3)
	pos += 3

	// Set satellite mask (64 bits)
	// Set bits for PRN 1 and 5
	gnssgo.SetBitU(msg.Data, pos, 32, 0x00000021) // Bits 0 and 5 set in first 32 bits
	pos += 32
	gnssgo.SetBitU(msg.Data, pos, 32, 0x00000000) // No bits set in second 32 bits
	pos += 32

	// Set signal mask (32 bits)
	// Set bits for L1 C/A and L2P
	gnssgo.SetBitU(msg.Data, pos, 32, 0x00000005) // Bits 0 and 2 set
	pos += 32

	// Set cell mask (2 satellites * 2 signals = 4 bits)
	// Set all combinations
	gnssgo.SetBitU(msg.Data, pos, 4, 0x0000000F) // All bits set
	pos += 4

	// Set satellite data
	// For MSM4, we need to set the range integer for each satellite
	gnssgo.SetBitU(msg.Data, pos, 8, 100) // PRN 1 range integer
	pos += 8
	gnssgo.SetBitU(msg.Data, pos, 8, 150) // PRN 5 range integer
	pos += 8

	// Set range modulo for each satellite
	gnssgo.SetBitU(msg.Data, pos, 15, 1000) // PRN 1 range modulo
	pos += 15
	gnssgo.SetBitU(msg.Data, pos, 15, 2000) // PRN 5 range modulo
	pos += 15

	// Set pseudoranges for each cell
	gnssgo.SetBits(msg.Data, pos, 20, 5000) // PRN 1, L1 C/A
	pos += 20
	gnssgo.SetBits(msg.Data, pos, 20, 5100) // PRN 1, L2P
	pos += 20
	gnssgo.SetBits(msg.Data, pos, 20, 5200) // PRN 5, L1 C/A
	pos += 20
	gnssgo.SetBits(msg.Data, pos, 20, 5300) // PRN 5, L2P
	pos += 20

	// Set phase ranges for each cell
	gnssgo.SetBits(msg.Data, pos, 24, 6000) // PRN 1, L1 C/A
	pos += 24
	gnssgo.SetBits(msg.Data, pos, 24, 6100) // PRN 1, L2P
	pos += 24
	gnssgo.SetBits(msg.Data, pos, 24, 6200) // PRN 5, L1 C/A
	pos += 24
	gnssgo.SetBits(msg.Data, pos, 24, 6300) // PRN 5, L2P
	pos += 24

	// Set lock time indicators for each cell
	gnssgo.SetBitU(msg.Data, pos, 4, 5) // PRN 1, L1 C/A
	pos += 4
	gnssgo.SetBitU(msg.Data, pos, 4, 6) // PRN 1, L2P
	pos += 4
	gnssgo.SetBitU(msg.Data, pos, 4, 7) // PRN 5, L1 C/A
	pos += 4
	gnssgo.SetBitU(msg.Data, pos, 4, 8) // PRN 5, L2P
	pos += 4

	// Set half-cycle ambiguity indicators for each cell
	gnssgo.SetBitU(msg.Data, pos, 1, 0) // PRN 1, L1 C/A
	pos += 1
	gnssgo.SetBitU(msg.Data, pos, 1, 1) // PRN 1, L2P
	pos += 1
	gnssgo.SetBitU(msg.Data, pos, 1, 0) // PRN 5, L1 C/A
	pos += 1
	gnssgo.SetBitU(msg.Data, pos, 1, 1) // PRN 5, L2P
	pos += 1

	// Set CNR for each cell
	gnssgo.SetBitU(msg.Data, pos, 6, 40) // PRN 1, L1 C/A
	pos += 6
	gnssgo.SetBitU(msg.Data, pos, 6, 42) // PRN 1, L2P
	pos += 6
	gnssgo.SetBitU(msg.Data, pos, 6, 44) // PRN 5, L1 C/A
	pos += 6
	gnssgo.SetBitU(msg.Data, pos, 6, 46) // PRN 5, L2P
	pos += 6

	// Decode the MSM message
	msm, err := decodeMSMMessage(msg, gnssgo.SYS_GPS)
	if err != nil {
		t.Fatalf("Failed to decode MSM message: %v", err)
	}

	// Check the decoded values
	if msm.Header.NumSatellites != 2 {
		t.Errorf("Expected 2 satellites, got %d", msm.Header.NumSatellites)
	}
	if msm.Header.NumSignals != 2 {
		t.Errorf("Expected 2 signals, got %d", msm.Header.NumSignals)
	}
	if msm.Header.NumCells != 4 {
		t.Errorf("Expected 4 cells, got %d", msm.Header.NumCells)
	}

	// Check satellite data
	if len(msm.Satellites) != 2 {
		t.Fatalf("Expected 2 satellites, got %d", len(msm.Satellites))
	}
	if msm.Satellites[0].ID != 1 {
		t.Errorf("Expected satellite ID 1, got %d", msm.Satellites[0].ID)
	}
	if msm.Satellites[1].ID != 6 {
		t.Errorf("Expected satellite ID 6, got %d", msm.Satellites[1].ID)
	}
	if msm.Satellites[0].RangeInteger != 100 {
		t.Errorf("Expected range integer 100, got %d", msm.Satellites[0].RangeInteger)
	}
	if msm.Satellites[1].RangeInteger != 150 {
		t.Errorf("Expected range integer 150, got %d", msm.Satellites[1].RangeInteger)
	}

	// Check signal data
	if len(msm.Signals) != 4 {
		t.Fatalf("Expected 4 signals, got %d", len(msm.Signals))
	}
}
