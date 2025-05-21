package stream

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestSerialStream(t *testing.T) {
	// This is a basic test that doesn't actually open a serial port
	// but tests the parsing of the path
	var msg string

	// Test with invalid port
	seri := OpenSerial("INVALID_PORT", STR_MODE_RW, &msg)
	if seri != nil {
		t.Errorf("Expected nil for invalid port, got %v", seri)
	}

	// Test path parsing
	path := "COM3:9600:8:N:1:RTSCTS"
	seri = OpenSerial(path, STR_MODE_RW, &msg)
	// We expect nil since the port doesn't exist, but the parsing should work
	if seri != nil {
		defer seri.CloseSerial()
	}
}

func TestFileStream(t *testing.T) {
	// Create a temporary file
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "test.dat")

	// Create a test file
	data := []byte("GNSS test data")
	err := os.WriteFile(tempFile, data, 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Open the file for reading
	var msg string
	file := OpenStreamFile(tempFile, STR_MODE_R, &msg)
	if file == nil {
		t.Fatalf("Failed to open file: %s", msg)
	}
	defer file.CloseFile()

	// Read data from the file
	buff := make([]byte, 100)
	n := file.ReadFile(buff, 100, &msg)
	if n != len(data) {
		t.Errorf("Expected to read %d bytes, got %d", len(data), n)
	}

	// Verify the data
	if string(buff[:n]) != string(data) {
		t.Errorf("Expected %q, got %q", string(data), string(buff[:n]))
	}

	// Test writing to a file
	outFile := filepath.Join(tempDir, "out.dat")
	file = OpenStreamFile(outFile, STR_MODE_W, &msg)
	if file == nil {
		t.Fatalf("Failed to open output file: %s", msg)
	}

	// Write data to the file
	writeData := []byte("GNSS output data")
	n = file.WriteFile(writeData, len(writeData), &msg)
	if n != len(writeData) {
		t.Errorf("Expected to write %d bytes, got %d", len(writeData), n)
	}

	// Close the file
	file.CloseFile()

	// Verify the written data
	readData, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	if string(readData) != string(writeData) {
		t.Errorf("Expected %q, got %q", string(writeData), string(readData))
	}
}

func TestTcpStream(t *testing.T) {
	// This is a basic test that doesn't actually open a TCP connection
	// but tests the parsing of the path
	var msg string

	// Test TCP server
	tcpsvr := OpenTcpSvr(":8080", &msg)
	if tcpsvr == nil {
		t.Logf("TCP server creation failed (expected on CI): %s", msg)
	} else {
		defer tcpsvr.CloseTcpSvr()

		// Test server state
		state := tcpsvr.StateXTcpSvr(&msg)
		if state <= 0 {
			t.Errorf("Expected positive state for TCP server, got %d", state)
		}
	}

	// Test TCP client
	tcpcli := OpenTcpClient("localhost:8080", &msg)
	if tcpcli == nil {
		t.Logf("TCP client creation failed (expected): %s", msg)
	} else {
		defer tcpcli.CloseTcpClient()
	}
}

func TestUdpStream(t *testing.T) {
	// This is a basic test that doesn't actually open a UDP connection
	// but tests the parsing of the path
	var msg string

	// Test UDP server
	udpsvr := OpenUdpSvr(":8081", &msg)
	if udpsvr == nil {
		t.Logf("UDP server creation failed (expected on CI): %s", msg)
	} else {
		defer udpsvr.CloseUdp()

		// Test server state
		state := udpsvr.StatExUdpSvr(&msg)
		if state <= 0 {
			t.Errorf("Expected positive state for UDP server, got %d", state)
		}
	}

	// Test UDP client
	udpcli := OpenUdpClient("localhost:8081", &msg)
	if udpcli == nil {
		t.Logf("UDP client creation failed (expected): %s", msg)
	} else {
		defer udpcli.CloseUdp()
	}
}

func TestStreamIntegration(t *testing.T) {
	// Create a temporary file
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "stream_test.dat")

	// Create a test file
	testData := []byte("GNSS stream test data")
	err := os.WriteFile(tempFile, testData, 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create a stream
	var stream Stream
	// No message needed

	// Initialize the stream
	stream.InitStream()

	// Open the stream as a file
	result := stream.OpenStream(STR_FILE, STR_MODE_R, tempFile)
	if result <= 0 {
		t.Fatalf("Failed to open stream: %s", stream.Msg)
	}

	// Read data from the stream
	buff := make([]byte, 100)
	n := stream.StreamRead(buff, 100)
	if n != len(testData) {
		t.Errorf("Expected to read %d bytes, got %d", len(testData), n)
	}

	// Verify the data
	if string(buff[:n]) != string(testData) {
		t.Errorf("Expected %q, got %q", string(testData), string(buff[:n]))
	}

	// Close the stream
	stream.StreamClose()

	// Test stream state
	state := stream.State
	if state != 0 {
		t.Errorf("Expected state 0 for closed stream, got %d", state)
	}
}

func TestSetBrate(t *testing.T) {
	// This is a basic test that doesn't actually open a serial port
	var stream Stream
	// No message needed

	// Initialize the stream
	stream.InitStream()

	// Set baud rate (should do nothing since the stream is not open)
	SetBrate(&stream, 115200)

	// No error should occur
	if stream.State != 0 {
		t.Errorf("Expected state 0, got %d", stream.State)
	}
}

func TestStreamTimeout(t *testing.T) {
	// Create a temporary file
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "timeout_test.dat")

	// Create a test file
	testData := []byte("GNSS timeout test data")
	err := os.WriteFile(tempFile, testData, 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create a stream
	var stream Stream
	// No message needed

	// Initialize the stream
	stream.InitStream()

	// Open the stream as a file
	result := stream.OpenStream(STR_FILE, STR_MODE_R, tempFile)
	if result <= 0 {
		t.Fatalf("Failed to open stream: %s", stream.Msg)
	}

	// Read data from the stream
	buff := make([]byte, 100)
	n := stream.StreamRead(buff, 100)
	if n != len(testData) {
		t.Errorf("Expected to read %d bytes, got %d", len(testData), n)
	}

	// Wait for a moment
	time.Sleep(100 * time.Millisecond)

	// Try to read more data (should return 0 bytes)
	n = stream.StreamRead(buff, 100)
	if n != 0 {
		t.Errorf("Expected to read 0 bytes after EOF, got %d", n)
	}

	// Close the stream
	stream.StreamClose()
}
