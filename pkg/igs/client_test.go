package igs

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGPSWeekAndDay(t *testing.T) {
	tests := []struct {
		name     string
		time     time.Time
		wantWeek int
		wantDay  int
	}{
		{
			name:     "GPS start time",
			time:     time.Date(1980, 1, 6, 0, 0, 0, 0, time.UTC),
			wantWeek: 0,
			wantDay:  0,
		},
		{
			name:     "One week after GPS start",
			time:     time.Date(1980, 1, 13, 0, 0, 0, 0, time.UTC),
			wantWeek: 1,
			wantDay:  0,
		},
		{
			name:     "Middle of week",
			time:     time.Date(1980, 1, 9, 0, 0, 0, 0, time.UTC),
			wantWeek: 0,
			wantDay:  3,
		},
		{
			name:     "Recent date",
			time:     time.Date(2023, 5, 15, 0, 0, 0, 0, time.UTC),
			wantWeek: 2262,
			wantDay:  1, // Monday
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotWeek, gotDay := GPSWeekAndDay(tt.time)
			assert.Equal(t, tt.wantWeek, gotWeek)
			assert.Equal(t, tt.wantDay, gotDay)
		})
	}
}

func TestGetProductURL(t *testing.T) {
	client := NewClient("")
	client.SetBaseURL("https://igs.ign.fr/pub/igs/products")

	tests := []struct {
		name        string
		time        time.Time
		productType ProductType
		ac          AnalysisCenter
		want        string
	}{
		{
			name:        "IGS SP3 for 2023-05-15",
			time:        time.Date(2023, 5, 15, 0, 0, 0, 0, time.UTC),
			productType: ProductTypeSP3,
			ac:          AnalysisCenterIGS,
			want:        "https://igs.ign.fr/pub/igs/products/2262/igs22621.sp3.Z",
		},
		{
			name:        "IGS CLK for 2023-05-15",
			time:        time.Date(2023, 5, 15, 0, 0, 0, 0, time.UTC),
			productType: ProductTypeCLK,
			ac:          AnalysisCenterIGS,
			want:        "https://igs.ign.fr/pub/igs/products/2262/igs22621.clk.Z",
		},
		{
			name:        "JPL SP3 for 2023-05-15",
			time:        time.Date(2023, 5, 15, 0, 0, 0, 0, time.UTC),
			productType: ProductTypeSP3,
			ac:          AnalysisCenterJPL,
			want:        "https://igs.ign.fr/pub/igs/products/2262/jpl22621.sp3.Z",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := client.GetProductURL(tt.time, tt.productType, tt.ac)
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestDownloadFile(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test file content"))
	}))
	defer server.Close()

	// Create a temporary directory for downloads
	tempDir, err := os.MkdirTemp("", "igs-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a client
	client := NewClient(tempDir)
	client.HTTPClient = server.Client()

	// Test downloading a file
	localPath := filepath.Join(tempDir, "test.txt")
	err = client.DownloadFile(server.URL, localPath)
	assert.NoError(t, err)

	// Verify the file was downloaded
	content, err := os.ReadFile(localPath)
	assert.NoError(t, err)
	assert.Equal(t, "test file content", string(content))
}

func TestDownloadProduct_MockServer(t *testing.T) {
	// Create a test server that simulates an IGS data server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return 404 for all paths to simulate a failed download
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	// Create a temporary directory for downloads
	tempDir, err := os.MkdirTemp("", "igs-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a client with the test server URL
	client := NewClient(tempDir)
	client.SetBaseURL(server.URL)
	client.HTTPClient = server.Client()

	// Test downloading a product
	testTime := time.Date(2023, 5, 15, 0, 0, 0, 0, time.UTC)
	_, err = client.DownloadProduct(testTime, ProductTypeSP3, AnalysisCenterIGS)

	// This should fail because the test server returns 404
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "bad status")
}
