package rtcm

import (
	"fmt"

	"github.com/bramburn/gnssgo/pkg/gnssgo"
)

// MSM message types
const (
	MSM1 = 1 // Compact pseudoranges only
	MSM2 = 2 // Compact phase-ranges only
	MSM3 = 3 // Compact pseudoranges and phase-ranges
	MSM4 = 4 // Full pseudoranges and phase-ranges plus CNR
	MSM5 = 5 // Full pseudoranges, phase-ranges, phase-range-rates and CNR
	MSM6 = 6 // Full pseudoranges and phase-ranges plus CNR (high resolution)
	MSM7 = 7 // Full pseudoranges, phase-ranges, phase-range-rates and CNR (high resolution)
)

// MSMHeader represents the header of an MSM message
type MSMHeader struct {
	MessageType            int     // Message type
	StationID              uint16  // Reference station ID
	GNSSID                 int     // GNSS ID (0:GPS, 1:GLONASS, 2:Galileo, 3:SBAS, 4:QZSS, 5:BeiDou, 6:IRNSS)
	Epoch                  uint32  // GNSS epoch time
	MultipleMessage        bool    // Multiple message bit
	IssueOfDataStation     uint8   // IODS
	ClockSteeringIndicator uint8   // Clock steering indicator
	ExternalClockIndicator uint8   // External clock indicator
	SmoothingIndicator     bool    // Divergence-free smoothing indicator
	SmoothingInterval      uint8   // Smoothing interval
	SatelliteMask          uint64  // Satellite mask
	SignalMask             uint32  // Signal mask
	CellMask               []uint8 // Cell mask
	NumSatellites          int     // Number of satellites
	NumSignals             int     // Number of signals
	NumCells               int     // Number of cells (satellite-signal combinations)
}

// MSMSatellite represents satellite data in an MSM message
type MSMSatellite struct {
	ID             int     // Satellite ID
	RangeInteger   uint8   // Integer milliseconds of ranges
	ExtendedInfo   uint8   // Extended satellite info
	RangeModulo    float64 // Range modulo 1 millisecond (m)
	PhaseRangeRate float64 // Phase range rate (m/s)
}

// MSMSignal represents signal data in an MSM message
type MSMSignal struct {
	Type               int     // Signal type
	Code               int     // Signal code
	Pseudorange        float64 // Pseudorange (m)
	PhaseRange         float64 // Phase range (cycles)
	PhaseRangeLockTime uint16  // Lock time indicator
	HalfCycleAmbiguity bool    // Half-cycle ambiguity indicator
	CNR                float64 // Carrier-to-noise ratio (dB-Hz)
	PhaseRangeRate     float64 // Phase range rate (m/s)
}

// MSMData represents the decoded data from an MSM message
type MSMData struct {
	Header     MSMHeader      // MSM header
	Satellites []MSMSatellite // Satellite data
	Signals    []MSMSignal    // Signal data
	Cells      []int          // Cell indices (satellite-signal combinations)
}

// decodeMSMMessage decodes an MSM message
func decodeMSMMessage(msg *RTCMMessage, sys int) (*MSMData, error) {
	if msg == nil {
		return nil, fmt.Errorf("nil message")
	}

	// Determine MSM type (1-7)
	msmType := 0
	switch {
	case msg.Type >= MSM_GPS_RANGE_START && msg.Type <= MSM_GPS_RANGE_END:
		msmType = msg.Type - MSM_GPS_RANGE_START + 1
	case msg.Type >= MSM_GLONASS_RANGE_START && msg.Type <= MSM_GLONASS_RANGE_END:
		msmType = msg.Type - MSM_GLONASS_RANGE_START + 1
	case msg.Type >= MSM_GALILEO_RANGE_START && msg.Type <= MSM_GALILEO_RANGE_END:
		msmType = msg.Type - MSM_GALILEO_RANGE_START + 1
	case msg.Type >= MSM_SBAS_RANGE_START && msg.Type <= MSM_SBAS_RANGE_END:
		msmType = msg.Type - MSM_SBAS_RANGE_START + 1
	case msg.Type >= MSM_QZSS_RANGE_START && msg.Type <= MSM_QZSS_RANGE_END:
		msmType = msg.Type - MSM_QZSS_RANGE_START + 1
	case msg.Type >= MSM_BEIDOU_RANGE_START && msg.Type <= MSM_BEIDOU_RANGE_END:
		msmType = msg.Type - MSM_BEIDOU_RANGE_START + 1
	case msg.Type >= MSM_IRNSS_RANGE_START && msg.Type <= MSM_IRNSS_RANGE_END:
		msmType = msg.Type - MSM_IRNSS_RANGE_START + 1
	default:
		return nil, fmt.Errorf("not an MSM message: type %d", msg.Type)
	}

	// Decode MSM header
	header, pos, err := decodeMSMHeader(msg, sys)
	if err != nil {
		return nil, err
	}

	// Create MSM data structure
	data := &MSMData{
		Header:     *header,
		Satellites: make([]MSMSatellite, header.NumSatellites),
		Signals:    make([]MSMSignal, header.NumCells),
		Cells:      make([]int, header.NumCells),
	}

	// Decode satellite data
	pos, err = decodeMSMSatellites(msg, data, pos, msmType)
	if err != nil {
		return nil, err
	}

	// Decode signal data
	_, err = decodeMSMSignals(msg, data, pos, msmType)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// decodeMSMHeader decodes the header of an MSM message
func decodeMSMHeader(msg *RTCMMessage, sys int) (*MSMHeader, int, error) {
	if msg == nil || len(msg.Data) < 10 {
		return nil, 0, fmt.Errorf("message too short for MSM header")
	}

	header := &MSMHeader{
		MessageType: msg.Type,
		StationID:   msg.StationID,
		GNSSID:      getGNSSIDFromSystem(sys),
	}

	// Start position after message type and station ID (24 + 12 = 36 bits)
	pos := 36

	// Decode epoch time
	if sys == gnssgo.SYS_GLO {
		// GLONASS uses 27-bit epoch time
		header.Epoch = uint32(gnssgo.GetBitU(msg.Data, pos, 27))
		pos += 27
	} else {
		// Other systems use 30-bit epoch time
		header.Epoch = uint32(gnssgo.GetBitU(msg.Data, pos, 30))
		pos += 30
	}

	// Decode flags
	header.MultipleMessage = gnssgo.GetBitU(msg.Data, pos, 1) != 0
	pos++
	header.IssueOfDataStation = uint8(gnssgo.GetBitU(msg.Data, pos, 3))
	pos += 3
	header.ClockSteeringIndicator = uint8(gnssgo.GetBitU(msg.Data, pos, 2))
	pos += 2
	header.ExternalClockIndicator = uint8(gnssgo.GetBitU(msg.Data, pos, 2))
	pos += 2
	header.SmoothingIndicator = gnssgo.GetBitU(msg.Data, pos, 1) != 0
	pos++
	header.SmoothingInterval = uint8(gnssgo.GetBitU(msg.Data, pos, 3))
	pos += 3

	// Decode satellite mask (up to 64 satellites)
	// We need to read this in two 32-bit chunks since GetBitU returns uint32
	header.SatelliteMask = uint64(gnssgo.GetBitU(msg.Data, pos, 32)) |
		(uint64(gnssgo.GetBitU(msg.Data, pos+32, 32)) << 32)
	pos += 64

	// Count number of satellites
	header.NumSatellites = countBits(header.SatelliteMask)

	// Decode signal mask (up to 32 signals)
	header.SignalMask = uint32(gnssgo.GetBitU(msg.Data, pos, 32))
	pos += 32

	// Count number of signals
	header.NumSignals = countBits32(header.SignalMask)

	// Decode cell mask
	cellMaskSize := header.NumSatellites * header.NumSignals
	header.CellMask = make([]uint8, (cellMaskSize+7)/8) // Round up to nearest byte

	for i := 0; i < cellMaskSize; i++ {
		if gnssgo.GetBitU(msg.Data, pos, 1) != 0 {
			header.CellMask[i/8] |= 1 << (i % 8)
			header.NumCells++
		}
		pos++
	}

	return header, pos, nil
}

// decodeMSMSatellites decodes satellite data from an MSM message
func decodeMSMSatellites(msg *RTCMMessage, data *MSMData, pos int, msmType int) (int, error) {
	header := &data.Header

	// For each satellite in the mask
	satIndex := 0
	for i := 0; i < 64; i++ {
		if (header.SatelliteMask & (1 << i)) == 0 {
			continue
		}

		// Create satellite entry
		sat := &data.Satellites[satIndex]
		sat.ID = i + 1 // Satellite IDs are 1-based

		// Decode satellite data based on MSM type
		switch {
		case msmType >= MSM4: // MSM4-7
			// Decode range integer (8 bits)
			sat.RangeInteger = uint8(gnssgo.GetBitU(msg.Data, pos, 8))
			pos += 8

			// For MSM5 and MSM7, decode extended info
			if msmType == MSM5 || msmType == MSM7 {
				sat.ExtendedInfo = uint8(gnssgo.GetBitU(msg.Data, pos, 4))
				pos += 4
			}
		}

		satIndex++
	}

	// Decode satellite data fields
	satIndex = 0
	for i := 0; i < 64; i++ {
		if (header.SatelliteMask & (1 << i)) == 0 {
			continue
		}

		sat := &data.Satellites[satIndex]

		// Decode range modulo based on MSM type
		switch msmType {
		case MSM1, MSM2, MSM3:
			// 10-bit range modulo (1 ms resolution)
			sat.RangeModulo = float64(gnssgo.GetBitU(msg.Data, pos, 10)) * 1.0
			pos += 10
		case MSM4, MSM5:
			// 15-bit range modulo (1/1024 ms resolution)
			sat.RangeModulo = float64(gnssgo.GetBitU(msg.Data, pos, 15)) * (1.0 / 1024.0)
			pos += 15
		case MSM6, MSM7:
			// 20-bit range modulo (1/16384 ms resolution)
			sat.RangeModulo = float64(gnssgo.GetBitU(msg.Data, pos, 20)) * (1.0 / 16384.0)
			pos += 20
		}

		// For MSM5 and MSM7, decode phase range rate
		if msmType == MSM5 || msmType == MSM7 {
			// Phase range rate
			rate := int32(gnssgo.GetBits(msg.Data, pos, 15))
			if msmType == MSM5 {
				// MSM5: 15-bit phase range rate (0.1 m/s resolution)
				sat.PhaseRangeRate = float64(rate) * 0.1
				pos += 15
			} else {
				// MSM7: 20-bit phase range rate (0.0001 m/s resolution)
				sat.PhaseRangeRate = float64(rate) * 0.0001
				pos += 20
			}
		}

		satIndex++
	}

	return pos, nil
}

// decodeMSMSignals decodes signal data from an MSM message
func decodeMSMSignals(msg *RTCMMessage, data *MSMData, pos int, msmType int) (int, error) {
	header := &data.Header

	// For each cell in the mask
	cellIndex := 0
	for i := 0; i < 64; i++ {
		if (header.SatelliteMask & (1 << i)) == 0 {
			continue
		}

		for j := 0; j < 32; j++ {
			if (header.SignalMask & (1 << j)) == 0 {
				continue
			}

			// Check if this cell is present
			cellBit := i*header.NumSignals + j
			if (header.CellMask[cellBit/8] & (1 << (cellBit % 8))) == 0 {
				continue
			}

			// Store cell index
			data.Cells[cellIndex] = cellBit

			// Create signal entry
			signal := &data.Signals[cellIndex]
			signal.Type = j + 1 // Signal types are 1-based
			signal.Code = getSignalCode(header.GNSSID, j)

			cellIndex++
		}
	}

	// Decode pseudoranges
	if msmType == MSM1 || msmType == MSM3 || msmType == MSM4 || msmType == MSM5 || msmType == MSM6 || msmType == MSM7 {
		for i := 0; i < header.NumCells; i++ {
			cellBit := data.Cells[i]
			satIdx := cellBit / header.NumSignals
			satID := 0

			// Find satellite index
			for j := 0; j < 64; j++ {
				if (header.SatelliteMask & (1 << j)) != 0 {
					if satIdx == 0 {
						satID = j
						break
					}
					satIdx--
				}
			}

			// Get satellite data
			var sat *MSMSatellite
			for j := 0; j < header.NumSatellites; j++ {
				if data.Satellites[j].ID == satID+1 {
					sat = &data.Satellites[j]
					break
				}
			}

			if sat == nil {
				continue
			}

			// Decode pseudorange
			signal := &data.Signals[i]

			switch msmType {
			case MSM1, MSM3:
				// 15-bit pseudorange (1 dm resolution)
				pr := int32(gnssgo.GetBits(msg.Data, pos, 15))
				if pr != -16384 { // Not invalid
					signal.Pseudorange = float64(sat.RangeInteger)*299792.458 +
						sat.RangeModulo*299792.458 +
						float64(pr)*0.1
				}
				pos += 15
			case MSM4, MSM5:
				// 20-bit pseudorange (1 cm resolution)
				pr := int32(gnssgo.GetBits(msg.Data, pos, 20))
				if pr != -524288 { // Not invalid
					signal.Pseudorange = float64(sat.RangeInteger)*299792.458 +
						sat.RangeModulo*299792.458 +
						float64(pr)*0.01
				}
				pos += 20
			case MSM6, MSM7:
				// 24-bit pseudorange (0.1 mm resolution)
				pr := int32(gnssgo.GetBits(msg.Data, pos, 24))
				if pr != -8388608 { // Not invalid
					signal.Pseudorange = float64(sat.RangeInteger)*299792.458 +
						sat.RangeModulo*299792.458 +
						float64(pr)*0.0001
				}
				pos += 24
			}
		}
	}

	// Decode phase ranges
	if msmType == MSM2 || msmType == MSM3 || msmType == MSM4 || msmType == MSM5 || msmType == MSM6 || msmType == MSM7 {
		for i := 0; i < header.NumCells; i++ {
			cellBit := data.Cells[i]
			satIdx := cellBit / header.NumSignals
			satID := 0

			// Find satellite index
			for j := 0; j < 64; j++ {
				if (header.SatelliteMask & (1 << j)) != 0 {
					if satIdx == 0 {
						satID = j
						break
					}
					satIdx--
				}
			}

			// Get satellite data
			var sat *MSMSatellite
			for j := 0; j < header.NumSatellites; j++ {
				if data.Satellites[j].ID == satID+1 {
					sat = &data.Satellites[j]
					break
				}
			}

			if sat == nil {
				continue
			}

			// Decode phase range
			signal := &data.Signals[i]

			switch msmType {
			case MSM2, MSM3:
				// 22-bit phase range (0.0001 cycles resolution)
				phr := int32(gnssgo.GetBits(msg.Data, pos, 22))
				if phr != -2097152 { // Not invalid
					signal.PhaseRange = float64(sat.RangeInteger)*299792.458/gnssgo.CLIGHT +
						sat.RangeModulo*299792.458/gnssgo.CLIGHT +
						float64(phr)*0.0001
				}
				pos += 22
			case MSM4, MSM5:
				// 24-bit phase range (0.0001 cycles resolution)
				phr := int32(gnssgo.GetBits(msg.Data, pos, 24))
				if phr != -8388608 { // Not invalid
					signal.PhaseRange = float64(sat.RangeInteger)*299792.458/gnssgo.CLIGHT +
						sat.RangeModulo*299792.458/gnssgo.CLIGHT +
						float64(phr)*0.0001
				}
				pos += 24
			case MSM6, MSM7:
				// 29-bit phase range (0.0000001 cycles resolution)
				phr := int32(gnssgo.GetBits(msg.Data, pos, 29))
				if phr != -268435456 { // Not invalid
					signal.PhaseRange = float64(sat.RangeInteger)*299792.458/gnssgo.CLIGHT +
						sat.RangeModulo*299792.458/gnssgo.CLIGHT +
						float64(phr)*0.0000001
				}
				pos += 29
			}
		}
	}

	// Decode lock time indicators
	if msmType == MSM2 || msmType == MSM3 || msmType == MSM4 || msmType == MSM5 || msmType == MSM6 || msmType == MSM7 {
		for i := 0; i < header.NumCells; i++ {
			signal := &data.Signals[i]

			switch msmType {
			case MSM2, MSM3, MSM4, MSM5:
				// 4-bit lock time indicator
				lock := uint16(gnssgo.GetBitU(msg.Data, pos, 4))
				signal.PhaseRangeLockTime = lock
				pos += 4
			case MSM6, MSM7:
				// 10-bit lock time indicator
				lock := uint16(gnssgo.GetBitU(msg.Data, pos, 10))
				signal.PhaseRangeLockTime = lock
				pos += 10
			}
		}
	}

	// Decode half-cycle ambiguity indicators
	if msmType == MSM2 || msmType == MSM3 || msmType == MSM4 || msmType == MSM5 || msmType == MSM6 || msmType == MSM7 {
		for i := 0; i < header.NumCells; i++ {
			signal := &data.Signals[i]

			// 1-bit half-cycle ambiguity indicator
			half := gnssgo.GetBitU(msg.Data, pos, 1) != 0
			signal.HalfCycleAmbiguity = half
			pos += 1
		}
	}

	// Decode CNR
	if msmType == MSM4 || msmType == MSM5 || msmType == MSM6 || msmType == MSM7 {
		for i := 0; i < header.NumCells; i++ {
			signal := &data.Signals[i]

			switch msmType {
			case MSM4, MSM5:
				// 6-bit CNR (1 dB-Hz resolution)
				cnr := uint8(gnssgo.GetBitU(msg.Data, pos, 6))
				signal.CNR = float64(cnr)
				pos += 6
			case MSM6, MSM7:
				// 10-bit CNR (0.0625 dB-Hz resolution)
				cnr := uint16(gnssgo.GetBitU(msg.Data, pos, 10))
				signal.CNR = float64(cnr) * 0.0625
				pos += 10
			}
		}
	}

	// Decode phase range rates
	if msmType == MSM5 || msmType == MSM7 {
		for i := 0; i < header.NumCells; i++ {
			signal := &data.Signals[i]

			switch msmType {
			case MSM5:
				// 8-bit phase range rate (0.1 m/s resolution)
				rate := int8(gnssgo.GetBits(msg.Data, pos, 8))
				signal.PhaseRangeRate = float64(rate) * 0.1
				pos += 8
			case MSM7:
				// 14-bit phase range rate (0.0001 m/s resolution)
				rate := int16(gnssgo.GetBits(msg.Data, pos, 14))
				signal.PhaseRangeRate = float64(rate) * 0.0001
				pos += 14
			}
		}
	}

	return pos, nil
}

// Helper functions

// countBits counts the number of bits set in a 64-bit value
func countBits(value uint64) int {
	count := 0
	for i := 0; i < 64; i++ {
		if (value & (1 << i)) != 0 {
			count++
		}
	}
	return count
}

// countBits32 counts the number of bits set in a 32-bit value
func countBits32(value uint32) int {
	count := 0
	for i := 0; i < 32; i++ {
		if (value & (1 << i)) != 0 {
			count++
		}
	}
	return count
}

// getGNSSIDFromSystem converts a GNSS system ID to an MSM GNSS ID
func getGNSSIDFromSystem(sys int) int {
	switch sys {
	case gnssgo.SYS_GPS:
		return 0
	case gnssgo.SYS_GLO:
		return 1
	case gnssgo.SYS_GAL:
		return 2
	case gnssgo.SYS_SBS:
		return 3
	case gnssgo.SYS_QZS:
		return 4
	case gnssgo.SYS_CMP:
		return 5
	case gnssgo.SYS_IRN:
		return 6
	default:
		return 0
	}
}

// getSignalCode returns the signal code for a given GNSS ID and signal ID
// Based on RTCM 3.3 signal definitions
func getSignalCode(gnssID, signalID int) int {
	// Signal codes based on RTCM 3.3 Table 3.5-91
	switch gnssID {
	case 0: // GPS
		switch signalID {
		case 0: // L1 C/A
			return 1
		case 1: // L1 P
			return 2
		case 2: // L1 Z-tracking
			return 3
		case 3: // L1C (1) - pilot component
			return 4
		case 4: // L1C (2) - data component
			return 5
		case 5: // L2 CM
			return 6
		case 6: // L2 CL
			return 7
		case 7: // L2 P
			return 8
		case 8: // L2 Z-tracking
			return 9
		case 9: // L5 I
			return 10
		case 10: // L5 Q
			return 11
		case 15: // L1 L-SAIF
			return 16
		default:
			return 0
		}
	case 1: // GLONASS
		switch signalID {
		case 0: // G1 C/A
			return 1
		case 1: // G1 P
			return 2
		case 2: // G2 C/A
			return 3
		case 3: // G2 P
			return 4
		case 4: // G3 I
			return 5
		case 5: // G3 Q
			return 6
		default:
			return 0
		}
	case 2: // Galileo
		switch signalID {
		case 0: // E1 C
			return 1
		case 1: // E1 A
			return 2
		case 2: // E1 B
			return 3
		case 3: // E1 B+C
			return 4
		case 4: // E1 A+B+C
			return 5
		case 5: // E5a I
			return 6
		case 6: // E5a Q
			return 7
		case 7: // E5a I+Q
			return 8
		case 8: // E5b I
			return 9
		case 9: // E5b Q
			return 10
		case 10: // E5b I+Q
			return 11
		case 11: // E5 I
			return 12
		case 12: // E5 Q
			return 13
		case 13: // E5 I+Q
			return 14
		case 14: // E6 C
			return 15
		case 15: // E6 A
			return 16
		case 16: // E6 B
			return 17
		case 17: // E6 B+C
			return 18
		case 18: // E6 A+B+C
			return 19
		default:
			return 0
		}
	case 3: // SBAS
		switch signalID {
		case 0: // L1 C/A
			return 1
		case 1: // L5 I
			return 2
		case 2: // L5 Q
			return 3
		case 3: // L5 I+Q
			return 4
		default:
			return 0
		}
	case 4: // QZSS
		switch signalID {
		case 0: // L1 C/A
			return 1
		case 1: // L1 SAIF
			return 2
		case 2: // L1C (1) - pilot component
			return 3
		case 3: // L1C (2) - data component
			return 4
		case 4: // L2 CM
			return 5
		case 5: // L2 CL
			return 6
		case 6: // L5 I
			return 7
		case 7: // L5 Q
			return 8
		case 8: // L5 I+Q
			return 9
		case 9: // L6 D
			return 10
		case 10: // L6 E
			return 11
		case 11: // L6 D+E
			return 12
		default:
			return 0
		}
	case 5: // BeiDou
		switch signalID {
		case 0: // B1I
			return 1
		case 1: // B1Q
			return 2
		case 2: // B1I+Q
			return 3
		case 3: // B3I
			return 4
		case 4: // B3Q
			return 5
		case 5: // B3I+Q
			return 6
		case 6: // B2I
			return 7
		case 7: // B2Q
			return 8
		case 8: // B2I+Q
			return 9
		case 9: // B1C (1) - pilot component
			return 10
		case 10: // B1C (2) - data component
			return 11
		case 11: // B5 I
			return 12
		case 12: // B5 Q
			return 13
		case 13: // B5 I+Q
			return 14
		case 14: // B2a I
			return 15
		case 15: // B2a Q
			return 16
		case 16: // B2a I+Q
			return 17
		case 17: // B2b I
			return 18
		case 18: // B2b Q
			return 19
		case 19: // B2b I+Q
			return 20
		default:
			return 0
		}
	case 6: // IRNSS
		switch signalID {
		case 0: // L5 SPS
			return 1
		case 1: // S SPS
			return 2
		case 2: // L5 RS
			return 3
		case 3: // S RS
			return 4
		case 4: // L5 + S SPS
			return 5
		default:
			return 0
		}
	default:
		return 0
	}
}
