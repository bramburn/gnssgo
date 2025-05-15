/*------------------------------------------------------------------------------
* GNSS-GO unit test : IGS product downloader
*-----------------------------------------------------------------------------*/

package gnss_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/bramburn/gnssgo/pkg/igs"
	"github.com/stretchr/testify/assert"
)

func TestIGSClient_Integration(t *testing.T) {
	// Skip this test in CI environments or when running quick tests
	if os.Getenv("CI") != "" || testing.Short() {
		t.Skip("Skipping integration test in CI environment")
	}

	// Create a temporary directory for downloads
	tempDir, err := os.MkdirTemp("", "igs-integration-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a client
	client := igs.NewClient(tempDir)

	// Test GPS week and day calculation
	t.Run("GPSWeekAndDay", func(t *testing.T) {
		// Test with a known date
		testTime := time.Date(2023, 5, 15, 0, 0, 0, 0, time.UTC)
		week, day := igs.GPSWeekAndDay(testTime)

		// These values should be verified with an external source
		assert.Equal(t, 2262, week)
		assert.Equal(t, 1, day) // Monday
	})

	// Test URL generation
	t.Run("GetProductURL", func(t *testing.T) {
		testTime := time.Date(2023, 5, 15, 0, 0, 0, 0, time.UTC)

		// Test SP3 URL
		sp3URL, err := client.GetProductURL(testTime, igs.ProductTypeSP3, igs.AnalysisCenterIGS)
		assert.NoError(t, err)
		assert.Contains(t, sp3URL, "igs22621.sp3.Z")

		// Test CLK URL
		clkURL, err := client.GetProductURL(testTime, igs.ProductTypeCLK, igs.AnalysisCenterIGS)
		assert.NoError(t, err)
		assert.Contains(t, clkURL, "igs22621.clk.Z")
	})
}

func TestIGSClient_MockServer(t *testing.T) {
	// Create a test server that simulates an IGS data server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if the request path contains expected patterns
		path := r.URL.Path

		if path == "/2262/igs22621.sp3.Z" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("mock sp3 file content"))
			return
		}

		if path == "/2262/igs22621.clk.Z" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("mock clk file content"))
			return
		}

		// Return 404 for other paths
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	// Create a temporary directory for downloads
	tempDir, err := os.MkdirTemp("", "igs-mock-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a client with the test server URL
	client := igs.NewClient(tempDir)
	client.SetBaseURL(server.URL)
	client.HTTPClient = server.Client()

	// Test time
	testTime := time.Date(2023, 5, 15, 0, 0, 0, 0, time.UTC)

	// Test downloading SP3 file
	t.Run("DownloadSP3", func(t *testing.T) {
		filePath, err := client.DownloadSP3(testTime, igs.AnalysisCenterIGS)
		assert.NoError(t, err)

		// Verify the file was downloaded
		content, err := os.ReadFile(filePath)
		assert.NoError(t, err)
		assert.Equal(t, "mock sp3 file content", string(content))
	})

	// Test downloading CLK file
	t.Run("DownloadCLK", func(t *testing.T) {
		filePath, err := client.DownloadCLK(testTime, igs.AnalysisCenterIGS)
		assert.NoError(t, err)

		// Verify the file was downloaded
		content, err := os.ReadFile(filePath)
		assert.NoError(t, err)
		assert.Equal(t, "mock clk file content", string(content))
	})

	// Test error handling for non-existent product
	t.Run("DownloadNonExistentProduct", func(t *testing.T) {
		// Use a different analysis center that won't match our mock server paths
		_, err := client.DownloadSP3(testTime, igs.AnalysisCenterJPL)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "bad status")
	})
}
