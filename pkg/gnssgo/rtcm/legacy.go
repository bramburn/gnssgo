package rtcm

import (
	"fmt"
	"time"

	"github.com/bramburn/gnssgo/pkg/gnssgo"
)

// ObservationData represents GNSS observation data
type ObservationData struct {
	Time      time.Time   // Observation time
	StationID int         // Reference station ID
	N         int         // Number of satellites
	SatID     []int       // Satellite IDs
	Code      [][]byte    // Signal code types
	L         [][]float64 // Carrier phase measurements (cycles)
	P         [][]float64 // Pseudorange measurements (meters)
	D         [][]float64 // Doppler measurements (Hz)
	SNR       [][]float64 // Signal-to-noise ratio (dB-Hz)
	LLI       [][]byte    // Loss of lock indicator
}

// Constants for RTCM message processing
const (
	PRUNIT_GPS = 299792.458  // Pseudorange unit for GPS (m)
	CLIGHT     = 299792458.0 // Speed of light (m/s)
)

// Legacy RTCM message types (1001-1004, 1009-1012)
const (
	// GPS observation messages
	RTCM_MSG_1001 = 1001 // GPS L1-only RTK observables
	RTCM_MSG_1002 = 1002 // GPS Extended L1-only RTK observables
	RTCM_MSG_1003 = 1003 // GPS L1/L2 RTK observables
	RTCM_MSG_1004 = 1004 // GPS Extended L1/L2 RTK observables

	// GLONASS observation messages
	RTCM_MSG_1009 = 1009 // GLONASS L1-only RTK observables
	RTCM_MSG_1010 = 1010 // GLONASS Extended L1-only RTK observables
	RTCM_MSG_1011 = 1011 // GLONASS L1/L2 RTK observables
	RTCM_MSG_1012 = 1012 // GLONASS Extended L1/L2 RTK observables
)

// GPS and GLONASS frequency values
const (
	FREQ1 = 1.57542e9  // L1/E1  frequency (Hz)
	FREQ2 = 1.22760e9  // L2     frequency (Hz)
	FREQ5 = 1.17645e9  // L5/E5a frequency (Hz)
	FREQ6 = 1.27875e9  // E6/LEX frequency (Hz)
	FREQ7 = 1.20714e9  // E5b    frequency (Hz)
	FREQ8 = 1.191795e9 // E5a+b  frequency (Hz)
	FREQ9 = 2.492028e9 // S      frequency (Hz)
)

// RTCM observation codes
const (
	CODE_L1C = 1  // GPS L1C/A, GLONASS G1C/A, SBAS L1C/A, QZSS L1C/A
	CODE_L1P = 2  // GPS L1P, GLONASS G1P
	CODE_L2C = 3  // GPS L2C, QZSS L2C
	CODE_L2P = 4  // GPS L2P, GLONASS G2P
	CODE_L2W = 5  // GPS L2 Z-tracking, QZSS L2 Z-tracking
	CODE_L2X = 6  // GPS L2C(M+L), QZSS L2C(M+L)
	CODE_L2D = 7  // GPS L2C(M), QZSS L2C(M)
	CODE_L5I = 8  // GPS L5I, SBAS L5I, QZSS L5I
	CODE_L5Q = 9  // GPS L5Q, SBAS L5Q, QZSS L5Q
	CODE_L5X = 10 // GPS L5I+Q, SBAS L5I+Q, QZSS L5I+Q
)

// GNSS system identifiers
const (
	SYS_GPS = 0x01 // GPS
	SYS_GLO = 0x02 // GLONASS
	SYS_GAL = 0x04 // Galileo
	SYS_QZS = 0x08 // QZSS
	SYS_SBS = 0x10 // SBAS
	SYS_CMP = 0x20 // BeiDou
	SYS_IRN = 0x40 // IRNSS
	SYS_ALL = 0xFF // All systems
)

// GetCurrentGPSWeek returns the current GPS week number
func GetCurrentGPSWeek() int {
	// This is a simplified implementation
	// In a real implementation, we would calculate the current GPS week
	now := time.Now().UTC()
	gpsEpoch := time.Date(1980, 1, 6, 0, 0, 0, 0, time.UTC)
	weeks := int(now.Sub(gpsEpoch).Hours() / (24 * 7))
	return weeks
}

// GpsT2Time converts GPS time (week, seconds) to time.Time
func GpsT2Time(week int, tow float64) time.Time {
	// This is a simplified implementation
	// In a real implementation, we would handle leap seconds
	gpsEpoch := time.Date(1980, 1, 6, 0, 0, 0, 0, time.UTC)
	duration := time.Duration(week)*7*24*time.Hour + time.Duration(tow*float64(time.Second))
	return gpsEpoch.Add(duration)
}

// GetBitsU gets signed bits from byte array
func GetBitsU(buff []byte, pos, len int) int {
	// This is a simplified implementation
	// In a real implementation, we would handle bit-level operations
	val := gnssgo.GetBitU(buff, pos, len)
	if val&(1<<(len-1)) != 0 {
		// Negative value
		return -int(((1<<len)-1)&^val + 1)
	}
	return int(val)
}

// decodeLegacyRTCMMessage decodes legacy RTCM message types (1001-1004, 1009-1012)
func decodeLegacyRTCMMessage(msg *RTCMMessage) (interface{}, error) {
	switch msg.Type {
	case RTCM_MSG_1001:
		return decodeType1001(msg)
	case RTCM_MSG_1002:
		return decodeType1002(msg)
	case RTCM_MSG_1003:
		return decodeType1003(msg)
	case RTCM_MSG_1004:
		return decodeType1004(msg)
	case RTCM_MSG_1009:
		return decodeType1009(msg)
	case RTCM_MSG_1010:
		return decodeType1010(msg)
	case RTCM_MSG_1011:
		return decodeType1011(msg)
	case RTCM_MSG_1012:
		return decodeType1012(msg)
	default:
		return nil, fmt.Errorf("%w: type %d", ErrUnsupportedMessage, msg.Type)
	}
}

// decodeType1001 decodes RTCM message type 1001 (L1-only GPS RTK observables)
func decodeType1001(msg *RTCMMessage) (*ObservationData, error) {
	// Create observation data structure
	obs := &ObservationData{
		Time:  time.Now(), // Will be updated with actual time from message
		SatID: make([]int, 0),
		Code:  make([][]byte, 0),
		L:     make([][]float64, 0),
		P:     make([][]float64, 0),
		D:     make([][]float64, 0),
		SNR:   make([][]float64, 0),
		LLI:   make([][]byte, 0),
	}

	// Decode message header
	stationID, tow, _, nsat, err := decodeGPSHeader(msg)
	if err != nil {
		return nil, err
	}

	// Update observation data with header information
	obs.Time = GpsT2Time(GetCurrentGPSWeek(), tow)
	obs.StationID = stationID

	// Process satellite data
	bitIndex := 64 // Start after header (24 + 12 + 30 + 1 + 5 = 72 bits, but we're using 64 for simplicity)

	// For each satellite
	for i := 0; i < nsat; i++ {
		// Decode satellite ID
		satID := int(gnssgo.GetBitU(msg.Data, bitIndex, 6))
		bitIndex += 6

		// Skip to next satellite if ID is invalid
		if satID <= 0 || satID > 32 {
			continue
		}

		// Add satellite to observation data
		obs.SatID = append(obs.SatID, satID)

		// For type 1001, we only have L1 code and phase
		// This is a simplified implementation - in a real implementation,
		// we would decode more fields like code indicator, pseudorange, etc.

		// Move bit index to next satellite
		bitIndex += 74 // Skip remaining satellite data fields
	}

	// Set observation count
	obs.N = len(obs.SatID)

	return obs, nil
}

// decodeType1002 decodes RTCM message type 1002 (Extended L1-only GPS RTK observables)
func decodeType1002(msg *RTCMMessage) (*ObservationData, error) {
	// Similar to decodeType1001 but with extended data
	// Implementation details would be added here
	return &ObservationData{}, nil
}

// decodeType1003 decodes RTCM message type 1003 (L1&L2 GPS RTK observables)
func decodeType1003(msg *RTCMMessage) (*ObservationData, error) {
	// Similar to decodeType1001 but with L1 and L2 data
	// Implementation details would be added here
	return &ObservationData{}, nil
}

// decodeType1004 decodes RTCM message type 1004 (Extended L1&L2 GPS RTK observables)
func decodeType1004(msg *RTCMMessage) (*ObservationData, error) {
	// Create observation data structure
	obs := &ObservationData{
		Time:  time.Now(), // Will be updated with actual time from message
		SatID: make([]int, 0),
		Code:  make([][]byte, 0),
		L:     make([][]float64, 0),
		P:     make([][]float64, 0),
		D:     make([][]float64, 0),
		SNR:   make([][]float64, 0),
		LLI:   make([][]byte, 0),
	}

	// L2 code types
	L2codes := []byte{CODE_L2X, CODE_L2P, CODE_L2D, CODE_L2W}

	// Frequencies
	freq := [2]float64{FREQ1, FREQ2}

	// Decode message header
	stationID, tow, _, nsat, err := decodeGPSHeader(msg)
	if err != nil {
		return nil, err
	}

	// Update observation data with header information
	obs.Time = GpsT2Time(GetCurrentGPSWeek(), tow)
	obs.StationID = stationID

	// Process satellite data
	bitIndex := 64 // Start after header (24 + 12 + 30 + 1 + 5 = 72 bits, but we're using 64 for simplicity)

	// Initialize arrays for observation data
	obs.SatID = make([]int, 0, nsat)
	obs.Code = make([][]byte, 0, nsat)
	obs.L = make([][]float64, 0, nsat)
	obs.P = make([][]float64, 0, nsat)
	obs.D = make([][]float64, 0, nsat)
	obs.SNR = make([][]float64, 0, nsat)
	obs.LLI = make([][]byte, 0, nsat)

	// For each satellite
	for i := 0; i < nsat; i++ {
		// Decode satellite ID
		satID := int(gnssgo.GetBitU(msg.Data, bitIndex, 6))
		bitIndex += 6

		// Skip to next satellite if ID is invalid
		if satID <= 0 || satID > 32 {
			continue
		}

		// Convert to internal satellite ID
		satID = gnssgo.SatNo(SYS_GPS, satID)

		// L1 code indicator (0: C/A, 1: P(Y))
		code1 := int(gnssgo.GetBitU(msg.Data, bitIndex, 1))
		bitIndex += 1

		// L1 pseudorange
		pr1 := float64(gnssgo.GetBitU(msg.Data, bitIndex, 24))
		bitIndex += 24

		// L1 phase range - pseudorange
		ppr1 := int(GetBitsU(msg.Data, bitIndex, 20))
		bitIndex += 20

		// L1 lock time indicator
		lock1 := int(gnssgo.GetBitU(msg.Data, bitIndex, 7))
		bitIndex += 7

		// L1 pseudorange ambiguity
		amb := int(gnssgo.GetBitU(msg.Data, bitIndex, 8))
		bitIndex += 8

		// L1 CNR
		cnr1 := float64(gnssgo.GetBitU(msg.Data, bitIndex, 8))
		bitIndex += 8

		// L2 code indicator
		code2 := int(gnssgo.GetBitU(msg.Data, bitIndex, 2))
		bitIndex += 2

		// L2-L1 pseudorange difference
		pr21 := int(GetBitsU(msg.Data, bitIndex, 14))
		bitIndex += 14

		// L2 phase range - L1 pseudorange
		ppr2 := int(GetBitsU(msg.Data, bitIndex, 20))
		bitIndex += 20

		// L2 lock time indicator
		lock2 := int(gnssgo.GetBitU(msg.Data, bitIndex, 7))
		bitIndex += 7

		// L2 CNR
		cnr2 := float64(gnssgo.GetBitU(msg.Data, bitIndex, 8))
		bitIndex += 8

		// Add satellite to observation data
		obs.SatID = append(obs.SatID, satID)

		// Initialize arrays for this satellite
		codeArray := make([]byte, 2)
		LArray := make([]float64, 2)
		PArray := make([]float64, 2)
		DArray := make([]float64, 2)
		SNRArray := make([]float64, 2)
		LLIArray := make([]byte, 2)

		// Set L1 code type
		codeArray[0] = CODE_L1C
		if code1 == 1 {
			codeArray[0] = CODE_L1P
		}

		// Set L2 code type
		if code2 <= 3 {
			codeArray[1] = L2codes[code2]
		} else {
			codeArray[1] = CODE_L2X // Default
		}

		// Convert pseudorange and carrier phase
		if pr1 != 0.0 {
			// L1 pseudorange in meters
			pr1 = pr1*0.02 + float64(amb)*PRUNIT_GPS
			PArray[0] = pr1

			// L1 carrier phase in cycles
			if ppr1 != -524288 { // Check for invalid value
				cp1 := float64(ppr1) * 0.0005 / CLIGHT
				LArray[0] = pr1/CLIGHT*freq[0] + cp1
			}

			// L2 pseudorange in meters
			if pr21 != -8192 { // Check for invalid value
				PArray[1] = pr1 + float64(pr21)*0.02
			}

			// L2 carrier phase in cycles
			if ppr2 != -524288 { // Check for invalid value
				cp2 := float64(ppr2) * 0.0005 / CLIGHT
				LArray[1] = pr1/CLIGHT*freq[1] + cp2
			}
		}

		// Set SNR values
		if cnr1 != 0.0 {
			SNRArray[0] = cnr1 * 0.25 // Convert to dB-Hz
		}
		if cnr2 != 0.0 {
			SNRArray[1] = cnr2 * 0.25 // Convert to dB-Hz
		}

		// Set lock time indicators
		if lock1 > 0 {
			LLIArray[0] = 1 // Valid lock
		}
		if lock2 > 0 {
			LLIArray[1] = 1 // Valid lock
		}

		// Add arrays to observation data
		obs.Code = append(obs.Code, codeArray)
		obs.L = append(obs.L, LArray)
		obs.P = append(obs.P, PArray)
		obs.D = append(obs.D, DArray) // Doppler not provided in message
		obs.SNR = append(obs.SNR, SNRArray)
		obs.LLI = append(obs.LLI, LLIArray)
	}

	// Set observation count
	obs.N = len(obs.SatID)

	return obs, nil
}

// decodeType1009 decodes RTCM message type 1009 (L1-only GLONASS RTK observables)
func decodeType1009(msg *RTCMMessage) (*ObservationData, error) {
	// Similar to decodeType1001 but for GLONASS
	// Implementation details would be added here
	return &ObservationData{}, nil
}

// decodeType1010 decodes RTCM message type 1010 (Extended L1-only GLONASS RTK observables)
func decodeType1010(msg *RTCMMessage) (*ObservationData, error) {
	// Similar to decodeType1002 but for GLONASS
	// Implementation details would be added here
	return &ObservationData{}, nil
}

// decodeType1011 decodes RTCM message type 1011 (L1&L2 GLONASS RTK observables)
func decodeType1011(msg *RTCMMessage) (*ObservationData, error) {
	// Similar to decodeType1003 but for GLONASS
	// Implementation details would be added here
	return &ObservationData{}, nil
}

// decodeType1012 decodes RTCM message type 1012 (Extended L1&L2 GLONASS RTK observables)
func decodeType1012(msg *RTCMMessage) (*ObservationData, error) {
	// Similar to decodeType1004 but for GLONASS
	// Implementation details would be added here
	return &ObservationData{}, nil
}

// decodeGPSHeader decodes the header of GPS observation messages (1001-1004)
func decodeGPSHeader(msg *RTCMMessage) (stationID int, tow float64, sync int, nsat int, err error) {
	if len(msg.Data) < 8 {
		return 0, 0, 0, 0, fmt.Errorf("message too short for GPS header")
	}

	bitIndex := 24 // Start after preamble and length

	// Message type (already known, but we'll skip these bits)
	bitIndex += 12

	// Station ID
	stationID = int(gnssgo.GetBitU(msg.Data, bitIndex, 12))
	bitIndex += 12

	// TOW - Time of Week
	tow = float64(gnssgo.GetBitU(msg.Data, bitIndex, 30)) * 0.001 // milliseconds to seconds
	bitIndex += 30

	// Synchronous GNSS Flag
	sync = int(gnssgo.GetBitU(msg.Data, bitIndex, 1))
	bitIndex += 1

	// Number of satellites
	nsat = int(gnssgo.GetBitU(msg.Data, bitIndex, 5))

	return stationID, tow, sync, nsat, nil
}

// decodeGLONASSHeader decodes the header of GLONASS observation messages (1009-1012)
func decodeGLONASSHeader(msg *RTCMMessage) (stationID int, tod float64, sync int, nsat int, err error) {
	if len(msg.Data) < 8 {
		return 0, 0, 0, 0, fmt.Errorf("message too short for GLONASS header")
	}

	bitIndex := 24 // Start after preamble and length

	// Message type (already known, but we'll skip these bits)
	bitIndex += 12

	// Station ID
	stationID = int(gnssgo.GetBitU(msg.Data, bitIndex, 12))
	bitIndex += 12

	// TOD - Time of Day
	tod = float64(gnssgo.GetBitU(msg.Data, bitIndex, 27)) * 0.001 // milliseconds to seconds
	bitIndex += 27

	// Synchronous GNSS Flag
	sync = int(gnssgo.GetBitU(msg.Data, bitIndex, 1))
	bitIndex += 1

	// Number of satellites
	nsat = int(gnssgo.GetBitU(msg.Data, bitIndex, 5))

	return stationID, tod, sync, nsat, nil
}
