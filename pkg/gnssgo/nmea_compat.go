package gnssgo

import (
	"github.com/bramburn/gnssgo/pkg/gnssgo/nmea"
)

// NMEASentence represents a parsed NMEA sentence (compatibility wrapper)
type NMEASentence = nmea.NMEASentence

// GGAData represents parsed GGA sentence data (compatibility wrapper)
type GGAData = nmea.GGAData

// NMEAParser is a parser for NMEA sentences (compatibility wrapper)
type NMEAParser = nmea.NMEAParser

// ParseNMEA parses an NMEA sentence (compatibility wrapper)
func ParseNMEA(sentence string) (NMEASentence, error) {
	return nmea.ParseNMEA(sentence)
}

// CalculateNMEAChecksum calculates the checksum for an NMEA sentence (compatibility wrapper)
func CalculateNMEAChecksum(data string) string {
	return nmea.CalculateNMEAChecksum(data)
}

// ParseGGA parses a GGA sentence (compatibility wrapper)
func ParseGGA(sentence string) (GGAData, error) {
	return nmea.ParseGGA(sentence)
}

// GetFixQualityName returns a string representation of the fix quality (compatibility wrapper)
func GetFixQualityName(quality int) string {
	return nmea.GetFixQualityName(quality)
}

// FindNMEASentences finds all NMEA sentences in a string (compatibility wrapper)
func FindNMEASentences(data string) []string {
	return nmea.FindNMEASentences(data)
}

// NewNMEAParser creates a new NMEA parser (compatibility wrapper)
func NewNMEAParser() *NMEAParser {
	return nmea.NewNMEAParser()
}
