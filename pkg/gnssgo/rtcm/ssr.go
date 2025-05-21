package rtcm

import (
	"fmt"
	"math"

	"github.com/bramburn/gnssgo/pkg/gnssgo"
)

// SSRHeader represents the common header for SSR messages
type SSRHeader struct {
	MessageType             int    // Message type
	GNSSID                  int    // GNSS ID (0:GPS, 1:GLONASS, 2:Galileo, 3:QZSS, 4:BeiDou, 5:SBAS, 6:IRNSS)
	Epoch                   uint32 // GNSS epoch time
	UpdateInterval          uint8  // SSR update interval
	MultipleMessage         bool   // Multiple message flag
	SatelliteReferenceDatum bool   // Satellite reference datum flag
	IODSSRIndicator         uint8  // IOD SSR indicator
	SSRProviderID           uint16 // SSR provider ID
	SolutionID              uint8  // SSR solution ID
	NumSatellites           int    // Number of satellites
	SatelliteMask           uint64 // Satellite mask
}

// SSROrbitCorrection represents orbit correction data for a satellite
type SSROrbitCorrection struct {
	SatID              uint8   // Satellite ID
	IODE               uint8   // Issue of data, ephemeris
	DeltaRadial        float64 // Radial orbit correction (m)
	DeltaAlongTrack    float64 // Along-track orbit correction (m)
	DeltaCrossTrack    float64 // Cross-track orbit correction (m)
	DotDeltaRadial     float64 // Rate of radial orbit correction (m/s)
	DotDeltaAlongTrack float64 // Rate of along-track orbit correction (m/s)
	DotDeltaCrossTrack float64 // Rate of cross-track orbit correction (m/s)
}

// SSRClockCorrection represents clock correction data for a satellite
type SSRClockCorrection struct {
	SatID        uint8   // Satellite ID
	DeltaClockC0 float64 // Clock offset (m)
	DeltaClockC1 float64 // Clock drift (m/s)
	DeltaClockC2 float64 // Clock drift rate (m/s²)
}

// SSROrbitClockCorrection represents combined orbit and clock correction data
type SSROrbitClockCorrection struct {
	Header           SSRHeader            // SSR header
	OrbitCorrections []SSROrbitCorrection // Orbit corrections
	ClockCorrections []SSRClockCorrection // Clock corrections
}

// SSRCodeBias represents code bias data for a satellite
type SSRCodeBias struct {
	SatID      uint8     // Satellite ID
	NumBiases  int       // Number of biases
	SignalIDs  []uint8   // Signal IDs
	CodeBiases []float64 // Code biases (m)
}

// SSRCodeBiasCorrection represents code bias correction data
type SSRCodeBiasCorrection struct {
	Header     SSRHeader     // SSR header
	CodeBiases []SSRCodeBias // Code biases
}

// SSRPhaseBias represents phase bias data for a satellite
type SSRPhaseBias struct {
	SatID                     uint8     // Satellite ID
	NumBiases                 int       // Number of biases
	YawAngle                  float64   // Yaw angle (rad)
	YawRate                   float64   // Yaw rate (rad/s)
	SignalIDs                 []uint8   // Signal IDs
	IntegerIndicators         []bool    // Integer indicators
	WideLaneIntegerIndicators []bool    // Wide-lane integer indicators
	DiscontinuityCounters     []uint8   // Discontinuity counters
	PhaseBiases               []float64 // Phase biases (m)
}

// SSRPhaseBiasCorrection represents phase bias correction data
type SSRPhaseBiasCorrection struct {
	Header      SSRHeader      // SSR header
	PhaseBiases []SSRPhaseBias // Phase biases
}

// decodeSSRHeader decodes the common header for SSR messages
func decodeSSRHeader(msg *RTCMMessage) (*SSRHeader, int, error) {
	if msg == nil {
		return nil, 0, fmt.Errorf("nil message")
	}

	// Start position after message type and station ID (24 + 12 = 36 bits)
	pos := 36

	// Create SSR header
	header := &SSRHeader{
		MessageType: msg.Type,
	}

	// Determine GNSS ID from message type
	switch {
	case msg.Type >= 1057 && msg.Type <= 1068:
		header.GNSSID = 0 // GPS
	case msg.Type >= 1063 && msg.Type <= 1068:
		header.GNSSID = 1 // GLONASS
	case msg.Type >= 1240 && msg.Type <= 1245:
		header.GNSSID = 2 // Galileo
	case msg.Type >= 1246 && msg.Type <= 1251:
		header.GNSSID = 3 // QZSS
	case msg.Type >= 1252 && msg.Type <= 1257:
		header.GNSSID = 4 // BeiDou
	case msg.Type >= 1258 && msg.Type <= 1263:
		header.GNSSID = 5 // SBAS
	case msg.Type >= 1264 && msg.Type <= 1269:
		header.GNSSID = 6 // IRNSS
	default:
		return nil, 0, fmt.Errorf("unknown SSR message type: %d", msg.Type)
	}

	// Decode epoch time
	header.Epoch = uint32(gnssgo.GetBitU(msg.Data, pos, 20))
	pos += 20

	// Decode update interval
	header.UpdateInterval = uint8(gnssgo.GetBitU(msg.Data, pos, 4))
	pos += 4

	// Decode multiple message flag
	header.MultipleMessage = gnssgo.GetBitU(msg.Data, pos, 1) != 0
	pos += 1

	// Decode satellite reference datum flag
	header.SatelliteReferenceDatum = gnssgo.GetBitU(msg.Data, pos, 1) != 0
	pos += 1

	// Decode IOD SSR indicator
	header.IODSSRIndicator = uint8(gnssgo.GetBitU(msg.Data, pos, 4))
	pos += 4

	// Decode SSR provider ID
	header.SSRProviderID = uint16(gnssgo.GetBitU(msg.Data, pos, 16))
	pos += 16

	// Decode SSR solution ID
	header.SolutionID = uint8(gnssgo.GetBitU(msg.Data, pos, 4))
	pos += 4

	// Decode number of satellites
	numSats := int(gnssgo.GetBitU(msg.Data, pos, 6))
	header.NumSatellites = numSats
	pos += 6

	// Decode satellite mask
	header.SatelliteMask = 0
	for i := 0; i < numSats; i++ {
		satID := int(gnssgo.GetBitU(msg.Data, pos, 6))
		pos += 6
		header.SatelliteMask |= 1 << (satID - 1)
	}

	return header, pos, nil
}

// decodeSSROrbitCorrection decodes orbit correction data for a satellite
func decodeSSROrbitCorrection(msg *RTCMMessage, pos int) (*SSROrbitCorrection, int, error) {
	if msg == nil {
		return nil, 0, fmt.Errorf("nil message")
	}

	// Create orbit correction
	orb := &SSROrbitCorrection{}

	// Decode satellite ID
	orb.SatID = uint8(gnssgo.GetBitU(msg.Data, pos, 6))
	pos += 6

	// Decode IODE
	orb.IODE = uint8(gnssgo.GetBitU(msg.Data, pos, 8))
	pos += 8

	// Decode delta radial
	orb.DeltaRadial = float64(gnssgo.GetBits(msg.Data, pos, 22)) * 0.1 * 0.001 // 0.1 mm
	pos += 22

	// Decode delta along-track
	orb.DeltaAlongTrack = float64(gnssgo.GetBits(msg.Data, pos, 20)) * 0.4 * 0.001 // 0.4 mm
	pos += 20

	// Decode delta cross-track
	orb.DeltaCrossTrack = float64(gnssgo.GetBits(msg.Data, pos, 20)) * 0.4 * 0.001 // 0.4 mm
	pos += 20

	// Decode dot delta radial
	orb.DotDeltaRadial = float64(gnssgo.GetBits(msg.Data, pos, 21)) * 0.001 * 0.001 // 0.001 mm/s
	pos += 21

	// Decode dot delta along-track
	orb.DotDeltaAlongTrack = float64(gnssgo.GetBits(msg.Data, pos, 19)) * 0.004 * 0.001 // 0.004 mm/s
	pos += 19

	// Decode dot delta cross-track
	orb.DotDeltaCrossTrack = float64(gnssgo.GetBits(msg.Data, pos, 19)) * 0.004 * 0.001 // 0.004 mm/s
	pos += 19

	return orb, pos, nil
}

// decodeSSRClockCorrection decodes clock correction data for a satellite
func decodeSSRClockCorrection(msg *RTCMMessage, pos int) (*SSRClockCorrection, int, error) {
	if msg == nil {
		return nil, 0, fmt.Errorf("nil message")
	}

	// Create clock correction
	clk := &SSRClockCorrection{}

	// Decode satellite ID
	clk.SatID = uint8(gnssgo.GetBitU(msg.Data, pos, 6))
	pos += 6

	// Decode delta clock C0
	clk.DeltaClockC0 = float64(gnssgo.GetBits(msg.Data, pos, 22)) * 0.1 * 0.001 // 0.1 mm
	pos += 22

	// Decode delta clock C1
	clk.DeltaClockC1 = float64(gnssgo.GetBits(msg.Data, pos, 21)) * 0.001 * 0.001 // 0.001 mm/s
	pos += 21

	// Decode delta clock C2
	clk.DeltaClockC2 = float64(gnssgo.GetBits(msg.Data, pos, 27)) * 0.00002 * 0.001 // 0.00002 mm/s²
	pos += 27

	return clk, pos, nil
}

// decodeSSROrbitClockCorrection decodes combined orbit and clock correction data
func decodeSSROrbitClockCorrection(msg *RTCMMessage) (*SSROrbitClockCorrection, error) {
	if msg == nil {
		return nil, fmt.Errorf("nil message")
	}

	// Decode SSR header
	header, pos, err := decodeSSRHeader(msg)
	if err != nil {
		return nil, err
	}

	// Create orbit and clock correction
	correction := &SSROrbitClockCorrection{
		Header:           *header,
		OrbitCorrections: make([]SSROrbitCorrection, header.NumSatellites),
		ClockCorrections: make([]SSRClockCorrection, header.NumSatellites),
	}

	// Decode orbit corrections
	for i := 0; i < header.NumSatellites; i++ {
		orb, newPos, err := decodeSSROrbitCorrection(msg, pos)
		if err != nil {
			return nil, err
		}
		correction.OrbitCorrections[i] = *orb
		pos = newPos
	}

	// Decode clock corrections
	for i := 0; i < header.NumSatellites; i++ {
		clk, newPos, err := decodeSSRClockCorrection(msg, pos)
		if err != nil {
			return nil, err
		}
		correction.ClockCorrections[i] = *clk
		pos = newPos
	}

	return correction, nil
}

// decodeSSRCodeBias decodes code bias data for a satellite
func decodeSSRCodeBias(msg *RTCMMessage) (*SSRCodeBiasCorrection, error) {
	if msg == nil {
		return nil, fmt.Errorf("nil message")
	}

	// Decode SSR header
	header, pos, err := decodeSSRHeader(msg)
	if err != nil {
		return nil, err
	}

	// Create code bias correction
	correction := &SSRCodeBiasCorrection{
		Header:     *header,
		CodeBiases: make([]SSRCodeBias, header.NumSatellites),
	}

	// Decode code biases for each satellite
	for i := 0; i < header.NumSatellites; i++ {
		// Decode satellite ID
		satID := uint8(gnssgo.GetBitU(msg.Data, pos, 6))
		pos += 6

		// Decode number of biases
		numBiases := int(gnssgo.GetBitU(msg.Data, pos, 5))
		pos += 5

		// Create code bias
		bias := &SSRCodeBias{
			SatID:      satID,
			NumBiases:  numBiases,
			SignalIDs:  make([]uint8, numBiases),
			CodeBiases: make([]float64, numBiases),
		}

		// Decode biases
		for j := 0; j < numBiases; j++ {
			// Decode signal ID
			bias.SignalIDs[j] = uint8(gnssgo.GetBitU(msg.Data, pos, 5))
			pos += 5

			// Decode code bias
			bias.CodeBiases[j] = float64(gnssgo.GetBits(msg.Data, pos, 14)) * 0.01 // 0.01 m
			pos += 14
		}

		correction.CodeBiases[i] = *bias
	}

	return correction, nil
}

// decodeSSRPhaseBias decodes phase bias data for a satellite
func decodeSSRPhaseBias(msg *RTCMMessage) (*SSRPhaseBiasCorrection, error) {
	if msg == nil {
		return nil, fmt.Errorf("nil message")
	}

	// Decode SSR header
	header, pos, err := decodeSSRHeader(msg)
	if err != nil {
		return nil, err
	}

	// Create phase bias correction
	correction := &SSRPhaseBiasCorrection{
		Header:      *header,
		PhaseBiases: make([]SSRPhaseBias, header.NumSatellites),
	}

	// Decode phase biases for each satellite
	for i := 0; i < header.NumSatellites; i++ {
		// Decode satellite ID
		satID := uint8(gnssgo.GetBitU(msg.Data, pos, 6))
		pos += 6

		// Decode number of biases
		numBiases := int(gnssgo.GetBitU(msg.Data, pos, 5))
		pos += 5

		// Decode yaw angle
		yawAngle := float64(gnssgo.GetBitU(msg.Data, pos, 9)) * 1.0 * math.Pi / 180.0 // 1 degree to rad
		pos += 9

		// Decode yaw rate
		yawRate := float64(gnssgo.GetBits(msg.Data, pos, 8)) * 0.1 * math.Pi / 180.0 // 0.1 degree/s to rad/s
		pos += 8

		// Create phase bias
		bias := &SSRPhaseBias{
			SatID:                     satID,
			NumBiases:                 numBiases,
			YawAngle:                  yawAngle,
			YawRate:                   yawRate,
			SignalIDs:                 make([]uint8, numBiases),
			IntegerIndicators:         make([]bool, numBiases),
			WideLaneIntegerIndicators: make([]bool, numBiases),
			DiscontinuityCounters:     make([]uint8, numBiases),
			PhaseBiases:               make([]float64, numBiases),
		}

		// Decode biases
		for j := 0; j < numBiases; j++ {
			// Decode signal ID
			bias.SignalIDs[j] = uint8(gnssgo.GetBitU(msg.Data, pos, 5))
			pos += 5

			// Decode integer indicator
			bias.IntegerIndicators[j] = gnssgo.GetBitU(msg.Data, pos, 1) != 0
			pos += 1

			// Decode wide-lane integer indicator
			wlIntInd := gnssgo.GetBitU(msg.Data, pos, 2)
			bias.WideLaneIntegerIndicators[j] = wlIntInd != 0
			pos += 2

			// Decode discontinuity counter
			bias.DiscontinuityCounters[j] = uint8(gnssgo.GetBitU(msg.Data, pos, 4))
			pos += 4

			// Decode phase bias
			bias.PhaseBiases[j] = float64(gnssgo.GetBits(msg.Data, pos, 20)) * 0.0001 // 0.0001 m
			pos += 20
		}

		correction.PhaseBiases[i] = *bias
	}

	return correction, nil
}
