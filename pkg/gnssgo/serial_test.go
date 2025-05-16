package gnssgo

import (
	"testing"
)

// TestListSerialPorts tests the ListSerialPorts function
func TestListSerialPorts(t *testing.T) {
	// This test just verifies that the function doesn't crash
	ports, err := ListSerialPorts()

	// We don't assert on the number of ports since it depends on the system
	// Just check that the function runs without error
	if err != nil {
		t.Errorf("ListSerialPorts returned an error: %v", err)
	}

	t.Logf("Found %d serial ports", len(ports))
	for i, port := range ports {
		t.Logf("Port %d: %s", i, port)
	}
}

// TestOpenSerial tests the OpenSerial function
func TestOpenSerial(t *testing.T) {
	// Create a mock serial path
	path := "COM1:9600:8:N:1:off"
	var msg string

	// This test will not actually open a serial port, but it will test the parsing logic
	seri := OpenSerial(path, STR_MODE_RW, &msg)

	// On most systems without the actual hardware, this will return nil
	// We're just testing that the function doesn't crash
	if seri != nil {
		defer seri.CloseSerial()

		// Test the state function
		state := seri.StateSerial()
		if state < 0 {
			t.Errorf("Serial state should be >= 0, got %d", state)
		}

		// Test the extended state function
		var extMsg string
		extState := seri.StatExSerial(&extMsg)
		if state != extState {
			t.Errorf("Extended state should match state, got %d and %d", state, extState)
		}
		if extMsg == "" {
			t.Errorf("Extended state message should not be empty")
		}

		t.Logf("Serial extended state: %s", extMsg)
	} else {
		// If we couldn't open the port, check that there's an error message
		if msg == "" {
			t.Errorf("Error message should not be empty when port open fails")
		}
		t.Logf("Could not open serial port: %s", msg)
	}
}

// TestStreamWithSerial tests the Stream functionality with serial ports
func TestStreamWithSerial(t *testing.T) {
	// Create a stream
	var stream Stream

	// Initialize the stream
	stream.InitStream()

	// Test opening a serial stream (this will likely fail without actual hardware)
	path := "COM1:9600:8:N:1:off"
	result := stream.OpenStream(STR_SERIAL, STR_MODE_RW, path)

	// We're just testing that the function doesn't crash
	if result > 0 && stream.State > 0 {
		// If the stream opened successfully, test reading and writing
		buff := make([]byte, 1024)
		n := stream.StreamRead(buff, 1024)
		t.Logf("Read %d bytes from serial port", n)

		// Test writing (this will likely fail without actual hardware)
		testData := []byte("Test data")
		ns := stream.StreamWrite(testData, len(testData))
		t.Logf("Wrote %d bytes to serial port", ns)

		// Test getting stream status
		var msg string
		state := stream.StreamStat(&msg)
		t.Logf("Stream state: %d, message: %s", state, msg)

		// Test getting extended stream status
		var extMsg string
		extState := stream.StreamStateX(&extMsg)
		t.Logf("Stream extended state: %d, message: %s", extState, extMsg)
	} else {
		t.Logf("Could not open serial stream, state: %d", stream.State)
	}

	// Close the stream
	stream.StreamClose()
}

// TestSerialBaudRate tests changing the baud rate of a serial port
func TestSerialBaudRate(t *testing.T) {
	// Create a stream
	var stream Stream

	// Initialize the stream
	stream.InitStream()

	// Test opening a serial stream
	path := "COM1:9600:8:N:1:off"
	result := stream.OpenStream(STR_SERIAL, STR_MODE_RW, path)

	// We're just testing that the function doesn't crash
	if result > 0 && stream.State > 0 {
		// Test setting the baud rate
		// This is done by closing and reopening the stream with a new path
		stream.StreamClose()

		// Open with a new baud rate
		path = "COM1:115200:8:N:1:off"
		result = stream.OpenStream(STR_SERIAL, STR_MODE_RW, path)

		if result > 0 && stream.State > 0 {
			t.Logf("Successfully reopened stream with new baud rate")
		} else {
			t.Logf("Could not reopen stream with new baud rate, state: %d", stream.State)
		}
	} else {
		t.Logf("Could not open serial stream, state: %d", stream.State)
	}

	// Close the stream
	stream.StreamClose()
}
