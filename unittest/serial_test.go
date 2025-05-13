package gnss_test

import (
	"testing"

	"github.com/bramburn/gnssgo"
	"github.com/stretchr/testify/assert"
)

func TestListSerialPorts(t *testing.T) {
	// This test just verifies that the function doesn't crash
	ports, err := gnssgo.ListSerialPorts()
	
	// We don't assert on the number of ports since it depends on the system
	// Just check that the function runs without error
	assert.NoError(t, err, "ListSerialPorts should not return an error")
	t.Logf("Found %d serial ports", len(ports))
	for i, port := range ports {
		t.Logf("Port %d: %s", i, port)
	}
}

func TestSerialCommStruct(t *testing.T) {
	// Create a mock serial path
	path := "COM1:9600:8:N:1:off"
	var msg string
	
	// This test will not actually open a serial port, but it will test the parsing logic
	seri := gnssgo.OpenSerial(path, gnssgo.STR_MODE_RW, &msg)
	
	// On most systems without the actual hardware, this will return nil
	// We're just testing that the function doesn't crash
	if seri != nil {
		defer seri.CloseSerial()
		
		// Test the state function
		state := seri.StateSerial()
		assert.GreaterOrEqual(t, state, 0, "Serial state should be >= 0")
		
		// Test the extended state function
		var extMsg string
		extState := seri.StatExSerial(&extMsg)
		assert.Equal(t, state, extState, "Extended state should match state")
		assert.NotEmpty(t, extMsg, "Extended state message should not be empty")
		
		t.Logf("Serial extended state: %s", extMsg)
	} else {
		// If we couldn't open the port, check that there's an error message
		assert.NotEmpty(t, msg, "Error message should not be empty when port open fails")
		t.Logf("Could not open serial port: %s", msg)
	}
}

func TestStreamWithSerial(t *testing.T) {
	// Create a stream
	var stream gnssgo.Stream
	
	// Initialize the stream
	stream.InitStream()
	
	// Test opening a serial stream (this will likely fail without actual hardware)
	path := "COM1:9600:8:N:1:off"
	result := stream.OpenStream(gnssgo.STR_SERIAL, gnssgo.STR_MODE_RW, path)
	
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
