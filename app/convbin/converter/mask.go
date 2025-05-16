/*
* mask.go : Signal mask handling for GNSS binary files
*
* This file provides functions for handling signal masks.
*/

package converter

import (
	"strings"

	"github.com/bramburn/gnssgo/pkg/gnssgo"
)

// SetMask sets signal mask in RINEX options
// argv: comma-separated list of signal masks
// opt: RINEX options
// mask: 1 to set mask, 0 to clear mask
func SetMask(argv string, opt *gnssgo.RnxOpt, mask int) {
	var i, code int

	args := strings.Split(argv, ",")
	for _, v := range args {
		if len(v) < 4 || v[1] != 'L' {
			continue
		}
		switch v[0] {
		case 'G':
			i = 0
		case 'R':
			i = 1
		case 'E':
			i = 2
		case 'J':
			i = 3
		case 'S':
			i = 4
		case 'C':
			i = 5
		case 'I':
			i = 6
		default:
			continue
		}
		if code = int(gnssgo.Obs2Code(v[2:])); code > 0 {
			opt.Mask[i][code-1] = '0'
			if mask > 0 {
				opt.Mask[i][code-1] = '1'
			}
		}
	}
}
