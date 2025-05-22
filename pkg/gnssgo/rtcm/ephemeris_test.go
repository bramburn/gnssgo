package rtcm

import (
	"testing"
	"time"

	"github.com/bramburn/gnssgo/pkg/gnssgo"
)

// TestDecodeGPSEphemeris tests the decoding of GPS ephemeris messages
func TestDecodeGPSEphemeris(t *testing.T) {
	// Create a sample GPS ephemeris message
	msg := &RTCMMessage{
		Type:      RTCM_GPS_EPHEMERIS,
		Length:    100,
		Data:      make([]byte, 100),
		Timestamp: time.Now(),
		StationID: 1234,
	}

	// Set ephemeris fields in the message data
	// Message type (24 bits) and station ID (12 bits) are already set
	pos := 36 // Start after message type and station ID

	// Set satellite ID (6 bits)
	gnssgo.SetBitU(msg.Data, pos, 6, 5) // PRN 5
	pos += 6

	// Set week number (10 bits)
	gnssgo.SetBitU(msg.Data, pos, 10, 2000)
	pos += 10

	// Set SV accuracy (4 bits)
	gnssgo.SetBitU(msg.Data, pos, 4, 2)
	pos += 4

	// Set code on L2 (2 bits)
	gnssgo.SetBitU(msg.Data, pos, 2, 1)
	pos += 2

	// Set IDOT (14 bits)
	gnssgo.SetBits(msg.Data, pos, 14, 100)
	pos += 14

	// Set IODE (8 bits)
	gnssgo.SetBitU(msg.Data, pos, 8, 50)
	pos += 8

	// Set Toc (16 bits)
	gnssgo.SetBitU(msg.Data, pos, 16, 100)
	pos += 16

	// Set Af2 (8 bits)
	gnssgo.SetBits(msg.Data, pos, 8, -5)
	pos += 8

	// Set Af1 (16 bits)
	gnssgo.SetBits(msg.Data, pos, 16, 1000)
	pos += 16

	// Set Af0 (22 bits)
	gnssgo.SetBits(msg.Data, pos, 22, -5000)
	pos += 22

	// Set IODC (10 bits)
	gnssgo.SetBitU(msg.Data, pos, 10, 51)
	pos += 10

	// Set Crs (16 bits)
	gnssgo.SetBits(msg.Data, pos, 16, 1234)
	pos += 16

	// Set DeltaN (16 bits)
	gnssgo.SetBits(msg.Data, pos, 16, 5678)
	pos += 16

	// Set M0 (32 bits)
	gnssgo.SetBits(msg.Data, pos, 32, 12345678)
	pos += 32

	// Set Cuc (16 bits)
	gnssgo.SetBits(msg.Data, pos, 16, 1234)
	pos += 16

	// Set Eccentricity (32 bits)
	gnssgo.SetBitU(msg.Data, pos, 32, 12345678)
	pos += 32

	// Set Cus (16 bits)
	gnssgo.SetBits(msg.Data, pos, 16, 1234)
	pos += 16

	// Set SqrtA (32 bits)
	gnssgo.SetBitU(msg.Data, pos, 32, 12345678)
	pos += 32

	// Set Toe (16 bits)
	gnssgo.SetBitU(msg.Data, pos, 16, 100)
	pos += 16

	// Set Cic (16 bits)
	gnssgo.SetBits(msg.Data, pos, 16, 1234)
	pos += 16

	// Set Omega0 (32 bits)
	gnssgo.SetBits(msg.Data, pos, 32, 12345678)
	pos += 32

	// Set Cis (16 bits)
	gnssgo.SetBits(msg.Data, pos, 16, 1234)
	pos += 16

	// Set Inclination (32 bits)
	gnssgo.SetBits(msg.Data, pos, 32, 12345678)
	pos += 32

	// Set Crc (16 bits)
	gnssgo.SetBits(msg.Data, pos, 16, 1234)
	pos += 16

	// Set Omega (32 bits)
	gnssgo.SetBits(msg.Data, pos, 32, 12345678)
	pos += 32

	// Set OmegaDot (24 bits)
	gnssgo.SetBits(msg.Data, pos, 24, 123456)
	pos += 24

	// Set TGD (8 bits)
	gnssgo.SetBits(msg.Data, pos, 8, 10)
	pos += 8

	// Set SvHealth (6 bits)
	gnssgo.SetBitU(msg.Data, pos, 6, 0)
	pos += 6

	// Set L2PDataFlag (1 bit)
	gnssgo.SetBitU(msg.Data, pos, 1, 1)
	pos += 1

	// Set FitInterval (1 bit)
	gnssgo.SetBitU(msg.Data, pos, 1, 0)
	pos += 1

	// Decode the ephemeris
	eph, err := decodeGPSEphemeris(msg)
	if err != nil {
		t.Fatalf("Failed to decode GPS ephemeris: %v", err)
	}

	// Check the decoded values
	if eph.SatID != 5 {
		t.Errorf("Expected satellite ID 5, got %d", eph.SatID)
	}
	if eph.Week != 2000 {
		t.Errorf("Expected week 2000, got %d", eph.Week)
	}
	if eph.IODE != 50 {
		t.Errorf("Expected IODE 50, got %d", eph.IODE)
	}
	if eph.IODC != 51 {
		t.Errorf("Expected IODC 51, got %d", eph.IODC)
	}
	if eph.L2PDataFlag != true {
		t.Errorf("Expected L2PDataFlag true, got false")
	}
	if eph.FitInterval != false {
		t.Errorf("Expected FitInterval false, got true")
	}
}

// TestDecodeGalileoEphemeris tests the decoding of Galileo ephemeris messages
func TestDecodeGalileoEphemeris(t *testing.T) {
	// Create a sample Galileo ephemeris message
	msg := &RTCMMessage{
		Type:      RTCM_GALILEO_EPHEMERIS,
		Length:    150,
		Data:      make([]byte, 150),
		Timestamp: time.Now(),
		StationID: 1234,
	}

	// Set ephemeris fields in the message data
	// Message type (24 bits) and station ID (12 bits) are already set
	pos := 36 // Start after message type and station ID

	// Set satellite ID (6 bits)
	gnssgo.SetBitU(msg.Data, pos, 6, 5) // PRN 5
	pos += 6

	// Set week number (12 bits)
	gnssgo.SetBitU(msg.Data, pos, 12, 1000)
	pos += 12

	// Set IODNav (10 bits)
	gnssgo.SetBitU(msg.Data, pos, 10, 100)
	pos += 10

	// Set SV health (8 bits)
	gnssgo.SetBitU(msg.Data, pos, 8, 0)
	pos += 8

	// Set BGD_E1E5a (10 bits)
	gnssgo.SetBits(msg.Data, pos, 10, 10)
	pos += 10

	// Set BGD_E1E5b (10 bits)
	gnssgo.SetBits(msg.Data, pos, 10, 20)
	pos += 10

	// Set E5a/E5b/E1B Health Status (2 bits each)
	gnssgo.SetBitU(msg.Data, pos, 2, 0)
	pos += 2
	gnssgo.SetBitU(msg.Data, pos, 2, 0)
	pos += 2
	gnssgo.SetBitU(msg.Data, pos, 2, 0)
	pos += 2

	// Set E5a/E5b/E1B Data Validity Status (1 bit each)
	gnssgo.SetBitU(msg.Data, pos, 1, 1)
	pos += 1
	gnssgo.SetBitU(msg.Data, pos, 1, 1)
	pos += 1
	gnssgo.SetBitU(msg.Data, pos, 1, 1)
	pos += 1

	// Set Toc (14 bits)
	gnssgo.SetBitU(msg.Data, pos, 14, 100)
	pos += 14

	// Set Af0 (31 bits)
	gnssgo.SetBits(msg.Data, pos, 31, -5000)
	pos += 31

	// Set Af1 (21 bits)
	gnssgo.SetBits(msg.Data, pos, 21, 1000)
	pos += 21

	// Set Af2 (6 bits)
	gnssgo.SetBits(msg.Data, pos, 6, -5)
	pos += 6

	// Set remaining orbital parameters (similar to GPS)
	// For brevity, we'll skip setting all the parameters in this test

	// Decode the ephemeris
	eph, err := decodeGalileoEphemeris(msg)
	if err != nil {
		t.Fatalf("Failed to decode Galileo ephemeris: %v", err)
	}

	// Check the decoded values
	if eph.SatID != 5 {
		t.Errorf("Expected satellite ID 5, got %d", eph.SatID)
	}
	if eph.Week != 1000 {
		t.Errorf("Expected week 1000, got %d", eph.Week)
	}
	if eph.IODNav != 100 {
		t.Errorf("Expected IODNav 100, got %d", eph.IODNav)
	}
	if !eph.E5aDataValidity {
		t.Errorf("Expected E5aDataValidity true, got false")
	}
	if !eph.E5bDataValidity {
		t.Errorf("Expected E5bDataValidity true, got false")
	}
	if !eph.E1BDataValidity {
		t.Errorf("Expected E1BDataValidity true, got false")
	}
}
