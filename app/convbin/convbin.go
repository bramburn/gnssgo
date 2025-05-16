/*------------------------------------------------------------------------------
* convbin.go : convert receiver binary log file to rinex obs/nav, sbas messages
*
* This is a Go implementation of the RTKLIB convbin tool, which converts
* GNSS receiver binary log files to RINEX observation and navigation files.
*
* The code has been refactored to separate the core functionality from the
* command-line interface, making it usable both as a standalone executable
* and as an imported library.
*
* Original RTKLIB version by T.TAKASU
* Go implementation by GNSSGO project
*-----------------------------------------------------------------------------*/

package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/bramburn/gnssgo/app/convbin/converter"
	"github.com/bramburn/gnssgo/pkg/gnssgo"
)

const (
	PRGNAME   = "CONVBIN"
	TRACEFILE = "convbin.trace"
	NOUTFILE  = 9 /* number of output files */
)

/* help text -----------------------------------------------------------------*/
var help []string = []string{
	"",
	" Synopsys",
	"",
	" convbin [option ...] file",
	"",
	" Description",
	"",
	" Convert RTCM, receiver raw data log and RINEX file to RINEX and SBAS/LEX",
	" message file. SBAS message file complies with RTKLIB SBAS/LEX message",
	" format. It supports the following messages or files.",
	"",
	" RTCM 2                : Type 1, 3, 9, 14, 16, 17, 18, 19, 22",
	" RTCM 3                : Type 1002, 1004, 1005, 1006, 1010, 1012, 1019, 1020",
	"                         Type 1071-1127 (MSM except for compact msg)",
	" NovAtel OEMV/4,OEMStar: RANGECMPB, RANGEB, RAWEPHEMB, IONUTCB, RAWWASSFRAMEB",
	" NovAtel OEM3          : RGEB, REGD, REPB, FRMB, IONB, UTCB",
	" u-blox LEA-4T/5T/6T   : RXM-RAW, RXM-SFRB",
	" NovAtel Superstar II  : ID#20, ID#21, ID#22, ID#23, ID#67",
	" Hemisphere            : BIN76, BIN80, BIN94, BIN95, BIN96",
	" SkyTraq S1315F        : msg0xDD, msg0xE0, msg0xDC",
	" GW10                  : msg0x08, msg0x03, msg0x27, msg0x20",
	" Javad                 : [R*],[r*],[*R],[*r],[P*],[p*],[*P],[*p],[D*],[*d],",
	"                         [E*],[*E],[F*],[TC],[GE],[NE],[EN],[QE],[UO],[IO],",
	"                         [WD]",
	" NVS                   : BINR",
	" BINEX                 : big-endian, regular CRC, forward record (0xE2)",
	"                         0x01-01,0x01-02,0x01-03,0x01-04,0x01-06,0x7f-05",
	" Trimble               : RT17",
	" Septentrio            : SBF",
	" RINEX                 : OBS, NAV, GNAV, HNAV, LNAV, QNAV",
	"",
	" Options [default]",
	"",
	"     file         input receiver binary log file",
	"     -ts y/m/d,h:m:s  start time [all]",
	"     -te y/m/d,h:m:s  end time [all]",
	"     -tr y/m/d,h:m:s  approximated time for RTCM",
	"     -ti tint     observation data interval (s) [all]",
	"     -tt ttol     observation data epoch tolerance (s) [0.005]",
	"     -span span   time span (h) [all]",
	"     -r format    log format type",
	"                  rtcm2= RTCM 2",
	"                  rtcm3= RTCM 3",
	"                  nov  = NovAtel OEM/4/V/6/7,OEMStar",
	"                  oem3 = NovAtel OEM3",
	"                  ubx  = ublox LEA-4T/5T/6T/7T/M8T/F9",
	"                  ss2  = NovAtel Superstar II",
	"                  hemis= Hemisphere Eclipse/Crescent",
	"                  stq  = SkyTraq S1315F",
	"                  javad= Javad GREIS",
	"                  nvs  = NVS NV08C BINR",
	"                  binex= BINEX",
	"                  rt17 = Trimble RT17",
	"                  sbf  = Septentrio SBF",
	"                  rinex= RINEX",
	"     -ro opt      receiver options",
	"     -f freq      number of frequencies [5]",
	"     -hc comment  rinex header: comment line",
	"     -hm marker   rinex header: marker name",
	"     -hn markno   rinex header: marker number",
	"     -ht marktype rinex header: marker type",
	"     -ho observ   rinex header: oberver name and agency separated by /",
	"     -hr rec      rinex header: receiver number, type and version separated by /",
	"     -ha ant      rinex header: antenna number and type separated by /",
	"     -hp pos      rinex header: approx position x/y/z separated by /",
	"     -hd delta    rinex header: antenna delta h/e/n separated by /",
	"     -v ver       rinex version [3.04]",
	"     -od          include doppler frequency in rinex obs [on]",
	"     -os          include snr in rinex obs [on]",
	"     -oi          include iono correction in rinex nav header [off]",
	"     -ot          include time correction in rinex nav header [off]",
	"     -ol          include leap seconds in rinex nav header [off]",
	"     -halfc       half-cycle ambiguity correction [off]",
	"     -mask   [sig[,...]] signal mask(s) (sig={G|R|E|J|S|C|I}L{1C|1P|1W|...})",
	"     -nomask [sig[,...]] signal no mask (same as above)",
	"     -x sat       exclude satellite",
	"     -y sys       exclude systems (G:GPS,R:GLO,E:GAL,J:QZS,S:SBS,C:BDS,I:IRN)",
	"     -d dir       output directory [same as input file]",
	"     -c staid     use RINEX file name convention with staid [off]",
	"     -o ofile     output RINEX OBS file",
	"     -n nfile     output RINEX NAV file",
	"     -g gfile     output RINEX GNAV file",
	"     -h hfile     output RINEX HNAV file",
	"     -q qfile     output RINEX QNAV file",
	"     -l lfile     output RINEX LNAV file",
	"     -b cfile     output RINEX CNAV file",
	"     -i ifile     output RINEX INAV file",
	"     -s sfile     output SBAS message file",
	"     -trace level output trace level [off]",
	"",
	" If any output file specified, default output files (<file>.obs,",
	" <file>.nav, <file>.gnav, <file>.hnav, <file>.qnav, <file>.lnav,",
	" <file>.cnav, <file>.inav and <file>.sbs) are used. To obtain week number info",
	" for RTCM file, use -tr option to specify the approximated log start time.",
	" Without -tr option, the program obtains the week number from the time-tag file (if it exists) or the last modified time of the log file instead.",
	"",
	" If receiver type is not specified, type is recognized by the input",
	" file extension as follows.",
	"     *.rtcm2       RTCM 2",
	"     *.rtcm3       RTCM 3",
	"     *.gps         NovAtel OEM4/V/6/7,OEMStar",
	"     *.ubx         u-blox LEA-4T/5T/6T/7T/M8T/F9",
	"     *.log         NovAtel Superstar II",
	"     *.bin         Hemisphere Eclipse/Crescent",
	"     *.stq         SkyTraq S1315F",
	"     *.jps         Javad GREIS",
	"     *.bnx,*binex  BINEX",
	"     *.rt17        Trimble RT17",
	"     *.sbf         Septentrio SBF",
	"     *.obs,*.*o    RINEX OBS",
	"     *.rnx         RINEX OBS",
	"     *.nav,*.*n    RINEX NAV"}

/* print help ----------------------------------------------------------------*/
func printhelp() {
	for i := range help {
		fmt.Fprintf(os.Stderr, "%s\n", help[i])
	}
	os.Exit(0)
}

/* convert main --------------------------------------------------------------*/
func convbin(format int, opt *gnssgo.RnxOpt, ifile string, file []string, dir string) int {
	// Call the converter package's Convert function
	result := converter.Convert(format, opt, ifile, file, dir)

	// Convert result from 0/1 to -1/0 to maintain compatibility with original code
	if result == 0 {
		return -1
	}
	return 0
}

/* set signal mask -----------------------------------------------------------*/
func setmask(argv string, opt *gnssgo.RnxOpt, mask int) {
	converter.SetMask(argv, opt, mask)
}

/* get start time of input file -----------------------------------------------*/
func get_filetime(file string, time *gnssgo.Gtime) int {
	return converter.GetFileTime(file, time)
}

func searchHelp(key string) string {
	for _, v := range help {
		if strings.Contains(v, key) {
			return v
		}
	}
	return "no supported argument"
}

// Use the flag types from the converter package
var nc int = 2

/* parse command line options ------------------------------------------------*/
func cmdopts(opt *gnssgo.RnxOpt, ifile *string, ofile []string, dir *string, trace *int) int {
	var (
		span, ver                                         float64
		i, j, k, sat                                      int
		nf, format                                        int = 5, -1
		fmts, sys, names, recs, ants, satid, mask, nomask string
		bod, bos, boi, bot, bol, bscan, bhalfc            bool
	)
	opt.RnxVer = 304
	opt.ObsType = gnssgo.OBSTYPE_PR | gnssgo.OBSTYPE_CP
	opt.NavSys = gnssgo.SYS_GPS | gnssgo.SYS_GLO | gnssgo.SYS_GAL | gnssgo.SYS_QZS | gnssgo.SYS_SBS | gnssgo.SYS_CMP | gnssgo.SYS_IRN

	for i = 0; i < 6; i++ {
		for j = 0; j < 64; j++ {
			opt.Mask[i][j] = '1'
		}
	}
	flag.Var(converter.NewGtime(&opt.TS), "ts", searchHelp("-ts"))
	flag.Var(converter.NewGtime(&opt.TE), "te", searchHelp("-te"))
	flag.Var(converter.NewGtime(&opt.TRtcm), "tr", searchHelp("-tr"))
	flag.Float64Var(&opt.TInt, "ti", opt.TInt, searchHelp("-ti"))
	flag.Float64Var(&opt.TTol, "tt", opt.TTol, searchHelp("-tt"))
	flag.Float64Var(&span, "span", span, searchHelp("-span"))
	flag.StringVar(&fmts, "r", fmts, searchHelp("-r"))
	flag.StringVar(&opt.RcvOpt, "ro", opt.RcvOpt, searchHelp("-ro"))
	flag.IntVar(&nf, "f", nf, searchHelp("-f"))
	var comments converter.ArrayFlags
	flag.Var(&comments, "hc", searchHelp("-hc"))
	flag.StringVar(&opt.Marker, "hm", opt.Marker, searchHelp("-hm"))
	flag.StringVar(&opt.MarkerNo, "hn", opt.MarkerNo, searchHelp("-hn"))
	flag.StringVar(&opt.MarkerType, "ht", opt.MarkerType, searchHelp("-ht"))

	flag.StringVar(&names, "ho", names, searchHelp("-ho"))
	flag.StringVar(&recs, "hr", recs, searchHelp("-hr"))
	flag.StringVar(&ants, "ha", ants, searchHelp("-ha"))

	rp := opt.AppPos[:]
	rpFlag := converter.NewFloatSlice([]float64{}, &rp)
	flag.Var(rpFlag, "hp", searchHelp("-hp"))

	rd := opt.AntDel[:]
	rdFlag := converter.NewFloatSlice([]float64{}, &rd)
	flag.Var(rdFlag, "hd", searchHelp("-hd"))

	flag.Float64Var(&ver, "v", ver, searchHelp("-v"))

	flag.BoolVar(&bod, "od", bod, searchHelp("-od"))
	flag.BoolVar(&bos, "os", bos, searchHelp("-os"))
	flag.BoolVar(&boi, "oi", boi, searchHelp("-oi"))
	flag.BoolVar(&bot, "ot", bot, searchHelp("-ot"))
	flag.BoolVar(&bol, "ol", bol, searchHelp("-ol"))
	flag.BoolVar(&bscan, "scan", bscan, searchHelp("-scan"))
	flag.BoolVar(&bhalfc, "halfc", bod, searchHelp("-halfc"))
	flag.StringVar(&nomask, "nomask", nomask, searchHelp("-nomask"))
	flag.StringVar(&mask, "mask", mask, searchHelp("-mask"))

	flag.StringVar(&satid, "x", satid, searchHelp("-x"))
	flag.StringVar(&sys, "y", sys, searchHelp("-y"))
	flag.StringVar(dir, "d", *dir, searchHelp("-d"))
	flag.StringVar(&opt.Staid, "c", opt.Staid, searchHelp("-c"))

	flag.StringVar(&ofile[0], "o", ofile[0], searchHelp("-o"))
	flag.StringVar(&ofile[1], "n", ofile[1], searchHelp("-n"))
	flag.StringVar(&ofile[2], "g", ofile[2], searchHelp("-g"))
	flag.StringVar(&ofile[3], "h", ofile[3], searchHelp("-h"))
	flag.StringVar(&ofile[4], "q", ofile[4], searchHelp("-q"))
	flag.StringVar(&ofile[5], "l", ofile[5], searchHelp("-l"))
	flag.StringVar(&ofile[6], "b", ofile[6], searchHelp("-b"))
	flag.StringVar(&ofile[7], "i", ofile[7], searchHelp("-i"))
	flag.StringVar(&ofile[8], "s", ofile[8], searchHelp("-s"))

	flag.IntVar(trace, "trace", *trace, searchHelp("-trace"))

	flag.Parse()

	if len(comments) > 0 {
		for _, v := range comments {
			if nc < gnssgo.MAXCOMMENT {
				opt.Comment[nc] = v
				nc++
			}
		}
	}

	if len(names) > 0 {
		p := strings.Split(names, "/")
		for j = range p {
			opt.Name[j] = p[j]
			if j > 1 {
				break
			}
		}
	}

	if len(recs) > 0 {
		p := strings.Split(recs, "/")
		for j = range p {
			opt.Rec[j] = p[j]
			if j > 2 {
				break
			}
		}
	}

	if len(ants) > 0 {
		p := strings.Split(ants, "/")
		for j = range p {
			opt.Ant[j] = p[j]
			if j > 2 {
				break
			}
		}
	}

	if ver > 0 {
		opt.RnxVer = int(ver * 100.0)
	}

	if bod {
		opt.ObsType |= gnssgo.OBSTYPE_DOP
	}

	if bos {
		opt.ObsType |= gnssgo.OBSTYPE_SNR
	}
	if boi {
		opt.Outiono = 1
	}

	if bot {
		opt.OutputTime = 1
	}

	if bol {
		opt.Outleaps = 1
	}

	// if bscan {

	// }

	if bhalfc {
		opt.Halfcyc = 1
	}

	if len(mask) > 0 {
		for j = 0; j < 6; j++ {
			for k = 0; k < 64; k++ {
				opt.Mask[j][k] = '0'
			}
		}
		setmask(mask, opt, 1)
	}

	if len(nomask) > 0 {
		setmask(nomask, opt, 0)
	}

	if len(satid) > 0 {
		if sat = gnssgo.SatId2No(satid); sat > 0 {
			opt.ExSats[sat-1] = 1
		}
	}

	if len(sys) > 0 {
		switch sys[0] {
		case 'G':
			opt.NavSys &= ^gnssgo.SYS_GPS
		case 'R':
			opt.NavSys &= ^gnssgo.SYS_GLO
		case 'E':
			opt.NavSys &= ^gnssgo.SYS_GAL
		case 'J':
			opt.NavSys &= ^gnssgo.SYS_QZS
		case 'S':
			opt.NavSys &= ^gnssgo.SYS_SBS
		case 'C':
			opt.NavSys &= ^gnssgo.SYS_CMP
		case 'I':
			opt.NavSys &= ^gnssgo.SYS_IRN
		}
	}

	if flag.NFlag() < 1 {
		// if there is not any arguments, exit
		printhelp()
		os.Exit(0)
	}
	args := flag.CommandLine.Args()
	if len(args) > 0 {
		*ifile = args[0]
	}

	if span > 0.0 && opt.TStart.Time > 0 {
		opt.TEnd = gnssgo.TimeAdd(opt.TStart, span*3600.0-1e-3)
	}
	if nf >= 1 {
		opt.FreqType |= gnssgo.FREQTYPE_L1
	}
	if nf >= 2 {
		opt.FreqType |= gnssgo.FREQTYPE_L2
	}
	if nf >= 3 {
		opt.FreqType |= gnssgo.FREQTYPE_L3
	}
	if nf >= 4 {
		opt.FreqType |= gnssgo.FREQTYPE_L4
	}
	if nf >= 5 {
		opt.FreqType |= gnssgo.FREQTYPE_L5
	}

	if opt.TRtcm.Time == 0 {
		get_filetime(*ifile, &opt.TRtcm)
	}

	if len(fmts) > 0 {
		format = converter.FormatFromString(fmts)
	} else {
		format = converter.DetectFormat(*ifile)
	}

	return format
}

/* show message --------------------------------------------------------------*/
func showmsg(format string, v ...interface{}) int {
	fmt.Fprintf(os.Stderr, format, v...)
	if len(format) > 0 {
		fmt.Fprintf(os.Stderr, "\r")
	} else {
		fmt.Fprintf(os.Stderr, "\n")
	}
	return 0
}

/* main ----------------------------------------------------------------------*/
func main() {
	var (
		opt                 gnssgo.RnxOpt
		format, trace, stat int
		ifile, dir          string
		ofile               [NOUTFILE]string
	)
	// Parse command line options
	format = cmdopts(&opt, &ifile, ofile[:], &dir, &trace)

	// Validate input
	if len(ifile) == 0 {
		fmt.Fprintf(os.Stderr, "no input file\n")
		os.Exit(-1)
	}
	if format < 0 {
		fmt.Fprintf(os.Stderr, "input format can not be recognized\n")
		os.Exit(-1)
	}

	// Set program name in options
	opt.Prog = fmt.Sprintf("%s %s", PRGNAME, gnssgo.VER_GNSSGO)

	// Initialize tracing if requested
	if trace > 0 {
		gnssgo.TraceOpen(TRACEFILE)
		gnssgo.TraceLevel(trace)
	}

	// Set message handler
	gnssgo.ShowMsg_Ptr = showmsg

	// Perform the conversion
	stat = convbin(format, &opt, ifile, ofile[:], dir)

	// Clean up
	gnssgo.TraceClose()

	os.Exit(stat)
}
