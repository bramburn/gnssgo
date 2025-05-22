package main

import (
	"time"
)

// RTKStatus represents the current RTK status
type RTKStatus struct {
	Status    string    // RTK status (NONE, SINGLE, FLOAT, FIX)
	Time      time.Time // Time of the status
	Latitude  float64   // Latitude in degrees
	Longitude float64   // Longitude in degrees
	Altitude  float64   // Altitude in meters
	NSats     int       // Number of satellites
	HDOP      float64   // Horizontal dilution of precision
	Age       float64   // Age of differential (seconds)
}
