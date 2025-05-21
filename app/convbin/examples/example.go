/*
* example.go : Example of using the convbin converter as a library
*
* This example demonstrates how to use the convbin converter functionality
* as an imported library in other Go applications.
 */

package main

import (
	"fmt"
	"os"

	"github.com/bramburn/gnssgo/app/convbin/converter"
	gnssgo "github.com/bramburn/gnssgo/src"
)

func main() {
	// Create conversion options
	opt := gnssgo.RnxOpt{
		RnxVer:  304, // RINEX version 3.04
		ObsType: gnssgo.OBSTYPE_PR | gnssgo.OBSTYPE_CP,
		NavSys:  gnssgo.SYS_GPS | gnssgo.SYS_GLO,
	}

	// Initialize mask
	for i := 0; i < 6; i++ {
		for j := 0; j < 64; j++ {
			opt.Mask[i][j] = '1'
		}
	}

	// Set program name
	opt.Prog = fmt.Sprintf("CONVBIN_EXAMPLE %s", gnssgo.VER_GNSSGO)

	// Set up input and output files
	inputFile := "data.ubx"
	outputFiles := []string{
		"output.obs", // OBS file
		"output.nav", // NAV file
		"",           // GNAV file (not used)
		"",           // HNAV file (not used)
		"",           // QNAV file (not used)
		"",           // LNAV file (not used)
		"",           // CNAV file (not used)
		"",           // INAV file (not used)
		"",           // SBAS file (not used)
	}

	// Detect format from file extension
	format := converter.DetectFormat(inputFile)
	if format < 0 {
		fmt.Fprintf(os.Stderr, "Input format cannot be recognized\n")
		os.Exit(1)
	}

	// Perform conversion
	result := converter.Convert(format, &opt, inputFile, outputFiles, "")

	if result == 0 {
		fmt.Println("Conversion failed")
		os.Exit(1)
	} else {
		fmt.Println("Conversion successful")
	}
}
