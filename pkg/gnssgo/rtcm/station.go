package rtcm

import (
	"fmt"
	"strings"

	"github.com/bramburn/gnssgo/pkg/gnssgo"
)

// StationCoordinates represents the station coordinates from RTCM message 1005
type StationCoordinates struct {
	StationID      uint16  // Reference station ID
	ITRF           uint8   // ITRF realization year
	GPS            bool    // GPS indicator
	GLONASS        bool    // GLONASS indicator
	Galileo        bool    // Galileo indicator
	ReferencePoint bool    // Reference-station indicator
	SingleReceiver bool    // Single receiver oscillator indicator
	X              float64 // ECEF X coordinate (m)
	Y              float64 // ECEF Y coordinate (m)
	Z              float64 // ECEF Z coordinate (m)
}

// StationCoordinatesAlt represents the station coordinates with height from RTCM message 1006
type StationCoordinatesAlt struct {
	StationCoordinates
	AntennaHeight float64 // Antenna height (m)
}

// AntennaDescriptor represents the antenna descriptor from RTCM message 1007
type AntennaDescriptor struct {
	StationID      uint16 // Reference station ID
	AntennaSetupID uint8  // Antenna setup ID
	AntennaType    string // Antenna descriptor
}

// AntennaDescriptorSerial represents the antenna descriptor and serial number from RTCM message 1008
type AntennaDescriptorSerial struct {
	AntennaDescriptor
	AntennaSerial string // Antenna serial number
}

// ReceiverInfo represents the receiver and antenna descriptor from RTCM message 1033
type ReceiverInfo struct {
	StationID        uint16 // Reference station ID
	ReceiverType     string // Receiver type descriptor
	ReceiverFirmware string // Receiver firmware version
	ReceiverSerial   string // Receiver serial number
	AntennaType      string // Antenna type descriptor
	AntennaSerial    string // Antenna serial number
	AntennaSetupID   uint8  // Antenna setup ID
}

// decodeStationCoordinates decodes RTCM message 1005 (Station Coordinates)
func decodeStationCoordinates(msg *RTCMMessage) (*StationCoordinates, error) {
	if msg == nil || msg.Type != RTCM_STATION_COORDINATES {
		return nil, fmt.Errorf("not a station coordinates message")
	}

	if len(msg.Data) < 10 {
		return nil, fmt.Errorf("message too short for station coordinates")
	}

	// Start position after message type and station ID (24 + 12 = 36 bits)
	pos := 36

	// Create station coordinates
	sc := &StationCoordinates{
		StationID: msg.StationID,
	}

	// Decode flags
	sc.ITRF = uint8(gnssgo.GetBitU(msg.Data, pos, 6))
	pos += 6
	sc.GPS = gnssgo.GetBitU(msg.Data, pos, 1) != 0
	pos++
	sc.GLONASS = gnssgo.GetBitU(msg.Data, pos, 1) != 0
	pos++
	sc.Galileo = gnssgo.GetBitU(msg.Data, pos, 1) != 0
	pos++
	sc.ReferencePoint = gnssgo.GetBitU(msg.Data, pos, 1) != 0
	pos++
	sc.SingleReceiver = gnssgo.GetBitU(msg.Data, pos, 1) != 0
	pos++
	pos += 1 // Reserved bit

	// Decode coordinates
	// X coordinate (38 bits, 0.0001 m resolution, signed)
	x := int64(gnssgo.GetBits(msg.Data, pos, 38))
	sc.X = float64(x) * 0.0001
	pos += 38

	// Y coordinate (38 bits, 0.0001 m resolution, signed)
	y := int64(gnssgo.GetBits(msg.Data, pos, 38))
	sc.Y = float64(y) * 0.0001
	pos += 38

	// Z coordinate (38 bits, 0.0001 m resolution, signed)
	z := int64(gnssgo.GetBits(msg.Data, pos, 38))
	sc.Z = float64(z) * 0.0001

	return sc, nil
}

// decodeStationCoordinatesAlt decodes RTCM message 1006 (Station Coordinates with Height)
func decodeStationCoordinatesAlt(msg *RTCMMessage) (*StationCoordinatesAlt, error) {
	if msg == nil || msg.Type != RTCM_STATION_COORDINATES_ALT {
		return nil, fmt.Errorf("not a station coordinates with height message")
	}

	// First decode the base station coordinates
	sc, err := decodeStationCoordinates(msg)
	if err != nil {
		return nil, err
	}

	// Start position after the station coordinates (36 + 6 + 6 + 38*3 = 162 bits)
	pos := 162

	// Create station coordinates with height
	sca := &StationCoordinatesAlt{
		StationCoordinates: *sc,
	}

	// Decode antenna height (16 bits, 0.0001 m resolution, unsigned)
	height := gnssgo.GetBitU(msg.Data, pos, 16)
	sca.AntennaHeight = float64(height) * 0.0001

	return sca, nil
}

// decodeAntennaDescriptor decodes RTCM message 1007 (Antenna Descriptor)
func decodeAntennaDescriptor(msg *RTCMMessage) (*AntennaDescriptor, error) {
	if msg == nil || msg.Type != RTCM_ANTENNA_DESCRIPTOR {
		return nil, fmt.Errorf("not an antenna descriptor message")
	}

	if len(msg.Data) < 10 {
		return nil, fmt.Errorf("message too short for antenna descriptor")
	}

	// Start position after message type and station ID (24 + 12 = 36 bits)
	pos := 36

	// Create antenna descriptor
	ad := &AntennaDescriptor{
		StationID: msg.StationID,
	}

	// Decode antenna setup ID
	ad.AntennaSetupID = uint8(gnssgo.GetBitU(msg.Data, pos, 8))
	pos += 8

	// Decode antenna descriptor string length
	length := int(gnssgo.GetBitU(msg.Data, pos, 8))
	pos += 8

	// Decode antenna descriptor string
	var sb strings.Builder
	for i := 0; i < length; i++ {
		ch := byte(gnssgo.GetBitU(msg.Data, pos, 8))
		sb.WriteByte(ch)
		pos += 8
	}
	ad.AntennaType = sb.String()

	return ad, nil
}

// decodeAntennaDescriptorSerial decodes RTCM message 1008 (Antenna Descriptor and Serial Number)
func decodeAntennaDescriptorSerial(msg *RTCMMessage) (*AntennaDescriptorSerial, error) {
	if msg == nil || msg.Type != RTCM_ANTENNA_DESCRIPTOR_SERIAL {
		return nil, fmt.Errorf("not an antenna descriptor and serial number message")
	}

	// First decode the antenna descriptor
	ad, err := decodeAntennaDescriptor(msg)
	if err != nil {
		return nil, err
	}

	// Start position after the antenna descriptor
	pos := 36 + 8 + 8 + len(ad.AntennaType)*8

	// Create antenna descriptor and serial number
	ads := &AntennaDescriptorSerial{
		AntennaDescriptor: *ad,
	}

	// Decode antenna serial number string length
	length := int(gnssgo.GetBitU(msg.Data, pos, 8))
	pos += 8

	// Decode antenna serial number string
	var sb strings.Builder
	for i := 0; i < length; i++ {
		ch := byte(gnssgo.GetBitU(msg.Data, pos, 8))
		sb.WriteByte(ch)
		pos += 8
	}
	ads.AntennaSerial = sb.String()

	return ads, nil
}

// decodeReceiverInfo decodes RTCM message 1033 (Receiver and Antenna Descriptor)
func decodeReceiverInfo(msg *RTCMMessage) (*ReceiverInfo, error) {
	if msg == nil || msg.Type != RTCM_RECEIVER_INFO {
		return nil, fmt.Errorf("not a receiver info message")
	}

	if len(msg.Data) < 10 {
		return nil, fmt.Errorf("message too short for receiver info")
	}

	// Start position after message type and station ID (24 + 12 = 36 bits)
	pos := 36

	// Create receiver info
	ri := &ReceiverInfo{
		StationID: msg.StationID,
	}

	// Decode receiver type descriptor string length
	length := int(gnssgo.GetBitU(msg.Data, pos, 8))
	pos += 8

	// Decode receiver type descriptor string
	var sb strings.Builder
	for i := 0; i < length; i++ {
		ch := byte(gnssgo.GetBitU(msg.Data, pos, 8))
		sb.WriteByte(ch)
		pos += 8
	}
	ri.ReceiverType = sb.String()

	// Decode receiver firmware version string length
	length = int(gnssgo.GetBitU(msg.Data, pos, 8))
	pos += 8

	// Decode receiver firmware version string
	sb.Reset()
	for i := 0; i < length; i++ {
		ch := byte(gnssgo.GetBitU(msg.Data, pos, 8))
		sb.WriteByte(ch)
		pos += 8
	}
	ri.ReceiverFirmware = sb.String()

	// Decode receiver serial number string length
	length = int(gnssgo.GetBitU(msg.Data, pos, 8))
	pos += 8

	// Decode receiver serial number string
	sb.Reset()
	for i := 0; i < length; i++ {
		ch := byte(gnssgo.GetBitU(msg.Data, pos, 8))
		sb.WriteByte(ch)
		pos += 8
	}
	ri.ReceiverSerial = sb.String()

	// Decode antenna type descriptor string length
	length = int(gnssgo.GetBitU(msg.Data, pos, 8))
	pos += 8

	// Decode antenna type descriptor string
	sb.Reset()
	for i := 0; i < length; i++ {
		ch := byte(gnssgo.GetBitU(msg.Data, pos, 8))
		sb.WriteByte(ch)
		pos += 8
	}
	ri.AntennaType = sb.String()

	// Decode antenna serial number string length
	length = int(gnssgo.GetBitU(msg.Data, pos, 8))
	pos += 8

	// Decode antenna serial number string
	sb.Reset()
	for i := 0; i < length; i++ {
		ch := byte(gnssgo.GetBitU(msg.Data, pos, 8))
		sb.WriteByte(ch)
		pos += 8
	}
	ri.AntennaSerial = sb.String()

	// Decode antenna setup ID
	ri.AntennaSetupID = uint8(gnssgo.GetBitU(msg.Data, pos, 8))

	return ri, nil
}
