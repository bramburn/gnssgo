package gnssgo

import (
	"math"
	"time"
)

// RTK status constants
const (
	RTK_STATUS_NONE   = 0 // No position
	RTK_STATUS_SINGLE = 1 // Single solution
	RTK_STATUS_FLOAT  = 2 // Float solution
	RTK_STATUS_FIX    = 4 // Fixed solution
)

// RTKStatus represents the current RTK status
type RTKStatus struct {
	Status    int       // RTK status (NONE, SINGLE, FLOAT, FIX)
	Time      time.Time // Time of the status
	Latitude  float64   // Latitude in degrees
	Longitude float64   // Longitude in degrees
	Altitude  float64   // Altitude in meters
	NSats     int       // Number of satellites
	Covariance [3]float64 // Position covariance (m²)
	HDOP      float64   // Horizontal dilution of precision
	VDOP      float64   // Vertical dilution of precision
	Age       float64   // Age of differential (seconds)
}

// NewRTKStatus creates a new RTK status
func NewRTKStatus() *RTKStatus {
	return &RTKStatus{
		Status: RTK_STATUS_NONE,
		Time:   time.Now(),
	}
}

// StatusString returns the RTK status as a string
func (s *RTKStatus) StatusString() string {
	switch s.Status {
	case RTK_STATUS_NONE:
		return "NONE"
	case RTK_STATUS_SINGLE:
		return "SINGLE"
	case RTK_STATUS_FLOAT:
		return "FLOAT"
	case RTK_STATUS_FIX:
		return "FIX"
	default:
		return "UNKNOWN"
	}
}

// HorizontalAccuracy returns the horizontal accuracy in meters
func (s *RTKStatus) HorizontalAccuracy() float64 {
	// Calculate horizontal accuracy from covariance
	if s.Covariance[0] <= 0 || s.Covariance[1] <= 0 {
		return 0.0
	}
	return math.Sqrt(s.Covariance[0] + s.Covariance[1])
}

// VerticalAccuracy returns the vertical accuracy in meters
func (s *RTKStatus) VerticalAccuracy() float64 {
	// Calculate vertical accuracy from covariance
	if s.Covariance[2] <= 0 {
		return 0.0
	}
	return math.Sqrt(s.Covariance[2])
}

// UpdateFromCovariance updates the RTK status based on covariance values
func (s *RTKStatus) UpdateFromCovariance() {
	// Calculate horizontal accuracy
	hAcc := s.HorizontalAccuracy()

	// Update status based on covariance values
	if hAcc < 0.1 {
		s.Status = RTK_STATUS_FIX
	} else if hAcc < 1.0 {
		s.Status = RTK_STATUS_FLOAT
	} else {
		s.Status = RTK_STATUS_SINGLE
	}
}

// UpdateFromNMEA updates the RTK status from a NMEA GGA sentence
func (s *RTKStatus) UpdateFromNMEA(quality int, lat, lon, alt float64, nsats int, hdop, age float64) {
	s.Latitude = lat
	s.Longitude = lon
	s.Altitude = alt
	s.NSats = nsats
	s.HDOP = hdop
	s.Age = age
	s.Time = time.Now()

	// Update status based on NMEA quality indicator
	switch quality {
	case 1:
		s.Status = RTK_STATUS_SINGLE
	case 2:
		s.Status = RTK_STATUS_FLOAT
	case 4:
		s.Status = RTK_STATUS_FIX
	default:
		s.Status = RTK_STATUS_NONE
	}
}
