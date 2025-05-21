package nmea

import (
	"fmt"
	"testing"
)

func TestSimpleNMEAParsing(t *testing.T) {
	// Test with a valid NMEA sentence
	sentence := "$GPGGA,123519,4807.038,N,01131.000,E,1,08,0.9,545.4,M,46.9,M,,*47"
	result, err := ParseNMEA(sentence)
	
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	
	if !result.Valid {
		t.Error("Expected valid result")
	}
	
	if result.Type != "GPGGA" {
		t.Errorf("Expected type GPGGA, got %s", result.Type)
	}
	
	fmt.Println("NMEA parsing test passed")
}
