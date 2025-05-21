package main

import (
	"fmt"

	"github.com/bramburn/gnssgo/pkg/gnssgo/nmea"
)

func main() {
	fmt.Println("Testing NMEA parser refactoring...")
	
	// Test the NMEA parser
	sentence := "$GPGGA,123519,4807.038,N,01131.000,E,1,08,0.9,545.4,M,46.9,M,,*47"
	result, err := nmea.ParseNMEA(sentence)
	if err != nil {
		fmt.Printf("Error parsing NMEA: %v\n", err)
	} else {
		fmt.Printf("Successfully parsed NMEA: %s\n", result.Type)
		fmt.Printf("Fields: %v\n", result.Fields)
	}
	
	// Test GGA parsing
	ggaData, err := nmea.ParseGGA(sentence)
	if err != nil {
		fmt.Printf("Error parsing GGA: %v\n", err)
	} else {
		fmt.Printf("Successfully parsed GGA\n")
		fmt.Printf("Latitude: %.6f, Longitude: %.6f\n", ggaData.Latitude, ggaData.Longitude)
		fmt.Printf("Quality: %d (%s)\n", ggaData.Quality, nmea.GetFixQualityName(ggaData.Quality))
	}
}
