package server

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

// MockDataSource is a mock data source for testing
type MockDataSource struct {
	dataChan chan []byte
	running  bool
	data     []byte
}

// NewMockDataSource creates a new mock data source
func NewMockDataSource(data []byte) *MockDataSource {
	return &MockDataSource{
		dataChan: make(chan []byte, 10),
		data:     data,
	}
}

// Start starts the data source
func (ds *MockDataSource) Start() error {
	if ds.running {
		return nil
	}

	// Send the data to the channel
	ds.dataChan <- ds.data

	ds.running = true
	return nil
}

// Stop stops the data source
func (ds *MockDataSource) Stop() error {
	if !ds.running {
		return nil
	}

	// Close the data channel
	close(ds.dataChan)

	ds.running = false
	return nil
}

// Data returns the data channel
func (ds *MockDataSource) Data() <-chan []byte {
	return ds.dataChan
}

// MockCaster is a mock caster for testing
type MockCaster struct {
	server    *httptest.Server
	data      []byte
	dataReady chan struct{}
}

// NewMockCaster creates a new mock caster
func NewMockCaster() *MockCaster {
	return &MockCaster{
		data:      make([]byte, 0),
		dataReady: make(chan struct{}, 1),
	}
}

// Start starts the mock caster
func (c *MockCaster) Start() {
	// Create a handler for the caster
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if it's a POST request to a mountpoint
		if r.Method == http.MethodPost && r.URL.Path != "/" {
			// Read the request body
			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "Error reading request body", http.StatusInternalServerError)
				return
			}
			r.Body.Close()

			// Store the data
			c.data = body

			// Signal that data is ready
			select {
			case c.dataReady <- struct{}{}:
			default:
			}

			// Return a success response
			w.WriteHeader(http.StatusOK)
			return
		}

		// Return a 404 for any other request
		http.NotFound(w, r)
	})

	// Create a test server
	c.server = httptest.NewServer(handler)
}

// Stop stops the mock caster
func (c *MockCaster) Stop() {
	if c.server != nil {
		c.server.Close()
	}
}

// URL returns the URL of the mock caster
func (c *MockCaster) URL() string {
	return c.server.URL
}

// Data returns the data received by the caster
func (c *MockCaster) Data() []byte {
	return c.data
}

func TestServerStartStop(t *testing.T) {
	// Create a logger
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	// Create a server
	server := NewServer("localhost", "2101", "admin", "password", "TEST", logger)

	// Create a data source
	dataSource := &MockDataSource{
		dataChan: make(chan []byte, 10),
		data:     []byte("test data"),
	}

	// Set the data source
	server.SetDataSource(dataSource)

	// Start the server
	err := server.Start()
	assert.NoError(t, err)

	// Stop the server
	err = server.Stop()
	assert.NoError(t, err)
}

func TestServerNoDataSource(t *testing.T) {
	// Create a logger
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	// Create a server
	server := NewServer("localhost", "2101", "admin", "password", "TEST", logger)

	// Try to start the server without a data source
	err := server.Start()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no data source")
}
