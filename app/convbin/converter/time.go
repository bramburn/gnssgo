/*
* time.go : Time handling utilities for GNSS binary files
*
* This file provides functions for handling time-related operations.
*/

package converter

import (
	"fmt"
	"os"
	"strings"

	"github.com/bramburn/gnssgo/pkg/gnssgo"
)

// BytesToUint32 converts a byte slice to uint32
func BytesToUint32(b []byte) uint32 {
	_ = b[3] // bounds check hint to compiler; see golang.org/issue/14808
	return uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16 | uint32(b[3])<<24
}

// GetFileTime gets the start time of an input file
// Returns 1 on success, 0 on failure
func GetFileTime(file string, time *gnssgo.Gtime) int {
	var (
		fp             *os.File
		buff           [64]byte
		ep             [6]float64
		path, path_tag string
		paths          [1]string
	)

	if gnssgo.ExPath(file, paths[:], 1) == 0 {
		return 0
	}
	path = paths[0]

	// Get start time of time-tag file
	path_tag = fmt.Sprintf("%.1019s.tag", path)
	if fp, _ = os.OpenFile(path_tag, os.O_RDONLY, 0666); fp != nil {
		defer fp.Close()
		if n, _ := fp.Read(buff[:1]); n == 1 {
			if strings.Compare(string(buff[:7]), "TIMETAG") == 0 {
				if n, _ = fp.Read(buff[:4]); n == 4 {
					time.Time = uint64(BytesToUint32(buff[:4]))
					time.Sec = 0.0
					return 1
				}
			}
		}
	}

	// Get modified time of input file
	if fp2, _ := os.Open(path); fp2 != nil {
		defer fp2.Close()
		if stat, err := fp2.Stat(); err == nil {
			tm := stat.ModTime()
			ep[0] = float64(tm.Year())
			ep[1] = float64(tm.Month())
			ep[2] = float64(tm.Day())
			ep[3] = float64(tm.Hour())
			ep[4] = float64(tm.Minute())
			ep[5] = float64(tm.Second())
			*time = gnssgo.Utc2GpsT(gnssgo.Epoch2Time(ep[:]))
			return 1
		}
	}
	return 0
}
