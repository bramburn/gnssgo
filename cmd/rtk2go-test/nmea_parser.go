package main

import (
	"strings"
	"time"

	"github.com/bramburn/gnssgo/pkg/gnssgo"
)

// NMEAParserImpl implements the NMEAParser interface
type NMEAParserImpl struct{}

// NewNMEAParser creates a new NMEA parser
func NewNMEAParser() *NMEAParserImpl {
	return &NMEAParserImpl{}
}

// Parse parses an NMEA sentence
func (p *NMEAParserImpl) Parse(sentence string) (NMEASentence, error) {
	parsed, err := gnssgo.ParseNMEA(sentence)
	if err != nil {
		return NMEASentence{
			Raw:   sentence,
			Valid: false,
		}, err
	}

	return NMEASentence{
		Raw:      parsed.Raw,
		Type:     parsed.Type,
		Fields:   parsed.Fields,
		Valid:    parsed.Valid,
		Checksum: parsed.Checksum,
	}, nil
}

// ParseGGA parses a GGA sentence
func (p *NMEAParserImpl) ParseGGA(sentence string) (GGAData, error) {
	ggaData, err := gnssgo.ParseGGA(sentence)
	if err != nil {
		return GGAData{}, err
	}

	return GGAData{
		Time:      ggaData.Time,
		Latitude:  ggaData.Latitude,
		LatDir:    ggaData.LatDir,
		Longitude: ggaData.Longitude,
		LonDir:    ggaData.LonDir,
		Quality:   ggaData.Quality,
		NumSats:   ggaData.NumSats,
		HDOP:      ggaData.HDOP,
		Altitude:  ggaData.Altitude,
		AltUnit:   ggaData.AltUnit,
		GeoidSep:  ggaData.GeoidSep,
		GeoidUnit: ggaData.GeoidUnit,
		DGPSAge:   ggaData.DGPSAge,
		DGPSStaID: ggaData.DGPSStaID,
	}, nil
}

// FindNMEASentences finds all NMEA sentences in a string
func (p *NMEAParserImpl) FindNMEASentences(data string) []string {
	var sentences []string
	lines := strings.Split(data, "\r\n")

	for _, line := range lines {
		if strings.HasPrefix(line, "$") {
			sentences = append(sentences, line)
		}
	}

	return sentences
}

// GetFixQualityName returns a human-readable name for the fix quality
func GetFixQualityName(quality int) string {
	return gnssgo.GetFixQualityName(quality)
}

// RTKSolution represents an RTK solution
type RTKSolution struct {
	Status    string    // Current RTK status (NONE, SINGLE, FLOAT, FIX)
	Latitude  float64   // Latitude in degrees
	Longitude float64   // Longitude in degrees
	Altitude  float64   // Altitude in meters
	NSats     int       // Number of satellites
	HDOP      float64   // Horizontal dilution of precision
	Age       float64   // Age of differential corrections in seconds
	Time      time.Time // Time of the last update
}
