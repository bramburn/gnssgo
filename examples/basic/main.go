package main

import (
	"fmt"

	"github.com/bramburn/gnssgo/pkg/gnssgo"
)

// Basic example showing how to use the GNSSGO library
func main() {
	// Print the GNSSGO version
	fmt.Printf("GNSSGO Version: %s\n", gnssgo.VER_GNSSGO)

	// Create a time object
	var time gnssgo.Gtime
	var ep [6]float64 = [6]float64{2023, 1, 1, 0, 0, 0}
	time = gnssgo.Epoch2Time(ep[:])

	// Convert time to string
	var timeStr string
	gnssgo.Time2Str(time, &timeStr, 0)
	fmt.Printf("Time: %s\n", timeStr)

	// Create a position in ECEF coordinates
	var ecef [3]float64 = [3]float64{-2432174.0, 4799596.0, 3360475.0} // Example ECEF coordinates

	// Convert ECEF to geodetic coordinates (lat, lon, height)
	var pos [3]float64
	gnssgo.Ecef2Pos(ecef[:], pos[:])

	// Print the position in degrees
	fmt.Printf("Position (lat, lon, height): %.6f°, %.6f°, %.3f m\n",
		pos[0]*gnssgo.R2D, // Convert radians to degrees
		pos[1]*gnssgo.R2D, // Convert radians to degrees
		pos[2])
}
