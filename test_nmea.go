package main

import (
	"fmt"

	"github.com/bramburn/gnssgo/pkg/gnssgo"
	"github.com/bramburn/gnssgo/pkg/gnssgo/nmea"
)

func main() {
	// Test the original NMEA parser (compatibility wrapper)
	sentence := "$GPGGA,123519,4807.038,N,01131.000,E,1,08,0.9,545.4,M,46.9,M,,*47"
	result1, err := gnssgo.ParseNMEA(sentence)
	if err != nil {
		fmt.Printf("Error parsing NMEA with compatibility wrapper: %v\n", err)
	} else {
		fmt.Printf("Successfully parsed NMEA with compatibility wrapper: %s\n", result1.Type)
	}

	// Test the new NMEA parser
	result2, err := nmea.ParseNMEA(sentence)
	if err != nil {
		fmt.Printf("Error parsing NMEA with new implementation: %v\n", err)
	} else {
		fmt.Printf("Successfully parsed NMEA with new implementation: %s\n", result2.Type)
	}

	// Test the parser struct
	parser := nmea.NewNMEAParser()
	result3, err := parser.Parse(sentence)
	if err != nil {
		fmt.Printf("Error parsing NMEA with parser struct: %v\n", err)
	} else {
		fmt.Printf("Successfully parsed NMEA with parser struct: %s\n", result3.Type)
	}
}
