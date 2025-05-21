package rtcm

import (
	"fmt"
	"math"

	"github.com/bramburn/gnssgo/pkg/gnssgo"
)

// GPSEphemeris represents GPS ephemeris data from RTCM message 1019
type GPSEphemeris struct {
	SatID        uint8   // Satellite ID
	Week         uint16  // GPS week number
	SvAccuracy   uint8   // SV accuracy (URA index)
	CodeOnL2     uint8   // Code on L2
	IDOT         float64 // Rate of inclination angle (rad/s)
	IODE         uint8   // Issue of data, ephemeris
	Toc          uint32  // Clock data reference time (s)
	Af2          float64 // Clock correction polynomial coefficient (s/s²)
	Af1          float64 // Clock correction polynomial coefficient (s/s)
	Af0          float64 // Clock correction polynomial coefficient (s)
	IODC         uint16  // Issue of data, clock
	Crs          float64 // Amplitude of sine harmonic correction term to orbit radius (m)
	DeltaN       float64 // Mean motion difference from computed value (rad/s)
	M0           float64 // Mean anomaly at reference time (rad)
	Cuc          float64 // Amplitude of cosine harmonic correction term to argument of latitude (rad)
	Eccentricity float64 // Eccentricity
	Cus          float64 // Amplitude of sine harmonic correction term to argument of latitude (rad)
	SqrtA        float64 // Square root of semi-major axis (m^(1/2))
	Toe          uint32  // Ephemeris reference time (s)
	Cic          float64 // Amplitude of cosine harmonic correction term to inclination angle (rad)
	Omega0       float64 // Longitude of ascending node of orbit plane at weekly epoch (rad)
	Cis          float64 // Amplitude of sine harmonic correction term to inclination angle (rad)
	Inclination  float64 // Inclination angle at reference time (rad)
	Crc          float64 // Amplitude of cosine harmonic correction term to orbit radius (m)
	Omega        float64 // Argument of perigee (rad)
	OmegaDot     float64 // Rate of right ascension (rad/s)
	TGD          float64 // Group delay differential (s)
	SvHealth     uint8   // SV health
	L2PDataFlag  bool    // L2 P data flag
	FitInterval  bool    // Fit interval flag
}

// GLONASSEphemeris represents GLONASS ephemeris data from RTCM message 1020
type GLONASSEphemeris struct {
	SatID         uint8   // Satellite ID
	FreqNum       int8    // Frequency number (-7..+6)
	DayNumber     uint8   // Day number within 4-year period
	Tb            uint32  // Time of ephemeris (s)
	SvHealth      bool    // SV health
	P1            bool    // P1 flag
	P2            bool    // P2 flag
	P3            bool    // P3 flag
	P4            bool    // P4 flag
	X             float64 // X coordinate (km)
	Y             float64 // Y coordinate (km)
	Z             float64 // Z coordinate (km)
	VX            float64 // X velocity (km/s)
	VY            float64 // Y velocity (km/s)
	VZ            float64 // Z velocity (km/s)
	AX            float64 // X acceleration (km/s²)
	AY            float64 // Y acceleration (km/s²)
	AZ            float64 // Z acceleration (km/s²)
	GammaN        float64 // Relative frequency bias
	TauN          float64 // SV clock bias (s)
	DeltaTauN     float64 // Time difference between L1 and L2 (s)
	En            uint8   // Age of ephemeris data (days)
	P             bool    // P flag
	FT            uint8   // GLONASS-M flags
	NT            uint16  // Current date (days from 1990-01-01)
	N4            uint8   // 4-year interval number starting from 1996
	M             bool    // M flag
	AvailabilityA bool    // Availability of A parameters
	NA            uint16  // Calendar day number within 4-year period
	TauC          float64 // Time scale correction to UTC(SU) (s)
	N             uint16  // Calendar day number
	AvailabilityB bool    // Availability of B parameters
	TauGPS        float64 // Correction to GPS time (s)
}

// decodeGPSEphemeris decodes RTCM message 1019 (GPS Ephemeris)
func decodeGPSEphemeris(msg *RTCMMessage) (*GPSEphemeris, error) {
	if msg == nil || msg.Type != RTCM_GPS_EPHEMERIS {
		return nil, fmt.Errorf("not a GPS ephemeris message")
	}

	if len(msg.Data) < 15 {
		return nil, fmt.Errorf("message too short for GPS ephemeris")
	}

	// Start position after message type and station ID (24 + 12 = 36 bits)
	pos := 36

	// Create GPS ephemeris
	eph := &GPSEphemeris{}

	// Decode satellite ID
	eph.SatID = uint8(gnssgo.GetBitU(msg.Data, pos, 6))
	pos += 6

	// Decode week number
	eph.Week = uint16(gnssgo.GetBitU(msg.Data, pos, 10))
	pos += 10

	// Decode SV accuracy
	eph.SvAccuracy = uint8(gnssgo.GetBitU(msg.Data, pos, 4))
	pos += 4

	// Decode code on L2
	eph.CodeOnL2 = uint8(gnssgo.GetBitU(msg.Data, pos, 2))
	pos += 2

	// Decode IDOT
	eph.IDOT = float64(gnssgo.GetBits(msg.Data, pos, 14)) * math.Pow(2, -43) * math.Pi
	pos += 14

	// Decode IODE
	eph.IODE = uint8(gnssgo.GetBitU(msg.Data, pos, 8))
	pos += 8

	// Decode Toc
	eph.Toc = uint32(gnssgo.GetBitU(msg.Data, pos, 16)) * 16
	pos += 16

	// Decode Af2
	eph.Af2 = float64(gnssgo.GetBits(msg.Data, pos, 8)) * math.Pow(2, -55)
	pos += 8

	// Decode Af1
	eph.Af1 = float64(gnssgo.GetBits(msg.Data, pos, 16)) * math.Pow(2, -43)
	pos += 16

	// Decode Af0
	eph.Af0 = float64(gnssgo.GetBits(msg.Data, pos, 22)) * math.Pow(2, -31)
	pos += 22

	// Decode IODC
	eph.IODC = uint16(gnssgo.GetBitU(msg.Data, pos, 10))
	pos += 10

	// Decode Crs
	eph.Crs = float64(gnssgo.GetBits(msg.Data, pos, 16)) * math.Pow(2, -5)
	pos += 16

	// Decode DeltaN
	eph.DeltaN = float64(gnssgo.GetBits(msg.Data, pos, 16)) * math.Pow(2, -43) * math.Pi
	pos += 16

	// Decode M0
	eph.M0 = float64(gnssgo.GetBits(msg.Data, pos, 32)) * math.Pow(2, -31) * math.Pi
	pos += 32

	// Decode Cuc
	eph.Cuc = float64(gnssgo.GetBits(msg.Data, pos, 16)) * math.Pow(2, -29)
	pos += 16

	// Decode Eccentricity
	eph.Eccentricity = float64(gnssgo.GetBitU(msg.Data, pos, 32)) * math.Pow(2, -33)
	pos += 32

	// Decode Cus
	eph.Cus = float64(gnssgo.GetBits(msg.Data, pos, 16)) * math.Pow(2, -29)
	pos += 16

	// Decode SqrtA
	eph.SqrtA = float64(gnssgo.GetBitU(msg.Data, pos, 32)) * math.Pow(2, -19)
	pos += 32

	// Decode Toe
	eph.Toe = uint32(gnssgo.GetBitU(msg.Data, pos, 16)) * 16
	pos += 16

	// Decode Cic
	eph.Cic = float64(gnssgo.GetBits(msg.Data, pos, 16)) * math.Pow(2, -29)
	pos += 16

	// Decode Omega0
	eph.Omega0 = float64(gnssgo.GetBits(msg.Data, pos, 32)) * math.Pow(2, -31) * math.Pi
	pos += 32

	// Decode Cis
	eph.Cis = float64(gnssgo.GetBits(msg.Data, pos, 16)) * math.Pow(2, -29)
	pos += 16

	// Decode Inclination
	eph.Inclination = float64(gnssgo.GetBits(msg.Data, pos, 32)) * math.Pow(2, -31) * math.Pi
	pos += 32

	// Decode Crc
	eph.Crc = float64(gnssgo.GetBits(msg.Data, pos, 16)) * math.Pow(2, -5)
	pos += 16

	// Decode Omega
	eph.Omega = float64(gnssgo.GetBits(msg.Data, pos, 32)) * math.Pow(2, -31) * math.Pi
	pos += 32

	// Decode OmegaDot
	eph.OmegaDot = float64(gnssgo.GetBits(msg.Data, pos, 24)) * math.Pow(2, -43) * math.Pi
	pos += 24

	// Decode TGD
	eph.TGD = float64(gnssgo.GetBits(msg.Data, pos, 8)) * math.Pow(2, -31)
	pos += 8

	// Decode SvHealth
	eph.SvHealth = uint8(gnssgo.GetBitU(msg.Data, pos, 6))
	pos += 6

	// Decode L2PDataFlag
	eph.L2PDataFlag = gnssgo.GetBitU(msg.Data, pos, 1) != 0
	pos += 1

	// Decode FitInterval
	eph.FitInterval = gnssgo.GetBitU(msg.Data, pos, 1) != 0

	return eph, nil
}

// decodeGLONASSEphemeris decodes RTCM message 1020 (GLONASS Ephemeris)
func decodeGLONASSEphemeris(msg *RTCMMessage) (*GLONASSEphemeris, error) {
	if msg == nil || msg.Type != RTCM_GLONASS_EPHEMERIS {
		return nil, fmt.Errorf("not a GLONASS ephemeris message")
	}

	if len(msg.Data) < 15 {
		return nil, fmt.Errorf("message too short for GLONASS ephemeris")
	}

	// Start position after message type and station ID (24 + 12 = 36 bits)
	pos := 36

	// Create GLONASS ephemeris
	eph := &GLONASSEphemeris{}

	// Decode satellite ID
	eph.SatID = uint8(gnssgo.GetBitU(msg.Data, pos, 6))
	pos += 6

	// Decode frequency number
	eph.FreqNum = int8(gnssgo.GetBits(msg.Data, pos, 5))
	pos += 5

	// Decode day number
	eph.DayNumber = uint8(gnssgo.GetBitU(msg.Data, pos, 5))
	pos += 5

	// Decode Tb
	eph.Tb = uint32(gnssgo.GetBitU(msg.Data, pos, 7)) * 15 * 60
	pos += 7

	// Decode SV health
	eph.SvHealth = gnssgo.GetBitU(msg.Data, pos, 1) == 0 // 0 = healthy
	pos += 1

	// Decode P1 flag
	eph.P1 = gnssgo.GetBitU(msg.Data, pos, 1) != 0
	pos += 1

	// Decode P2 flag
	eph.P2 = gnssgo.GetBitU(msg.Data, pos, 1) != 0
	pos += 1

	// Decode P3 flag
	eph.P3 = gnssgo.GetBitU(msg.Data, pos, 1) != 0
	pos += 1

	// Decode P4 flag
	eph.P4 = gnssgo.GetBitU(msg.Data, pos, 1) != 0
	pos += 1

	// Decode X coordinate
	eph.X = float64(gnssgo.GetBits(msg.Data, pos, 27)) * math.Pow(2, -11)
	pos += 27

	// Decode Y coordinate
	eph.Y = float64(gnssgo.GetBits(msg.Data, pos, 27)) * math.Pow(2, -11)
	pos += 27

	// Decode Z coordinate
	eph.Z = float64(gnssgo.GetBits(msg.Data, pos, 27)) * math.Pow(2, -11)
	pos += 27

	// Decode X velocity
	eph.VX = float64(gnssgo.GetBits(msg.Data, pos, 24)) * math.Pow(2, -20)
	pos += 24

	// Decode Y velocity
	eph.VY = float64(gnssgo.GetBits(msg.Data, pos, 24)) * math.Pow(2, -20)
	pos += 24

	// Decode Z velocity
	eph.VZ = float64(gnssgo.GetBits(msg.Data, pos, 24)) * math.Pow(2, -20)
	pos += 24

	// Decode X acceleration
	eph.AX = float64(gnssgo.GetBits(msg.Data, pos, 5)) * math.Pow(2, -30)
	pos += 5

	// Decode Y acceleration
	eph.AY = float64(gnssgo.GetBits(msg.Data, pos, 5)) * math.Pow(2, -30)
	pos += 5

	// Decode Z acceleration
	eph.AZ = float64(gnssgo.GetBits(msg.Data, pos, 5)) * math.Pow(2, -30)
	pos += 5

	// Decode GammaN (relative frequency bias)
	eph.GammaN = float64(gnssgo.GetBits(msg.Data, pos, 11)) * math.Pow(2, -40)
	pos += 11

	// Decode TauN (SV clock bias)
	eph.TauN = float64(gnssgo.GetBits(msg.Data, pos, 22)) * math.Pow(2, -30)
	pos += 22

	// Decode DeltaTauN (time difference between L1 and L2)
	eph.DeltaTauN = float64(gnssgo.GetBits(msg.Data, pos, 5)) * math.Pow(2, -30)
	pos += 5

	// Decode En (age of ephemeris data)
	eph.En = uint8(gnssgo.GetBitU(msg.Data, pos, 5))
	pos += 5

	// Decode P flag
	eph.P = gnssgo.GetBitU(msg.Data, pos, 1) != 0
	pos += 1

	// Decode FT (GLONASS-M flags)
	eph.FT = uint8(gnssgo.GetBitU(msg.Data, pos, 4))
	pos += 4

	// Decode NT (current date)
	eph.NT = uint16(gnssgo.GetBitU(msg.Data, pos, 11))
	pos += 11

	// Decode N4 (4-year interval number)
	eph.N4 = uint8(gnssgo.GetBitU(msg.Data, pos, 5))
	pos += 5

	// Decode M flag
	eph.M = gnssgo.GetBitU(msg.Data, pos, 1) != 0
	pos += 1

	// Decode availability of additional data
	eph.AvailabilityA = gnssgo.GetBitU(msg.Data, pos, 1) != 0
	pos += 1

	// Decode NA (calendar day number within 4-year period)
	if eph.AvailabilityA {
		eph.NA = uint16(gnssgo.GetBitU(msg.Data, pos, 11))
		pos += 11

		// Decode TauC (time scale correction to UTC(SU))
		eph.TauC = float64(gnssgo.GetBits(msg.Data, pos, 32)) * math.Pow(2, -31)
		pos += 32

		// Decode N (calendar day number)
		eph.N = uint16(gnssgo.GetBitU(msg.Data, pos, 5))
		pos += 5

		// Decode availability of GPS time parameters
		eph.AvailabilityB = gnssgo.GetBitU(msg.Data, pos, 1) != 0
		pos += 1

		// Decode TauGPS (correction to GPS time)
		if eph.AvailabilityB {
			eph.TauGPS = float64(gnssgo.GetBits(msg.Data, pos, 22)) * math.Pow(2, -31)
		}
	}

	return eph, nil
}

// decodeSSROrbitClock decodes RTCM messages 1057-1062 (SSR Orbit and Clock Corrections)
func decodeSSROrbitClock(msg *RTCMMessage) (interface{}, error) {
	// Use the implementation from ssr.go
	return decodeSSROrbitClockCorrection(msg)
}
