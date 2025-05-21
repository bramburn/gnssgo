// Package gtime provides time-related functionality for GNSS applications
package gtime

import (
	"fmt"
	"time"
)

// Gtime represents a GNSS time
type Gtime struct {
	Time int64   // Time (s) expressed by standard time_t
	Sec  float64 // Fraction of second (s)
}

// Constants for time conversion
const (
	SECONDS_IN_WEEK = 604800.0
	SECONDS_IN_DAY  = 86400.0
	GPS_EPOCH       = 315964800 // GPS time reference epoch (1980/1/6 00:00:00 UTC)
)

// TimeGet returns the current time
func TimeGet() Gtime {
	var ep [6]float64

	// Get current time
	t := time.Now().UTC()

	// Convert to epoch
	ep[0] = float64(t.Year())
	ep[1] = float64(t.Month())
	ep[2] = float64(t.Day())
	ep[3] = float64(t.Hour())
	ep[4] = float64(t.Minute())
	ep[5] = float64(t.Second()) + float64(t.Nanosecond())/1e9

	// Convert to Gtime
	return Epoch2Time(ep)
}

// Epoch2Time converts epoch to Gtime
func Epoch2Time(ep [6]float64) Gtime {
	var (
		time Gtime
		days int64
		sec  float64
	)

	// Calculate days and seconds
	days = (int64(ep[0])-1970)*365 + (int64(ep[0])-1969)/4 + int64(ep[2]) - 1

	for i := 1; i < int(ep[1]); i++ {
		days += int64(DaysInMonth(int(ep[0]), i))
	}

	sec = float64(days)*SECONDS_IN_DAY + ep[3]*3600.0 + ep[4]*60.0 + ep[5]

	time.Time = int64(sec)
	time.Sec = sec - float64(time.Time)

	return time
}

// DaysInMonth returns the number of days in a month
func DaysInMonth(year, month int) int {
	switch month {
	case 2:
		if (year%4 == 0 && year%100 != 0) || year%400 == 0 {
			return 29
		}
		return 28
	case 4, 6, 9, 11:
		return 30
	default:
		return 31
	}
}

// Utc2GpsT converts UTC time to GPS time
func Utc2GpsT(t Gtime) Gtime {
	var tu Gtime

	tu.Time = t.Time + GPS_EPOCH
	tu.Sec = t.Sec

	return tu
}

// Time2GpsT converts time to GPS time of week
func Time2GpsT(t Gtime, week *int) float64 {
	var (
		sec float64
		w   int
	)

	sec = float64(t.Time-GPS_EPOCH) + t.Sec
	w = int(sec / SECONDS_IN_WEEK)
	sec -= float64(w) * SECONDS_IN_WEEK

	if week != nil {
		*week = w
	}

	return sec
}

// TimeStr converts time to string
func TimeStr(t Gtime, n int) string {
	if t.Time == 0 {
		return "0000/00/00 00:00:00.000000000"
	}

	// Convert to time.Time
	tm := time.Unix(t.Time, int64(t.Sec*1e9))

	// Format based on precision
	switch n {
	case 0:
		return tm.Format("2006/01/02 15:04:05.000000000")
	case 1:
		return tm.Format("2006/01/02 15:04:05")
	case 2:
		return tm.Format("2006/01/02")
	case 3:
		return tm.Format("15:04:05.000000000")
	case 4:
		return tm.Format("15:04:05")
	case 5:
		return tm.Format("15:04")
	default:
		return tm.Format("2006/01/02 15:04:05.000000000")
	}
}

// Str2Time converts string to time
func Str2Time(str string) Gtime {
	var (
		ep   [6]float64
		year int
		mon  int
		day  int
		hour int
		min  int
		sec  float64
	)

	// Parse string
	fmt.Sscanf(str, "%d/%d/%d %d:%d:%f", &year, &mon, &day, &hour, &min, &sec)

	// Convert to epoch
	ep[0] = float64(year)
	ep[1] = float64(mon)
	ep[2] = float64(day)
	ep[3] = float64(hour)
	ep[4] = float64(min)
	ep[5] = sec

	// Convert to Gtime
	return Epoch2Time(ep)
}

// TimeDiff returns time difference in seconds
func TimeDiff(t1, t2 Gtime) float64 {
	return float64(t1.Time-t2.Time) + (t1.Sec - t2.Sec)
}

// TimeAdd adds time offset to time
func TimeAdd(t Gtime, sec float64) Gtime {
	var tt Gtime

	tt.Time = t.Time
	tt.Sec = t.Sec + sec

	if tt.Sec >= 1.0 {
		tt.Time += int64(tt.Sec)
		tt.Sec -= float64(int64(tt.Sec))
	} else if tt.Sec < 0.0 {
		tt.Time += int64(tt.Sec) - 1
		tt.Sec = 1.0 + tt.Sec - float64(int64(tt.Sec))
	}

	return tt
}
