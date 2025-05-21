package nmea

// NMEAParser is a parser for NMEA sentences
type NMEAParser struct{}

// NewNMEAParser creates a new NMEA parser
func NewNMEAParser() *NMEAParser {
	return &NMEAParser{}
}

// Parse parses an NMEA sentence
func (p *NMEAParser) Parse(sentence string) (NMEASentence, error) {
	return ParseNMEA(sentence)
}

// ParseGGA parses a GGA sentence
func (p *NMEAParser) ParseGGA(sentence string) (GGAData, error) {
	return ParseGGA(sentence)
}

// FindNMEASentences finds all NMEA sentences in a string
func (p *NMEAParser) FindNMEASentences(data string) []string {
	return FindNMEASentences(data)
}
