package caster

import (
	"fmt"
	"strings"
)

// Sourcetable for NTRIP Casters, returned at / as a way for users to discover available mounts
type Sourcetable struct {
	Casters  []CasterEntry
	Networks []NetworkEntry
	Mounts   []StreamEntry
}

// String returns the sourcetable as a string
func (st Sourcetable) String() string {
	stLength := (len(st.Casters) + len(st.Networks) + len(st.Mounts) + 1)
	stStrs := make([]string, 0, stLength)

	for _, cas := range st.Casters {
		stStrs = append(stStrs, cas.String())
	}

	for _, net := range st.Networks {
		stStrs = append(stStrs, net.String())
	}

	for _, str := range st.Mounts {
		stStrs = append(stStrs, str.String())
	}

	stStrs = append(stStrs, "ENDSOURCETABLE\r\n")
	return strings.Join(stStrs, "\r\n")
}

// CasterEntry for an NTRIP Sourcetable
type CasterEntry struct {
	Host                string
	Port                int
	Identifier          string
	Operator            string
	NMEA                bool
	Country             string
	Latitude            float32
	Longitude           float32
	FallbackHostAddress string
	FallbackHostPort    int
	Misc                string
}

// String returns the caster entry as a string
func (c CasterEntry) String() string {
	nmea := "0"
	if c.NMEA {
		nmea = "1"
	}

	return strings.Join([]string{"CAS",
		c.Host, fmt.Sprintf("%d", c.Port), c.Identifier, c.Operator, nmea, c.Country,
		fmt.Sprintf("%.4f", c.Latitude), fmt.Sprintf("%.4f", c.Longitude),
		c.FallbackHostAddress, fmt.Sprintf("%d", c.FallbackHostPort), c.Misc}, ";")
}

// NetworkEntry for an NTRIP Sourcetable
type NetworkEntry struct {
	Identifier          string
	Operator            string
	Authentication      string
	Fee                 bool
	NetworkInfoURL      string
	StreamInfoURL       string
	RegistrationAddress string
	Misc                string
}

// String returns the network entry as a string
func (n NetworkEntry) String() string {
	fee := "N"
	if n.Fee {
		fee = "Y"
	}

	return strings.Join([]string{"NET",
		n.Identifier, n.Operator, n.Authentication, fee, n.NetworkInfoURL, n.StreamInfoURL,
		n.RegistrationAddress, n.Misc}, ";")
}

// StreamEntry for an NTRIP Sourcetable
type StreamEntry struct {
	Name          string
	Identifier    string
	Format        string
	FormatDetails string
	Carrier       string
	NavSystem     string
	Network       string
	CountryCode   string
	Latitude      float32
	Longitude     float32
	NMEA          bool
	Solution      bool
	Generator     string
	Compression   string
	Authentication string
	Fee            bool
	Bitrate        int
	Misc           string
}

// String returns the stream entry as a string
func (m StreamEntry) String() string {
	nmea := "0"
	if m.NMEA {
		nmea = "1"
	}

	solution := "0"
	if m.Solution {
		solution = "1"
	}

	fee := "N"
	if m.Fee {
		fee = "Y"
	}

	return strings.Join([]string{"STR",
		m.Name, m.Identifier, m.Format, m.FormatDetails, m.Carrier, m.NavSystem, m.Network,
		m.CountryCode, fmt.Sprintf("%.4f", m.Latitude), fmt.Sprintf("%.4f", m.Longitude),
		nmea, solution, m.Generator, m.Compression, m.Authentication, fee,
		fmt.Sprintf("%d", m.Bitrate), m.Misc}, ";")
}
