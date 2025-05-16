/*
* converter.go : Core functionality for converting GNSS binary files to RINEX format
*
* This package provides the core functionality of the convbin tool, separated from
* the command-line interface to allow for use as an imported library.
*/

package converter

import (
	"fmt"
	"os"
	"strings"

	"github.com/bramburn/gnssgo/pkg/gnssgo"
)

// NOUTFILE is the number of output files
const NOUTFILE = 9

// Convert performs the conversion from a binary GNSS format to RINEX files
// format: input format type (STRFMT_???)
// opt: RINEX options
// ifile: input file path
// ofiles: output file paths array (length must be NOUTFILE)
// dir: output directory (if empty, same as input file)
// Returns: status (0:error, 1:ok)
func Convert(format int, opt *gnssgo.RnxOpt, ifile string, ofiles []string, dir string) int {
	var (
		i, def         int
		work, ifile_   string
		ofile          [NOUTFILE]string = [NOUTFILE]string{"", "", "", "", "", "", "", "", ""}
		extnav, extlog string           = "P", "sbs"
	)
	if opt.RnxVer <= 299 || opt.NavSys == gnssgo.SYS_GPS {
		extnav = "N"
	}

	// Copy input ofiles to local ofile array
	for i = 0; i < NOUTFILE && i < len(ofiles); i++ {
		ofile[i] = ofiles[i]
	}

	/* replace wild-card (*) in input file by 0 */
	ifile_ = ifile
	ifile_ = strings.Replace(ifile_, "*", "0", -1)

	for i = range ofile {
		if len(ofile[i]) == 0 {
			def++
		}
	}
	if def >= 8 && i >= 8 {
		def = 1
	} else {
		def = 0
	}

	// Set default output file names if not provided
	switch {
	case len(ofile[0]) > 0:
		// Use provided name
	case len(opt.Staid) > 0:
		ofile[0] = "%r%n0.%yO"
	case def > 0:
		ofile[0] = ifile_
		if idx := strings.LastIndex(ofile[0], "."); idx >= 0 {
			ofile[0] = ofile[0][:idx] + ".obs"
		} else {
			ofile[0] += ".obs"
		}
	}

	switch {
	case len(ofile[1]) > 0:
		// Use provided name
	case len(opt.Staid) > 0:
		ofile[1] = "%r%n0.%y"
		ofile[1] += extnav
	case def > 0:
		ofile[1] = ifile_
		if idx := strings.LastIndex(ofile[1], "."); idx >= 0 {
			ofile[1] = ofile[1][:idx] + ".nav"
		} else {
			ofile[1] += ".nav"
		}
	}

	switch {
	case len(ofile[2]) > 0:
		// Use provided name
	case opt.RnxVer <= 299 && len(opt.Staid) > 0:
		ofile[2] = "%r%n0.%yG"
	case opt.RnxVer <= 299 && def > 0:
		ofile[2] = ifile_
		if idx := strings.LastIndex(ofile[2], "."); idx >= 0 {
			ofile[2] = ofile[2][:idx] + ".gnav"
		} else {
			ofile[2] += ".gnav"
		}
	}

	switch {
	case len(ofile[3]) > 0:
		// Use provided name
	case opt.RnxVer <= 299 && len(opt.Staid) > 0:
		ofile[3] = "%r%n0.%yH"
	case opt.RnxVer <= 299 && def > 0:
		ofile[3] = ifile_
		if idx := strings.LastIndex(ofile[3], "."); idx >= 0 {
			ofile[3] = ofile[3][:idx] + ".hnav"
		} else {
			ofile[3] += ".hnav"
		}
	}

	switch {
	case len(ofile[4]) > 0:
		// Use provided name
	case opt.RnxVer <= 299 && len(opt.Staid) > 0:
		ofile[4] = "%r%n0.%yQ"
	case opt.RnxVer <= 299 && def > 0:
		ofile[4] = ifile_
		if idx := strings.LastIndex(ofile[4], "."); idx >= 0 {
			ofile[4] = ofile[4][:idx] + ".qnav"
		} else {
			ofile[4] += ".qnav"
		}
	}

	switch {
	case len(ofile[5]) > 0:
		// Use provided name
	case opt.RnxVer <= 299 && len(opt.Staid) > 0:
		ofile[5] = "%r%n0.%yL"
	case opt.RnxVer <= 299 && def > 0:
		ofile[5] = ifile_
		if idx := strings.LastIndex(ofile[5], "."); idx >= 0 {
			ofile[5] = ofile[5][:idx] + ".lnav"
		} else {
			ofile[5] += ".lnav"
		}
	}

	switch {
	case len(ofile[6]) > 0:
		// Use provided name
	case opt.RnxVer <= 299 && len(opt.Staid) > 0:
		ofile[6] = "%r%n0.%yC"
	case opt.RnxVer <= 299 && def > 0:
		ofile[6] = ifile_
		if idx := strings.LastIndex(ofile[6], "."); idx >= 0 {
			ofile[6] = ofile[6][:idx] + ".cnav"
		} else {
			ofile[6] += ".cnav"
		}
	}

	switch {
	case len(ofile[7]) > 0:
		// Use provided name
	case opt.RnxVer <= 299 && len(opt.Staid) > 0:
		ofile[7] = "%r%n0.%yI"
	case opt.RnxVer <= 299 && def > 0:
		ofile[7] = ifile_
		if idx := strings.LastIndex(ofile[7], "."); idx >= 0 {
			ofile[7] = ofile[7][:idx] + ".inav"
		} else {
			ofile[7] += ".inav"
		}
	}

	switch {
	case len(ofile[8]) > 0:
		// Use provided name
	case len(opt.Staid) > 0:
		ofile[8] = "%r%n0_%y."
		ofile[8] += extlog
	case def > 0:
		ofile[8] = ifile_
		if idx := strings.LastIndex(ofile[8], "."); idx >= 0 {
			ofile[8] = ofile[8][:idx] + "."
		} else {
			ofile[8] += "."
		}
		ofile[8] += extlog
	}

	// Apply output directory to file paths if specified
	for i = 0; i < NOUTFILE; i++ {
		if len(dir) == 0 || len(ofile[i]) == 0 {
			continue
		}
		if idx := strings.LastIndex(ofile[i], gnssgo.FILEPATHSEP); idx >= 0 {
			work = ofile[i][idx+1:]
		} else {
			work = ofile[i]
		}
		ofile[i] = fmt.Sprintf("%s%s%s", dir, gnssgo.FILEPATHSEP, work)
	}

	// Log input and output files
	fmt.Fprintf(os.Stderr, "input file  : %s (%s)\n", ifile, gnssgo.FormatStrs[format])

	if len(ofile[0]) > 0 {
		fmt.Fprintf(os.Stderr, ".rinex obs : %s\n", ofile[0])
	}
	if len(ofile[1]) > 0 {
		fmt.Fprintf(os.Stderr, ".rinex nav : %s\n", ofile[1])
	}
	if len(ofile[2]) > 0 {
		fmt.Fprintf(os.Stderr, ".rinex gnav: %s\n", ofile[2])
	}
	if len(ofile[3]) > 0 {
		fmt.Fprintf(os.Stderr, ".rinex hnav: %s\n", ofile[3])
	}
	if len(ofile[4]) > 0 {
		fmt.Fprintf(os.Stderr, ".rinex qnav: %s\n", ofile[4])
	}
	if len(ofile[5]) > 0 {
		fmt.Fprintf(os.Stderr, ".rinex lnav: %s\n", ofile[5])
	}
	if len(ofile[6]) > 0 {
		fmt.Fprintf(os.Stderr, ".rinex cnav: %s\n", ofile[6])
	}
	if len(ofile[7]) > 0 {
		fmt.Fprintf(os.Stderr, ".rinex inav: %s\n", ofile[7])
	}
	if len(ofile[8]) > 0 {
		fmt.Fprintf(os.Stderr, ".sbas log  : %s\n", ofile[8])
	}

	// Perform the actual conversion
	if gnssgo.ConvRnx(format, opt, ifile, ofile[:]) == 0 {
		fmt.Fprintf(os.Stderr, "\n")
		return 0
	}
	fmt.Fprintf(os.Stderr, "\n")
	return 1
}
