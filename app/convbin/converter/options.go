/*
* options.go : Options handling for GNSS binary files
*
* This file provides functions for handling command-line options.
*/

package converter

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/bramburn/gnssgo/pkg/gnssgo"
)

// TimeFlag is a flag for time parameters
type TimeFlag struct {
	time       *gnssgo.Gtime
	configured bool
}

// Set implements the flag.Value interface
func (f *TimeFlag) Set(s string) error {
	var es []float64 = []float64{2000, 1, 1, 0, 0, 0}
	n, _ := fmt.Sscanf(s, "%f/%f/%f,%f:%f:%f", &es[0], &es[1], &es[2], &es[3], &es[4], &es[5])
	if n < 6 {
		return fmt.Errorf("too few argument")
	}
	*(f.time) = gnssgo.Epoch2Time(es)
	f.configured = true
	return nil
}

// String implements the flag.Value interface
func (f *TimeFlag) String() string {
	return "2000/1/1,0:0:0"
}

// NewGtime creates a new TimeFlag
func NewGtime(p *gnssgo.Gtime) *TimeFlag {
	tf := TimeFlag{p, false}
	return &tf
}

// PosFlag is a flag for position parameters
type PosFlag struct {
	pos        []float64
	configured bool
}

// Set implements the flag.Value interface
func (f *PosFlag) Set(s string) error {
	values := strings.Split(s, "/")
	if len(values) < 3 {
		return fmt.Errorf("too few arguments")
	}
	for i, v := range values {
		PosFlag(*f).pos[i], _ = strconv.ParseFloat(v, 64)
	}
	f.configured = true
	return nil
}

// String implements the flag.Value interface
func (f *PosFlag) String() string {
	return "0,0,0"
}

// NewFloatSlice creates a new PosFlag
func NewFloatSlice(value []float64, p *[]float64) *PosFlag {
	pf := PosFlag{value, false}
	*p = pf.pos
	return &pf
}

// ArrayFlags is a flag for array parameters
type ArrayFlags []string

// String implements the flag.Value interface
func (i *ArrayFlags) String() string {
	return ""
}

// Set implements the flag.Value interface
func (i *ArrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}
