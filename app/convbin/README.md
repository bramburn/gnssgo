# CONVBIN - GNSS Binary Converter

CONVBIN is a command-line tool for converting GNSS receiver binary log files to RINEX observation and navigation files. It's part of the GNSSGO project, which aims to replicate the functionality of RTKLIB in Go.

## Overview

CONVBIN converts various GNSS receiver binary formats to standard RINEX (Receiver Independent Exchange Format) files, which are widely used for post-processing GNSS data. The tool supports multiple input formats and provides various options for customizing the output.

## Supported Input Formats

- RTCM 2/3
- NovAtel OEM/4/V/6/7, OEMStar, OEM3
- u-blox LEA-4T/5T/6T/7T/M8T/F9
- NovAtel Superstar II
- Hemisphere Eclipse/Crescent
- SkyTraq S1315F
- Javad GREIS
- NVS NV08C BINR
- BINEX
- Trimble RT17
- Septentrio SBF
- RINEX

## Output Files

CONVBIN generates the following output files:
- RINEX observation file (.obs)
- RINEX navigation files (.nav, .gnav, .hnav, .qnav, .lnav, .cnav, .inav)
- SBAS message file (.sbs)

## Command-Line Usage

```
convbin [options] file
```

Where `file` is the input receiver binary log file.

### Basic Example

```
convbin -od -os -oi -ot -f 1 input.ubx
```

This command converts a u-blox binary file to RINEX format with doppler frequency, SNR data, and ionospheric/time corrections included.

### Common Options

- `-ts y/m/d,h:m:s` : Start time
- `-te y/m/d,h:m:s` : End time
- `-ti tint` : Observation data interval (s)
- `-r format` : Input format type (rtcm2, rtcm3, nov, oem3, ubx, ss2, hemis, stq, javad, nvs, binex, rt17, sbf, rinex)
- `-f freq` : Number of frequencies
- `-v ver` : RINEX version
- `-d dir` : Output directory
- `-o ofile` : Output RINEX OBS file
- `-n nfile` : Output RINEX NAV file
- `-trace level` : Output trace level

For a complete list of options, run `convbin` without arguments.

## Using as a Library

The CONVBIN functionality can also be used as an imported library in other Go applications. The core conversion functionality has been separated from the command-line interface to make this possible.

### Example Usage as a Library

```go
package main

import (
    "fmt"
    
    "github.com/bramburn/gnssgo"
    "github.com/bramburn/gnssgo/app/convbin/converter"
)

func main() {
    // Create conversion options
    opt := gnssgo.RnxOpt{
        RnxVer:  304, // RINEX version 3.04
        ObsType: gnssgo.OBSTYPE_PR | gnssgo.OBSTYPE_CP,
        NavSys:  gnssgo.SYS_GPS | gnssgo.SYS_GLO,
    }
    
    // Set up input and output files
    inputFile := "data.ubx"
    outputFiles := []string{
        "output.obs",  // OBS file
        "output.nav",  // NAV file
        "",            // GNAV file (not used)
        "",            // HNAV file (not used)
        "",            // QNAV file (not used)
        "",            // LNAV file (not used)
        "",            // CNAV file (not used)
        "",            // INAV file (not used)
        "",            // SBAS file (not used)
    }
    
    // Detect format from file extension
    format := converter.DetectFormat(inputFile)
    
    // Perform conversion
    result := converter.Convert(format, &opt, inputFile, outputFiles, "")
    
    if result == 0 {
        fmt.Println("Conversion successful")
    } else {
        fmt.Println("Conversion failed")
    }
}
```

## References

This tool is inspired by and aims to replicate the functionality of the CONVBIN tool from RTKLIB, an open-source GNSS processing software developed by Tomoji Takasu. For more information about RTKLIB, visit [RTKLIB's website](http://www.rtklib.com/).

## License

This project is part of GNSSGO, which is licensed under the same terms as the original RTKLIB.
