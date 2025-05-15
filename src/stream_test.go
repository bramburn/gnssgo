package gnssgo

import (
	"testing"
)

// TestStreamInitialization tests the initialization of a Stream
func TestStreamInitialization(t *testing.T) {
	var stream Stream

	// Initialize the stream
	stream.InitStream()

	// Check that the stream was initialized correctly
	if stream.Type != 0 {
		t.Errorf("Stream type should be 0, got %d", stream.Type)
	}
	if stream.Mode != 0 {
		t.Errorf("Stream mode should be 0, got %d", stream.Mode)
	}
	if stream.State != 0 {
		t.Errorf("Stream state should be 0, got %d", stream.State)
	}
	if stream.InBytes != 0 {
		t.Errorf("Stream InBytes should be 0, got %d", stream.InBytes)
	}
	if stream.OutBytes != 0 {
		t.Errorf("Stream OutBytes should be 0, got %d", stream.OutBytes)
	}
	if stream.Port != nil {
		t.Errorf("Stream Port should be nil")
	}
	if stream.Path != "" {
		t.Errorf("Stream Path should be empty, got %s", stream.Path)
	}
}

// TestStreamOpenClose tests opening and closing different types of streams
func TestStreamOpenClose(t *testing.T) {
	var stream Stream

	// Test cases for different stream types
	testCases := []struct {
		name       string
		streamType int
		path       string
	}{
		{"File", STR_FILE, "test.txt"},
		{"Serial", STR_SERIAL, "COM1:9600:8:N:1:off"},
		// Skip TCP Server and Client tests as they require network setup
		// {"TCP Server", STR_TCPSVR, ":12345"},
		// {"TCP Client", STR_TCPCLI, "localhost:12345"},
		{"Memory Buffer", STR_MEMBUF, ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Initialize the stream
			stream.InitStream()

			// Open the stream
			result := stream.OpenStream(tc.streamType, STR_MODE_RW, tc.path)

			// We're just testing that the function doesn't crash
			// Most of these will fail without actual hardware/setup
			t.Logf("%s stream open result: %d, state: %d", tc.name, result, stream.State)

			// Close the stream
			stream.StreamClose()

			// Check that the stream was closed
			if stream.State != 0 && stream.Port != nil {
				t.Errorf("%s stream should be closed, state: %d, port: %v", tc.name, stream.State, stream.Port)
			}
		})
	}
}

// TestStreamReadWrite tests reading and writing to a memory buffer stream
// This is one of the few stream types we can test without actual hardware
func TestStreamReadWrite(t *testing.T) {
	var stream Stream

	// Initialize the stream
	stream.InitStream()

	// Open a memory buffer stream
	result := stream.OpenStream(STR_MEMBUF, STR_MODE_RW, "")

	if result <= 0 || stream.State <= 0 {
		t.Fatalf("Could not open memory buffer stream, result: %d, state: %d", result, stream.State)
	}

	// Write some data to the stream
	testData := []byte("Test data for memory buffer stream")
	ns := stream.StreamWrite(testData, len(testData))

	if ns != len(testData) {
		t.Errorf("Expected to write %d bytes, wrote %d", len(testData), ns)
	}

	// Read the data back
	readBuff := make([]byte, 1024)
	nr := stream.StreamRead(readBuff, 1024)

	if nr != len(testData) {
		t.Errorf("Expected to read %d bytes, read %d", len(testData), nr)
	}

	// Compare the data
	for i := 0; i < nr; i++ {
		if readBuff[i] != testData[i] {
			t.Errorf("Data mismatch at position %d: expected %d, got %d", i, testData[i], readBuff[i])
		}
	}

	// Close the stream
	stream.StreamClose()
}

// TestStreamStatus tests getting the status of a stream
func TestStreamStatus(t *testing.T) {
	var stream Stream

	// Initialize the stream
	stream.InitStream()

	// Open a memory buffer stream
	result := stream.OpenStream(STR_MEMBUF, STR_MODE_RW, "")

	if result <= 0 || stream.State <= 0 {
		t.Fatalf("Could not open memory buffer stream, result: %d, state: %d", result, stream.State)
	}

	// Get the stream status
	var msg string
	state := stream.StreamStat(&msg)

	// Memory buffer streams should be in state 1 (open) or 2 (active)
	if state < 1 {
		t.Errorf("Expected stream state >= 1, got %d", state)
	}

	// Get the extended stream status
	var extMsg string
	extState := stream.StreamStateX(&extMsg)

	// Extended state should match regular state
	if extState != state {
		t.Errorf("Extended state should match state, got %d and %d", state, extState)
	}

	// Close the stream
	stream.StreamClose()

	// Check the state after closing
	state = stream.StreamStat(&msg)
	if state != 0 {
		t.Errorf("Stream state after closing should be 0, got %d", state)
	}
}
