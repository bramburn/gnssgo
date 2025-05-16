package ntrip

import (
	"errors"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockStream is a mock implementation of the gnssgo.Stream type
type MockStream struct {
	mock.Mock
	InitStreamCalled bool
	OpenStreamCalled bool
	OpenStreamArgs   []interface{}
	StreamReadCalled bool
	StreamReadArgs   []interface{}
	StreamReadReturn []interface{}
	State            int
	Msg              string
}

func (m *MockStream) InitStream() {
	m.InitStreamCalled = true
}

func (m *MockStream) OpenStream(strtype, strmode int, path string) int {
	m.OpenStreamCalled = true
	m.OpenStreamArgs = []interface{}{strtype, strmode, path}
	args := m.Called(strtype, strmode, path)
	return args.Int(0)
}

func (m *MockStream) StreamRead(buff []byte, n int) int {
	m.StreamReadCalled = true
	m.StreamReadArgs = []interface{}{buff, n}
	args := m.Called(buff, n)
	return args.Int(0)
}

func (m *MockStream) StreamClose() {
	m.Called()
}

// TestNewClient tests the NewClient function
func TestNewClient(t *testing.T) {
	// Test with valid parameters
	client, err := NewClient("example.com", "2101", "user", "pass", "MOUNT")
	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, "example.com", client.server)
	assert.Equal(t, "2101", client.port)
	assert.Equal(t, "user", client.username)
	assert.Equal(t, "pass", client.password)
	assert.Equal(t, "MOUNT", client.mountpoint)
	assert.False(t, client.connected)
}

// TestClientConnect tests the Connect method
func TestClientConnect(t *testing.T) {
	// Create a client with a mock stream
	client, _ := NewClient("example.com", "2101", "user", "pass", "MOUNT")
	
	// Replace the stream with our mock
	mockStream := new(MockStream)
	client.stream = mockStream
	
	// Setup mock expectations
	mockStream.On("OpenStream", mock.Anything, mock.Anything, mock.Anything).Return(1)
	mockStream.State = 1 // Simulate successful connection
	
	// Test connect
	err := client.Connect()
	assert.NoError(t, err)
	assert.True(t, client.connected)
	assert.True(t, mockStream.InitStreamCalled)
	assert.True(t, mockStream.OpenStreamCalled)
	
	// Verify the path was constructed correctly
	expectedPath := "user:pass@example.com:2101/MOUNT"
	assert.Equal(t, expectedPath, mockStream.OpenStreamArgs[2])
}

// TestClientConnectFailure tests the Connect method with a failure
func TestClientConnectFailure(t *testing.T) {
	// Create a client with a mock stream
	client, _ := NewClient("example.com", "2101", "user", "pass", "MOUNT")
	
	// Replace the stream with our mock
	mockStream := new(MockStream)
	client.stream = mockStream
	
	// Setup mock expectations for failure
	mockStream.On("OpenStream", mock.Anything, mock.Anything, mock.Anything).Return(0)
	mockStream.State = 0 // Simulate failed connection
	mockStream.Msg = "connection failed"
	
	// Test connect
	err := client.Connect()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to connect")
	assert.False(t, client.connected)
}

// TestClientDisconnect tests the Disconnect method
func TestClientDisconnect(t *testing.T) {
	// Create a client with a mock stream
	client, _ := NewClient("example.com", "2101", "user", "pass", "MOUNT")
	
	// Replace the stream with our mock
	mockStream := new(MockStream)
	client.stream = mockStream
	
	// Setup mock expectations
	mockStream.On("StreamClose").Return()
	
	// Set connected state
	client.connected = true
	
	// Test disconnect
	err := client.Disconnect()
	assert.NoError(t, err)
	assert.False(t, client.connected)
	mockStream.AssertCalled(t, "StreamClose")
}

// TestClientRead tests the Read method
func TestClientRead(t *testing.T) {
	// Create a client with a mock stream
	client, _ := NewClient("example.com", "2101", "user", "pass", "MOUNT")
	
	// Replace the stream with our mock
	mockStream := new(MockStream)
	client.stream = mockStream
	
	// Setup mock expectations
	testData := []byte("test data")
	mockStream.On("StreamRead", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		buffer := args.Get(0).([]byte)
		copy(buffer, testData)
	}).Return(len(testData))
	
	// Set connected state
	client.connected = true
	
	// Test read
	buffer := make([]byte, 1024)
	n, err := client.Read(buffer)
	assert.NoError(t, err)
	assert.Equal(t, len(testData), n)
	assert.Equal(t, testData, buffer[:n])
}

// TestClientReadNotConnected tests the Read method when not connected
func TestClientReadNotConnected(t *testing.T) {
	// Create a client
	client, _ := NewClient("example.com", "2101", "user", "pass", "MOUNT")
	
	// Test read when not connected
	buffer := make([]byte, 1024)
	_, err := client.Read(buffer)
	assert.Error(t, err)
	assert.Equal(t, "not connected", err.Error())
}

// TestClientIsConnected tests the IsConnected method
func TestClientIsConnected(t *testing.T) {
	// Create a client
	client, _ := NewClient("example.com", "2101", "user", "pass", "MOUNT")
	
	// Test when not connected
	assert.False(t, client.IsConnected())
	
	// Test when connected
	client.connected = true
	assert.True(t, client.IsConnected())
}
