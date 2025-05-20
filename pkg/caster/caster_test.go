package caster

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestCasterSourcetable(t *testing.T) {
	// Create a new source service
	svc := NewInMemorySourceService()
	svc.Sourcetable = Sourcetable{
		Casters: []CasterEntry{
			{
				Host:       "localhost",
				Port:       2101,
				Identifier: "Test Caster",
				Operator:   "Test",
				NMEA:       true,
				Country:    "USA",
				Latitude:   37.7749,
				Longitude:  -122.4194,
			},
		},
		Networks: []NetworkEntry{
			{
				Identifier:          "TEST",
				Operator:            "Test",
				Authentication:      "B",
				Fee:                 false,
				NetworkInfoURL:      "http://example.com",
				StreamInfoURL:       "http://example.com/streams",
				RegistrationAddress: "admin@example.com",
			},
		},
		Mounts: []StreamEntry{
			{
				Name:           "TEST",
				Identifier:     "TEST",
				Format:         "RTCM 3.3",
				FormatDetails:  "1004(1),1005/1006(5)",
				Carrier:        "2",
				NavSystem:      "GPS+GLO",
				Network:        "TEST",
				CountryCode:    "USA",
				Latitude:       37.7749,
				Longitude:      -122.4194,
				NMEA:           true,
				Solution:       false,
				Generator:      "Test",
				Compression:    "none",
				Authentication: "B",
				Fee:            false,
				Bitrate:        9600,
			},
		},
	}

	// Create a new caster
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	caster := NewCaster("N/A", svc, logger)

	// Create a test server
	ts := httptest.NewServer(caster.Handler)
	defer ts.Close()

	// Send a request to get the sourcetable
	resp, err := http.Get(ts.URL)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	resp.Body.Close()

	// Check that the response contains the sourcetable
	assert.Contains(t, string(body), "CAS;localhost;2101;Test Caster;Test;1;USA;37.7749;-122.4194")
	assert.Contains(t, string(body), "NET;TEST;Test;B;N;http://example.com;http://example.com/streams;admin@example.com")
	assert.Contains(t, string(body), "STR;TEST;TEST;RTCM 3.3;1004(1),1005/1006(5);2;GPS+GLO;TEST;USA;37.7749;-122.4194;1;0;Test;none;B;N;9600")
	assert.Contains(t, string(body), "ENDSOURCETABLE")
}

func TestCasterSourcetableOnly(t *testing.T) {
	// Create a new source service
	svc := NewInMemorySourceService()
	svc.Sourcetable = Sourcetable{
		Mounts: []StreamEntry{
			{
				Name:       "TEST",
				Identifier: "TEST",
				Format:     "RTCM 3.3",
			},
		},
	}

	// Create a new caster
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	caster := NewCaster("N/A", svc, logger)

	// Create a test server
	ts := httptest.NewServer(caster.Handler)
	defer ts.Close()

	// Send a request to get the sourcetable
	resp, err := http.Get(ts.URL + "/")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)

	// Check that the response contains the mount
	assert.Contains(t, string(body), "STR;TEST;TEST;RTCM 3.3")
}

func TestCasterNotFound(t *testing.T) {
	// Create a new source service
	svc := NewInMemorySourceService()
	svc.Sourcetable = Sourcetable{
		Mounts: []StreamEntry{
			{
				Name:       "TEST",
				Identifier: "TEST",
				Format:     "RTCM 3.3",
			},
		},
	}

	// Create a new caster
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	caster := NewCaster("N/A", svc, logger)

	// Create a test server
	ts := httptest.NewServer(caster.Handler)
	defer ts.Close()

	// Send a request to a non-existent mountpoint
	resp, err := http.Get(ts.URL + "/NONEXISTENT")
	assert.NoError(t, err)

	// For NTRIP v1, a 404 is returned as a 200 with the sourcetable
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	resp.Body.Close()

	// Check that the response is a sourcetable
	assert.Contains(t, string(body), "SOURCETABLE 200 OK")
	assert.Contains(t, string(body), "ENDSOURCETABLE")
}
