package top708

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.bug.st/serial/enumerator"
)

// MockSerialPort is a mock implementation of the SerialPort interface
type MockSerialPort struct {
	mock.Mock
	connected bool
	data      []byte
	written   []byte
	timeout   time.Duration
}

func (p *MockSerialPort) Open(portName string, baudRate int) error {
	args := p.Called(portName, baudRate)
	p.connected = true
	return args.Error(0)
}

func (p *MockSerialPort) Close() error {
	args := p.Called()
	p.connected = false
	return args.Error(0)
}

func (p *MockSerialPort) Read(buffer []byte) (int, error) {
	args := p.Called(buffer)

	// If the mock is not connected, return an error
	if !p.connected {
		return 0, errors.New("port not open")
	}

	// If there's no data, return 0
	if len(p.data) == 0 {
		return 0, nil
	}

	// Return the values specified in the test setup
	return args.Int(0), args.Error(1)
}

func (p *MockSerialPort) Write(data []byte) (int, error) {
	args := p.Called(data)

	// If the mock is not connected, return an error
	if !p.connected {
		return 0, errors.New("port not open")
	}

	// Return the values specified in the test setup
	return args.Int(0), args.Error(1)
}

func (p *MockSerialPort) SetReadTimeout(timeout time.Duration) error {
	args := p.Called(timeout)
	p.timeout = timeout
	return args.Error(0)
}

func (p *MockSerialPort) GetTimeout() time.Duration {
	args := p.Called()
	if len(args) > 0 {
		return args.Get(0).(time.Duration)
	}
	return p.timeout
}

func (p *MockSerialPort) ListPorts() ([]string, error) {
	args := p.Called()
	return args.Get(0).([]string), args.Error(1)
}

func (p *MockSerialPort) GetPortDetails() ([]*enumerator.PortDetails, error) {
	args := p.Called()
	return args.Get(0).([]*enumerator.PortDetails), args.Error(1)
}

// TestNewTOP708Device tests the NewTOP708Device function
func TestNewTOP708Device(t *testing.T) {
	// Create a mock serial port
	serialPort := new(MockSerialPort)

	// Create a new TOP708 device
	device := NewTOP708Device(serialPort)

	// Verify the device was created correctly
	assert.NotNil(t, device)
	assert.Equal(t, serialPort, device.serialPort)
	assert.False(t, device.connected)
}

// TestTOP708DeviceConnect tests the Connect method
func TestTOP708DeviceConnect(t *testing.T) {
	// Create a mock serial port
	serialPort := new(MockSerialPort)
	serialPort.On("Open", "COM1", 38400).Return(nil)

	// Create a new TOP708 device
	device := NewTOP708Device(serialPort)

	// Connect to the device
	err := device.Connect("COM1", 38400)

	// Verify the connection was successful
	assert.NoError(t, err)
	assert.True(t, device.connected)
	serialPort.AssertCalled(t, "Open", "COM1", 38400)
}

// TestTOP708DeviceConnectError tests the Connect method with an error
func TestTOP708DeviceConnectError(t *testing.T) {
	// Create a mock serial port
	serialPort := new(MockSerialPort)
	serialPort.On("Open", "COM1", 38400).Return(errors.New("open error"))

	// Create a new TOP708 device
	device := NewTOP708Device(serialPort)

	// Connect to the device
	err := device.Connect("COM1", 38400)

	// Verify the connection failed
	assert.Error(t, err)
	assert.False(t, device.connected)
	assert.Contains(t, err.Error(), "failed to connect to device")
	serialPort.AssertCalled(t, "Open", "COM1", 38400)
}

// TestTOP708DeviceDisconnect tests the Disconnect method
func TestTOP708DeviceDisconnect(t *testing.T) {
	// Create a mock serial port
	serialPort := new(MockSerialPort)
	serialPort.On("Close").Return(nil)

	// Create a new TOP708 device
	device := NewTOP708Device(serialPort)
	device.connected = true

	// Disconnect from the device
	err := device.Disconnect()

	// Verify the disconnection was successful
	assert.NoError(t, err)
	assert.False(t, device.connected)
	serialPort.AssertCalled(t, "Close")
}

// TestTOP708DeviceDisconnectError tests the Disconnect method with an error
func TestTOP708DeviceDisconnectError(t *testing.T) {
	// Create a mock serial port
	serialPort := new(MockSerialPort)
	serialPort.On("Close").Return(errors.New("close error"))

	// Create a new TOP708 device
	device := NewTOP708Device(serialPort)
	device.connected = true

	// Disconnect from the device
	err := device.Disconnect()

	// Verify the disconnection failed
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error disconnecting device")
	serialPort.AssertCalled(t, "Close")
}

// TestTOP708DeviceIsConnected tests the IsConnected method
func TestTOP708DeviceIsConnected(t *testing.T) {
	// Create a mock serial port
	serialPort := new(MockSerialPort)

	// Create a new TOP708 device
	device := NewTOP708Device(serialPort)

	// Test when not connected
	assert.False(t, device.IsConnected())

	// Test when connected
	device.connected = true
	assert.True(t, device.IsConnected())
}

// TestTOP708DeviceVerifyConnection tests the VerifyConnection method
func TestTOP708DeviceVerifyConnection(t *testing.T) {
	// Create a mock serial port
	serialPort := new(MockSerialPort)
	serialPort.connected = true

	// Setup the mock to copy the data to the buffer and return the length
	serialPort.data = []byte("$GPGGA,123519,4807.038,N,01131.000,E,1,08,0.9,545.4,M,46.9,M,,*47\r\n")
	serialPort.On("Read", mock.Anything).Run(func(args mock.Arguments) {
		buffer := args.Get(0).([]byte)
		copy(buffer, serialPort.data)
	}).Return(len(serialPort.data), nil)

	// Create a new TOP708 device
	device := NewTOP708Device(serialPort)
	device.connected = true

	// Verify the connection
	result := device.VerifyConnection(100 * time.Millisecond)

	// Verify the result
	assert.True(t, result)
	serialPort.AssertCalled(t, "Read", mock.Anything)
}

// TestTOP708DeviceVerifyConnectionNotConnected tests the VerifyConnection method when not connected
func TestTOP708DeviceVerifyConnectionNotConnected(t *testing.T) {
	// Create a mock serial port
	serialPort := new(MockSerialPort)

	// Create a new TOP708 device
	device := NewTOP708Device(serialPort)

	// Verify the connection
	result := device.VerifyConnection(100 * time.Millisecond)

	// Verify the result
	assert.False(t, result)
}

// TestTOP708DeviceReadRaw tests the ReadRaw method
func TestTOP708DeviceReadRaw(t *testing.T) {
	// Create a mock serial port
	serialPort := new(MockSerialPort)
	serialPort.connected = true
	serialPort.data = []byte("test data")

	// Setup the mock to copy the data to the buffer and return the length
	serialPort.On("Read", mock.Anything).Run(func(args mock.Arguments) {
		buffer := args.Get(0).([]byte)
		copy(buffer, serialPort.data)
	}).Return(9, nil)

	// Create a new TOP708 device
	device := NewTOP708Device(serialPort)
	device.connected = true

	// Read data from the device
	buffer := make([]byte, 1024)
	n, err := device.ReadRaw(buffer)

	// Verify the result
	assert.NoError(t, err)
	assert.Equal(t, 9, n)
	assert.Equal(t, "test data", string(buffer[:9]))
	serialPort.AssertCalled(t, "Read", buffer)
}

// TestTOP708DeviceReadRawNotConnected tests the ReadRaw method when not connected
func TestTOP708DeviceReadRawNotConnected(t *testing.T) {
	// Create a mock serial port
	serialPort := new(MockSerialPort)

	// Create a new TOP708 device
	device := NewTOP708Device(serialPort)

	// Read data from the device
	buffer := make([]byte, 1024)
	_, err := device.ReadRaw(buffer)

	// Verify the result
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "device not connected")
}

// TestTOP708DeviceWriteRaw tests the WriteRaw method
func TestTOP708DeviceWriteRaw(t *testing.T) {
	// Create a mock serial port
	serialPort := new(MockSerialPort)
	serialPort.connected = true

	// Setup the mock to store the written data and return the length
	serialPort.On("Write", mock.Anything).Run(func(args mock.Arguments) {
		data := args.Get(0).([]byte)
		serialPort.written = append(serialPort.written, data...)
	}).Return(9, nil)

	// Create a new TOP708 device
	device := NewTOP708Device(serialPort)
	device.connected = true

	// Write data to the device
	data := []byte("test data")
	n, err := device.WriteRaw(data)

	// Verify the result
	assert.NoError(t, err)
	assert.Equal(t, 9, n)
	serialPort.AssertCalled(t, "Write", data)
	assert.Equal(t, data, serialPort.written)
}

// TestTOP708DeviceWriteRawNotConnected tests the WriteRaw method when not connected
func TestTOP708DeviceWriteRawNotConnected(t *testing.T) {
	// Create a mock serial port
	serialPort := new(MockSerialPort)

	// Create a new TOP708 device
	device := NewTOP708Device(serialPort)

	// Write data to the device
	data := []byte("test data")
	_, err := device.WriteRaw(data)

	// Verify the result
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "device not connected")
}

// TestTOP708DeviceWriteCommand tests the WriteCommand method
func TestTOP708DeviceWriteCommand(t *testing.T) {
	// Create a mock serial port
	serialPort := new(MockSerialPort)
	serialPort.connected = true

	// Setup the mock to store the written data and return the length
	serialPort.On("Write", mock.Anything).Run(func(args mock.Arguments) {
		data := args.Get(0).([]byte)
		serialPort.written = append(serialPort.written, data...)
	}).Return(14, nil)

	// Create a new TOP708 device
	device := NewTOP708Device(serialPort)
	device.connected = true

	// Write a command to the device
	err := device.WriteCommand("test command")

	// Verify the result
	assert.NoError(t, err)
	serialPort.AssertCalled(t, "Write", []byte("test command\r\n"))
	assert.Equal(t, []byte("test command\r\n"), serialPort.written)
}

// TestTOP708DeviceWriteCommandNotConnected tests the WriteCommand method when not connected
func TestTOP708DeviceWriteCommandNotConnected(t *testing.T) {
	// Create a mock serial port
	serialPort := new(MockSerialPort)

	// Create a new TOP708 device
	device := NewTOP708Device(serialPort)

	// Write a command to the device
	err := device.WriteCommand("test command")

	// Verify the result
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "device not connected")
}

// TestTOP708DeviceWriteCommandWithResponse tests the WriteCommandWithResponse method
func TestTOP708DeviceWriteCommandWithResponse(t *testing.T) {
	// Create a mock serial port
	serialPort := new(MockSerialPort)
	serialPort.connected = true

	// Setup the mock to store the written data and return the length
	serialPort.On("Write", mock.Anything).Run(func(args mock.Arguments) {
		data := args.Get(0).([]byte)
		serialPort.written = append(serialPort.written, data...)
	}).Return(14, nil)

	// Setup the mock to return a response
	serialPort.data = []byte("$PMTK001,314,3*36\r\n")
	serialPort.On("Read", mock.Anything).Run(func(args mock.Arguments) {
		buffer := args.Get(0).([]byte)
		copy(buffer, serialPort.data)
	}).Return(len(serialPort.data), nil)

	// Setup the mock for timeout
	serialPort.On("SetReadTimeout", mock.Anything).Return(nil)
	serialPort.On("GetTimeout").Return(500 * time.Millisecond)

	// Create a new TOP708 device
	device := NewTOP708Device(serialPort)
	device.connected = true

	// Write a command to the device and get the response
	response, err := device.WriteCommandWithResponse("$PMTK314,0,1,0,1,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0*28", 1*time.Second)

	// Verify the result
	assert.NoError(t, err)
	assert.Equal(t, "$PMTK001,314,3*36\r\n", response)
	serialPort.AssertCalled(t, "Write", []byte("$PMTK314,0,1,0,1,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0*28\r\n"))
}

// TestTOP708DeviceConfigureOutputMessages tests the ConfigureOutputMessages method
func TestTOP708DeviceConfigureOutputMessages(t *testing.T) {
	// Create a mock serial port
	serialPort := new(MockSerialPort)
	serialPort.connected = true

	// Setup the mock to store the written data and return the length
	serialPort.On("Write", mock.Anything).Run(func(args mock.Arguments) {
		data := args.Get(0).([]byte)
		serialPort.written = append(serialPort.written, data...)
	}).Return(50, nil)

	// Setup the mock to return a response
	serialPort.data = []byte("$PMTK001,314,3*36\r\n")
	serialPort.On("Read", mock.Anything).Run(func(args mock.Arguments) {
		buffer := args.Get(0).([]byte)
		copy(buffer, serialPort.data)
	}).Return(len(serialPort.data), nil)

	// Setup the mock for timeout
	serialPort.On("SetReadTimeout", mock.Anything).Return(nil)
	serialPort.On("GetTimeout").Return(500 * time.Millisecond)

	// Create a new TOP708 device
	device := NewTOP708Device(serialPort)
	device.connected = true

	// Configure output messages
	messages := map[string]bool{
		"GGA": true,
		"RMC": true,
		"GSA": false,
		"GSV": false,
	}
	err := device.ConfigureOutputMessages(messages)

	// Verify the result
	assert.NoError(t, err)
	serialPort.AssertCalled(t, "Write", mock.Anything)
}

// TestTOP708DeviceConfigureUpdateRate tests the ConfigureUpdateRate method
func TestTOP708DeviceConfigureUpdateRate(t *testing.T) {
	// Create a mock serial port
	serialPort := new(MockSerialPort)
	serialPort.connected = true

	// Setup the mock to store the written data and return the length
	serialPort.On("Write", mock.Anything).Run(func(args mock.Arguments) {
		data := args.Get(0).([]byte)
		serialPort.written = append(serialPort.written, data...)
	}).Return(20, nil)

	// Setup the mock to return a response
	serialPort.data = []byte("$PMTK001,220,3*30\r\n")
	serialPort.On("Read", mock.Anything).Run(func(args mock.Arguments) {
		buffer := args.Get(0).([]byte)
		copy(buffer, serialPort.data)
	}).Return(len(serialPort.data), nil)

	// Setup the mock for timeout
	serialPort.On("SetReadTimeout", mock.Anything).Return(nil)
	serialPort.On("GetTimeout").Return(500 * time.Millisecond)

	// Create a new TOP708 device
	device := NewTOP708Device(serialPort)
	device.connected = true

	// Configure update rate
	err := device.ConfigureUpdateRate(1000)

	// Verify the result
	assert.NoError(t, err)
	serialPort.AssertCalled(t, "Write", mock.Anything)
}

// TestTOP708DeviceConfigurePositioningMode tests the ConfigurePositioningMode method
func TestTOP708DeviceConfigurePositioningMode(t *testing.T) {
	// Create a mock serial port
	serialPort := new(MockSerialPort)
	serialPort.connected = true

	// Setup the mock to store the written data and return the length
	serialPort.On("Write", mock.Anything).Run(func(args mock.Arguments) {
		data := args.Get(0).([]byte)
		serialPort.written = append(serialPort.written, data...)
	}).Return(15, nil)

	// Setup the mock to return a response
	serialPort.data = []byte("$PMTK001,886,3*32\r\n")
	serialPort.On("Read", mock.Anything).Run(func(args mock.Arguments) {
		buffer := args.Get(0).([]byte)
		copy(buffer, serialPort.data)
	}).Return(len(serialPort.data), nil)

	// Setup the mock for timeout
	serialPort.On("SetReadTimeout", mock.Anything).Return(nil)
	serialPort.On("GetTimeout").Return(500 * time.Millisecond)

	// Create a new TOP708 device
	device := NewTOP708Device(serialPort)
	device.connected = true

	// Configure positioning mode
	err := device.ConfigurePositioningMode(PositioningModeVehicle)

	// Verify the result
	assert.NoError(t, err)
	serialPort.AssertCalled(t, "Write", mock.Anything)
}

// TestTOP708DeviceConfigureSatelliteSystems tests the ConfigureSatelliteSystems method
func TestTOP708DeviceConfigureSatelliteSystems(t *testing.T) {
	// Create a mock serial port
	serialPort := new(MockSerialPort)
	serialPort.connected = true

	// Setup the mock to store the written data and return the length
	serialPort.On("Write", mock.Anything).Run(func(args mock.Arguments) {
		data := args.Get(0).([]byte)
		serialPort.written = append(serialPort.written, data...)
	}).Return(20, nil)

	// Setup the mock to return a response
	serialPort.data = []byte("$PMTK001,353,3*37\r\n")
	serialPort.On("Read", mock.Anything).Run(func(args mock.Arguments) {
		buffer := args.Get(0).([]byte)
		copy(buffer, serialPort.data)
	}).Return(len(serialPort.data), nil)

	// Setup the mock for timeout
	serialPort.On("SetReadTimeout", mock.Anything).Return(nil)
	serialPort.On("GetTimeout").Return(500 * time.Millisecond)

	// Create a new TOP708 device
	device := NewTOP708Device(serialPort)
	device.connected = true

	// Configure satellite systems
	err := device.ConfigureSatelliteSystems(SatelliteSystemGPS | SatelliteSystemGLONASS)

	// Verify the result
	assert.NoError(t, err)
	serialPort.AssertCalled(t, "Write", mock.Anything)
}
