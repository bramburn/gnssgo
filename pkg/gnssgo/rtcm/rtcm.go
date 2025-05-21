// Package rtcm provides functionality for parsing and handling RTCM 3.x messages
// used in GNSS applications for transmitting correction data.
package rtcm

import (
	"errors"
	"fmt"
	"time"

	"github.com/bramburn/gnssgo/pkg/gnssgo"
)

// Constants for RTCM message parsing
const (
	RTCM3PREAMB = 0xD3 // RTCM ver.3 frame preamble

	// Message type ranges
	MSM_GPS_RANGE_START     = 1071 // GPS MSM messages start
	MSM_GPS_RANGE_END       = 1077 // GPS MSM messages end
	MSM_GLONASS_RANGE_START = 1081 // GLONASS MSM messages start
	MSM_GLONASS_RANGE_END   = 1087 // GLONASS MSM messages end
	MSM_GALILEO_RANGE_START = 1091 // Galileo MSM messages start
	MSM_GALILEO_RANGE_END   = 1097 // Galileo MSM messages end
	MSM_SBAS_RANGE_START    = 1101 // SBAS MSM messages start
	MSM_SBAS_RANGE_END      = 1107 // SBAS MSM messages end
	MSM_QZSS_RANGE_START    = 1111 // QZSS MSM messages start
	MSM_QZSS_RANGE_END      = 1117 // QZSS MSM messages end
	MSM_BEIDOU_RANGE_START  = 1121 // BeiDou MSM messages start
	MSM_BEIDOU_RANGE_END    = 1127 // BeiDou MSM messages end
	MSM_IRNSS_RANGE_START   = 1131 // IRNSS MSM messages start
	MSM_IRNSS_RANGE_END     = 1137 // IRNSS MSM messages end

	// SSR message ranges
	SSR_ORBIT_CLOCK_START = 1057 // SSR orbit and clock correction start
	SSR_ORBIT_CLOCK_END   = 1062 // SSR orbit and clock correction end
	SSR_CODE_BIAS_START   = 1063 // SSR code bias start
	SSR_CODE_BIAS_END     = 1068 // SSR code bias end
	SSR_PHASE_BIAS_START  = 1265 // SSR phase bias start
	SSR_PHASE_BIAS_END    = 1270 // SSR phase bias end

	// Station information messages
	RTCM_STATION_COORDINATES       = 1005 // Station coordinates XYZ
	RTCM_STATION_COORDINATES_ALT   = 1006 // Station coordinates XYZ with height
	RTCM_ANTENNA_DESCRIPTOR        = 1007 // Antenna descriptor
	RTCM_ANTENNA_DESCRIPTOR_SERIAL = 1008 // Antenna descriptor and serial number
	RTCM_RECEIVER_INFO             = 1033 // Receiver and antenna descriptor

	// Ephemeris messages
	RTCM_GPS_EPHEMERIS     = 1019 // GPS ephemeris
	RTCM_GLONASS_EPHEMERIS = 1020 // GLONASS ephemeris
	RTCM_GALILEO_EPHEMERIS = 1046 // Galileo ephemeris
	RTCM_BEIDOU_EPHEMERIS  = 1042 // BeiDou ephemeris
	RTCM_QZSS_EPHEMERIS    = 1044 // QZSS ephemeris
)

// Error definitions
var (
	ErrInvalidPreamble    = errors.New("invalid RTCM preamble")
	ErrMessageTooShort    = errors.New("RTCM message too short")
	ErrInvalidCRC         = errors.New("invalid RTCM CRC")
	ErrUnsupportedMessage = errors.New("unsupported RTCM message type")
	ErrIncompleteMessage  = errors.New("incomplete RTCM message")
)

// RTCMMessage represents a parsed RTCM message
type RTCMMessage struct {
	Type      int       // Message type
	Length    int       // Message length (bytes)
	Data      []byte    // Raw message data
	Timestamp time.Time // Time when the message was received
	StationID uint16    // Reference station ID
}

// RTCMParser is responsible for parsing RTCM messages from a byte stream
type RTCMParser struct {
	buffer     []byte                    // Buffer for storing incomplete messages
	messages   []RTCMMessage             // Parsed messages
	stats      map[int]*RTCMMessageStats // Statistics for each message type
	lastUpdate time.Time                 // Time of last update
}

// RTCMMessageStats contains statistics for a specific RTCM message type
type RTCMMessageStats struct {
	MessageType  int       // RTCM message type
	Count        int       // Number of messages received
	LastReceived time.Time // Time of last message
	TotalBytes   int       // Total bytes received for this message type
}

// NewRTCMParser creates a new RTCM parser
func NewRTCMParser() *RTCMParser {
	return &RTCMParser{
		buffer:     make([]byte, 0, 1024),
		messages:   make([]RTCMMessage, 0),
		stats:      make(map[int]*RTCMMessageStats),
		lastUpdate: time.Now(),
	}
}

// ParseRTCMMessage parses RTCM messages from a byte stream
// It returns the parsed messages and any remaining bytes that couldn't be parsed
func (p *RTCMParser) ParseRTCMMessage(data []byte) ([]RTCMMessage, []byte, error) {
	// Append new data to existing buffer
	p.buffer = append(p.buffer, data...)

	// Process messages until we can't find any more complete ones
	var messages []RTCMMessage

	for {
		msg, remaining, err := p.extractMessage(p.buffer)
		if err == ErrIncompleteMessage {
			// Not enough data for a complete message, keep the buffer and wait for more
			p.buffer = remaining
			break
		} else if err != nil {
			// Error parsing message, discard the problematic part and continue
			if len(remaining) > 0 {
				p.buffer = remaining
				continue
			}
			// If no remaining data, clear buffer and return error
			p.buffer = p.buffer[:0]
			return messages, nil, err
		}

		// Successfully parsed a message
		messages = append(messages, msg)
		p.buffer = remaining

		// Update statistics
		p.updateStats(msg)

		// If no more data in buffer, break
		if len(p.buffer) == 0 {
			break
		}
	}

	return messages, p.buffer, nil
}

// extractMessage extracts a single RTCM message from the buffer
func (p *RTCMParser) extractMessage(buffer []byte) (RTCMMessage, []byte, error) {
	// Check if buffer has enough data for header (3 bytes minimum)
	if len(buffer) < 3 {
		return RTCMMessage{}, buffer, ErrIncompleteMessage
	}

	// Check for RTCM preamble
	if buffer[0] != RTCM3PREAMB {
		// Find next preamble
		for i := 1; i < len(buffer); i++ {
			if buffer[i] == RTCM3PREAMB {
				return RTCMMessage{}, buffer[i:], ErrInvalidPreamble
			}
		}
		// No preamble found, discard all data
		return RTCMMessage{}, nil, ErrInvalidPreamble
	}

	// Extract message length (10 bits starting at bit 14)
	msgLength := int(gnssgo.GetBitU(buffer, 14, 10)) + 3 // +3 for header

	// Check if we have the complete message including CRC (message + 3 bytes CRC)
	if len(buffer) < msgLength+3 {
		return RTCMMessage{}, buffer, ErrIncompleteMessage
	}

	// Validate CRC
	if !ValidateCRC(&RTCMMessage{Data: buffer[:msgLength]}) {
		// CRC error, discard this message and try to find next preamble
		for i := 1; i < len(buffer); i++ {
			if buffer[i] == RTCM3PREAMB {
				return RTCMMessage{}, buffer[i:], ErrInvalidCRC
			}
		}
		// No preamble found, discard all data
		return RTCMMessage{}, nil, ErrInvalidCRC
	}

	// Extract message type (12 bits starting at bit 24)
	msgType := int(gnssgo.GetBitU(buffer, 24, 12))

	// Extract station ID (12 bits starting at bit 36)
	stationID := uint16(gnssgo.GetBitU(buffer, 36, 12))

	// Create message
	msg := RTCMMessage{
		Type:      msgType,
		Length:    msgLength,
		Data:      make([]byte, msgLength+3), // Include CRC
		Timestamp: time.Now(),
		StationID: stationID,
	}

	// Copy data
	copy(msg.Data, buffer[:msgLength+3])

	// Return message and remaining buffer
	return msg, buffer[msgLength+3:], nil
}

// updateStats updates the statistics for a message type
func (p *RTCMParser) updateStats(msg RTCMMessage) {
	stats, ok := p.stats[msg.Type]
	if !ok {
		stats = &RTCMMessageStats{
			MessageType: msg.Type,
		}
		p.stats[msg.Type] = stats
	}

	stats.Count++
	stats.LastReceived = msg.Timestamp
	stats.TotalBytes += msg.Length
}

// GetStats returns the statistics for all message types
func (p *RTCMParser) GetStats() map[int]*RTCMMessageStats {
	return p.stats
}

// ValidateCRC validates the CRC of an RTCM message
func ValidateCRC(msg *RTCMMessage) bool {
	if msg == nil || len(msg.Data) < msg.Length+3 {
		return false
	}

	// Calculate CRC
	crc := gnssgo.Rtk_CRC24q(msg.Data[:msg.Length], msg.Length)

	// Extract CRC from message
	msgCRC := gnssgo.GetBitU(msg.Data, msg.Length*8, 24)

	return crc == msgCRC
}

// DecodeRTCMMessage decodes the content of an RTCM message based on its type
func DecodeRTCMMessage(msg *RTCMMessage) (interface{}, error) {
	if msg == nil {
		return nil, errors.New("nil message")
	}

	switch {
	case msg.Type == RTCM_STATION_COORDINATES:
		return decodeStationCoordinates(msg)
	case msg.Type == RTCM_STATION_COORDINATES_ALT:
		return decodeStationCoordinatesAlt(msg)
	case msg.Type == RTCM_GPS_EPHEMERIS:
		return decodeGPSEphemeris(msg)
	case msg.Type == RTCM_GLONASS_EPHEMERIS:
		return decodeGLONASSEphemeris(msg)
	case msg.Type >= MSM_GPS_RANGE_START && msg.Type <= MSM_GPS_RANGE_END:
		return decodeMSMMessage(msg, gnssgo.SYS_GPS)
	case msg.Type >= MSM_GLONASS_RANGE_START && msg.Type <= MSM_GLONASS_RANGE_END:
		return decodeMSMMessage(msg, gnssgo.SYS_GLO)
	case msg.Type >= MSM_GALILEO_RANGE_START && msg.Type <= MSM_GALILEO_RANGE_END:
		return decodeMSMMessage(msg, gnssgo.SYS_GAL)
	case msg.Type >= MSM_BEIDOU_RANGE_START && msg.Type <= MSM_BEIDOU_RANGE_END:
		return decodeMSMMessage(msg, gnssgo.SYS_CMP)
	case msg.Type >= MSM_QZSS_RANGE_START && msg.Type <= MSM_QZSS_RANGE_END:
		return decodeMSMMessage(msg, gnssgo.SYS_QZS)
	case msg.Type >= SSR_ORBIT_CLOCK_START && msg.Type <= SSR_ORBIT_CLOCK_END:
		return decodeSSROrbitClock(msg)
	case msg.Type >= SSR_CODE_BIAS_START && msg.Type <= SSR_CODE_BIAS_END:
		return decodeSSRCodeBias(msg)
	case msg.Type >= SSR_PHASE_BIAS_START && msg.Type <= SSR_PHASE_BIAS_END:
		return decodeSSRPhaseBias(msg)
	default:
		return nil, fmt.Errorf("%w: type %d", ErrUnsupportedMessage, msg.Type)
	}
}

// GetMessageTypeDescription returns a human-readable description of an RTCM message type
func GetMessageTypeDescription(msgType int) string {
	switch {
	case msgType == RTCM_STATION_COORDINATES:
		return "Station Coordinates XYZ"
	case msgType == RTCM_STATION_COORDINATES_ALT:
		return "Station Coordinates XYZ with Height"
	case msgType == RTCM_ANTENNA_DESCRIPTOR:
		return "Antenna Descriptor"
	case msgType == RTCM_ANTENNA_DESCRIPTOR_SERIAL:
		return "Antenna Descriptor and Serial Number"
	case msgType == RTCM_RECEIVER_INFO:
		return "Receiver and Antenna Descriptor"
	case msgType == RTCM_GPS_EPHEMERIS:
		return "GPS Ephemeris"
	case msgType == RTCM_GLONASS_EPHEMERIS:
		return "GLONASS Ephemeris"
	case msgType == RTCM_GALILEO_EPHEMERIS:
		return "Galileo Ephemeris"
	case msgType == RTCM_BEIDOU_EPHEMERIS:
		return "BeiDou Ephemeris"
	case msgType == RTCM_QZSS_EPHEMERIS:
		return "QZSS Ephemeris"
	case msgType >= MSM_GPS_RANGE_START && msgType <= MSM_GPS_RANGE_END:
		return fmt.Sprintf("GPS MSM%d", msgType-MSM_GPS_RANGE_START+1)
	case msgType >= MSM_GLONASS_RANGE_START && msgType <= MSM_GLONASS_RANGE_END:
		return fmt.Sprintf("GLONASS MSM%d", msgType-MSM_GLONASS_RANGE_START+1)
	case msgType >= MSM_GALILEO_RANGE_START && msgType <= MSM_GALILEO_RANGE_END:
		return fmt.Sprintf("Galileo MSM%d", msgType-MSM_GALILEO_RANGE_START+1)
	case msgType >= MSM_BEIDOU_RANGE_START && msgType <= MSM_BEIDOU_RANGE_END:
		return fmt.Sprintf("BeiDou MSM%d", msgType-MSM_BEIDOU_RANGE_START+1)
	case msgType >= MSM_QZSS_RANGE_START && msgType <= MSM_QZSS_RANGE_END:
		return fmt.Sprintf("QZSS MSM%d", msgType-MSM_QZSS_RANGE_START+1)
	case msgType >= SSR_ORBIT_CLOCK_START && msgType <= SSR_ORBIT_CLOCK_END:
		return "SSR Orbit and Clock Correction"
	case msgType >= SSR_CODE_BIAS_START && msgType <= SSR_CODE_BIAS_END:
		return "SSR Code Bias"
	case msgType >= SSR_PHASE_BIAS_START && msgType <= SSR_PHASE_BIAS_END:
		return "SSR Phase Bias"
	default:
		return fmt.Sprintf("Unknown (%d)", msgType)
	}
}
