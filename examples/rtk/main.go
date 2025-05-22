package main

import (
	"fmt"

	"github.com/bramburn/gnssgo/pkg/gnssgo"
)

// Example demonstrating RTK positioning with GNSSGO
func main() {
	// Initialize RTK control structure
	var rtk gnssgo.Rtk

	// Initialize processing options
	var opt gnssgo.PrcOpt

	// Set default processing options
	opt.Mode = gnssgo.PMODE_KINEMA               // Kinematic mode
	opt.NavSys = gnssgo.SYS_GPS | gnssgo.SYS_GLO // Use GPS and GLONASS
	opt.Elmin = 15.0 * gnssgo.D2R                // 15 degrees elevation mask
	// SNR mask would be configured here in a real application
	opt.ModeAr = 1 // AR mode

	// Set base station position (example coordinates)
	opt.RefPos = gnssgo.POSOPT_SINGLE
	opt.Rb[0] = -2432174.0 // X coordinate (ECEF)
	opt.Rb[1] = 4799596.0  // Y coordinate (ECEF)
	opt.Rb[2] = 3360475.0  // Z coordinate (ECEF)

	// Initialize RTK control with options
	rtk.InitRtk(&opt)

	fmt.Println("RTK initialized with the following settings:")
	fmt.Printf("Mode: %d\n", opt.Mode)
	fmt.Printf("Navigation systems: %d\n", opt.NavSys)
	fmt.Printf("Elevation mask: %.1f degrees\n", opt.Elmin*gnssgo.R2D)

	// In a real application, you would:
	// 1. Read observation data from receivers
	// 2. Process the data with rtk.RtkPos()
	// 3. Get the solution from rtk.RtkSol

	fmt.Println("\nTo use RTK positioning in a real application:")
	fmt.Println("1. Collect observation data from base and rover receivers")
	fmt.Println("2. Call rtk.RtkPos() with the observation data")
	fmt.Println("3. Get the solution from rtk.RtkSol")
}
