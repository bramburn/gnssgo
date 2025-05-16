package top708

import (
	"fmt"
	"strings"
)

// NMEAParser parses NMEA sentences
type NMEAParser struct{}

// NewNMEAParser creates a new NMEA parser
func NewNMEAParser() *NMEAParser {
	return &NMEAParser{}
}

// Parse parses an NMEA sentence
func (p *NMEAParser) Parse(sentence string) NMEASentence {
	result := NMEASentence{
		Raw:   sentence,
		Valid: false,
	}

	// Check if the sentence starts with $
	if !strings.HasPrefix(sentence, "$") {
		return result
	}

	// Check if the sentence has a checksum
	parts := strings.Split(sentence, "*")
	if len(parts) != 2 {
		return result
	}

	// Extract the checksum
	result.Checksum = parts[1]

	// Calculate the checksum
	data := parts[0][1:] // Remove the $ prefix
	calculatedChecksum := p.calculateChecksum(data)
	if calculatedChecksum != result.Checksum {
		return result
	}

	// Split the sentence into fields
	fields := strings.Split(parts[0], ",")
	if len(fields) < 1 {
		return result
	}

	// Extract the sentence type
	result.Type = fields[0][1:] // Remove the $ prefix
	result.Fields = fields[1:]
	result.Valid = true

	return result
}

// calculateChecksum calculates the checksum for an NMEA sentence
func (p *NMEAParser) calculateChecksum(data string) string {
	var checksum byte
	for i := 0; i < len(data); i++ {
		checksum ^= data[i]
	}
	return fmt.Sprintf("%02X", checksum)
}

// RTCMParser parses RTCM messages
type RTCMParser struct{}

// NewRTCMParser creates a new RTCM parser
func NewRTCMParser() *RTCMParser {
	return &RTCMParser{}
}

// Parse parses an RTCM message
func (p *RTCMParser) Parse(data []byte) RTCMMessage {
	result := RTCMMessage{
		Raw:   data,
		Valid: false,
	}

	// Check if the message is long enough
	if len(data) < 3 {
		return result
	}

	// Check if the message starts with the RTCM preamble (0xD3)
	if data[0] != 0xD3 {
		return result
	}

	// Extract the message length
	length := int(data[1])<<8 | int(data[2])
	result.Length = length

	// Check if the message is complete
	if len(data) < length+6 {
		return result
	}

	// Extract the message ID
	result.MessageID = int(data[3]&0xFC) >> 2

	// TODO: Add CRC check for RTCM messages

	result.Valid = true
	return result
}

// UBXParser parses UBX messages
type UBXParser struct{}

// NewUBXParser creates a new UBX parser
func NewUBXParser() *UBXParser {
	return &UBXParser{}
}

// Parse parses a UBX message
func (p *UBXParser) Parse(data []byte) UBXMessage {
	result := UBXMessage{
		Raw:   data,
		Valid: false,
	}

	// Check if the message is long enough
	if len(data) < 8 {
		return result
	}

	// Check if the message starts with the UBX header (0xB5 0x62)
	if data[0] != 0xB5 || data[1] != 0x62 {
		return result
	}

	// Extract the message class and ID
	result.Class = data[2]
	result.ID = data[3]

	// Extract the message length
	length := int(data[4]) | int(data[5])<<8
	result.Length = length

	// Check if the message is complete
	if len(data) < length+8 {
		return result
	}

	// Extract the payload
	result.Payload = data[6 : 6+length]

	// Extract the checksum
	result.Checksum = uint16(data[6+length]) | uint16(data[7+length])<<8

	// Calculate the checksum
	calculatedChecksum := p.calculateChecksum(data[2 : 6+length])
	if calculatedChecksum != result.Checksum {
		return result
	}

	result.Valid = true
	return result
}

// calculateChecksum calculates the checksum for a UBX message
func (p *UBXParser) calculateChecksum(data []byte) uint16 {
	var ck_a, ck_b byte
	for _, b := range data {
		ck_a = ck_a + b
		ck_b = ck_b + ck_a
	}
	return uint16(ck_a) | (uint16(ck_b) << 8)
}
