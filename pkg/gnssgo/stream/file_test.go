package stream

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/bramburn/gnssgo/pkg/gnssgo/gtime"
)

func TestReppath(t *testing.T) {
	// Create a test time
	testTime := gtime.Gtime{
		Time: 1609459200, // 2021-01-01 00:00:00 UTC
		Sec:  0.0,
	}

	tests := []struct {
		name     string
		path     string
		time     gtime.Gtime
		sta      string
		ext      string
		expected string
	}{
		{
			name:     "No keywords",
			path:     "/data/file.dat",
			time:     testTime,
			sta:      "STAT",
			ext:      "ext",
			expected: "/data/file.dat",
		},
		{
			name:     "Year keyword",
			path:     "/data/%Y/file.dat",
			time:     testTime,
			sta:      "STAT",
			ext:      "ext",
			expected: "/data/2021/file.dat",
		},
		{
			name:     "Month keyword",
			path:     "/data/%m/file.dat",
			time:     testTime,
			sta:      "STAT",
			ext:      "ext",
			expected: "/data/01/file.dat",
		},
		{
			name:     "Day keyword",
			path:     "/data/%d/file.dat",
			time:     testTime,
			sta:      "STAT",
			ext:      "ext",
			expected: "/data/01/file.dat",
		},
		{
			name:     "Station keyword",
			path:     "/data/%r/file.dat",
			time:     testTime,
			sta:      "STAT",
			ext:      "ext",
			expected: "/data/stat/file.dat",
		},
		{
			name:     "Multiple keywords",
			path:     "/data/%Y/%m/%d/%r.%e",
			time:     testTime,
			sta:      "STAT",
			ext:      "dat",
			expected: "/data/2021/01/01/stat.dat",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := reppath(tt.path, tt.time, tt.sta, tt.ext)
			if result != tt.expected {
				t.Errorf("reppath() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestFileSwapping(t *testing.T) {
	// Create temporary directory for test files
	tempDir, err := os.MkdirTemp("", "fileswap_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test file path with time pattern
	testPath := filepath.Join(tempDir, "%Y%m%d.dat")
	
	// Create a file stream with swap interval of 1 hour
	var msg string
	file := OpenStreamFile(testPath+"::S=1", STR_MODE_W, &msg)
	if file == nil {
		t.Fatalf("Failed to open file stream: %s", msg)
	}
	defer file.CloseFile()

	// Write some data
	data := []byte("test data")
	if n := file.WriteFile(data, len(data), &msg); n <= 0 {
		t.Fatalf("Failed to write to file: %s", msg)
	}

	// Get the current file path
	initialPath := file.openpath

	// Simulate time passing (more than swap interval)
	// This is a bit hacky but necessary for testing
	file.wtime = gtime.TimeAdd(file.wtime, -3600.1) // Subtract more than 1 hour

	// Write again to trigger swap
	if n := file.WriteFile(data, len(data), &msg); n <= 0 {
		t.Fatalf("Failed to write to file after time change: %s", msg)
	}

	// Check if file was swapped
	if file.openpath == initialPath {
		t.Errorf("File was not swapped after interval passed")
	}
}

func TestCompressedFileHandling(t *testing.T) {
	// Skip this test if not running in a full test environment
	if testing.Short() {
		t.Skip("Skipping compressed file test in short mode")
	}

	// Create temporary directory for test files
	tempDir, err := os.MkdirTemp("", "compress_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test gzip file
	gzipPath := filepath.Join(tempDir, "test.gz")
	var outPath string
	
	// Test the uncompress function
	result := uncompress(gzipPath, &outPath)
	
	// We expect it to fail since we didn't actually create the file,
	// but this at least tests that the function runs
	if result != -1 && result != 0 {
		t.Errorf("Expected uncompress to return -1 or 0 for non-existent file, got %d", result)
	}
}
