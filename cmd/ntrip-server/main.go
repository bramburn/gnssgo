package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bramburn/gnssgo/pkg/server"
	"github.com/sirupsen/logrus"
)

func main() {
	// Parse command-line flags
	host := flag.String("host", "localhost", "NTRIP caster host")
	port := flag.String("port", "2101", "NTRIP caster port")
	username := flag.String("username", "admin", "NTRIP caster username")
	password := flag.String("password", "password", "NTRIP caster password")
	mountpoint := flag.String("mountpoint", "RTCM33", "NTRIP caster mountpoint")
	filePath := flag.String("file", "", "Path to RTCM file (if not specified, random data will be generated)")
	logLevel := flag.String("log-level", "info", "Log level (debug, info, warn, error)")
	flag.Parse()

	// Configure logger
	logger := logrus.New()
	level, err := logrus.ParseLevel(*logLevel)
	if err != nil {
		logger.Fatalf("Invalid log level: %v", err)
	}
	logger.SetLevel(level)
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	// Create a new server
	srv := server.NewServer(*host, *port, *username, *password, *mountpoint, logger)

	// Create a data source
	var dataSource server.DataSource
	if *filePath != "" {
		dataSource = server.NewFileDataSource(*filePath, 1024, 1*time.Second)
	} else {
		// Create a mock data source that generates random RTCM data
		dataSource = &MockDataSource{
			dataChan: make(chan []byte, 10),
		}
	}

	// Set the data source
	srv.SetDataSource(dataSource)

	// Start the server
	if err := srv.Start(); err != nil {
		logger.Fatalf("Failed to start server: %v", err)
	}

	// Wait for interrupt signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	// Stop the server
	logger.Info("Shutting down server...")
	if err := srv.Stop(); err != nil {
		logger.Errorf("Error shutting down server: %v", err)
	}
}

// MockDataSource is a mock data source that generates random RTCM data
type MockDataSource struct {
	dataChan chan []byte
	running  bool
}

// Start starts the data source
func (ds *MockDataSource) Start() error {
	if ds.running {
		return nil
	}

	// Start generating data in a goroutine
	go ds.generateData()

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

// generateData generates random RTCM data
func (ds *MockDataSource) generateData() {
	// Generate a simple RTCM message header (not valid RTCM, just for testing)
	header := []byte{0xD3, 0x00, 0x01}

	for {
		// Generate a random message
		msg := make([]byte, 100)
		copy(msg, header)

		// Send the message
		select {
		case ds.dataChan <- msg:
		default:
			// Skip if the channel is full
		}

		// Wait before generating the next message
		time.Sleep(1 * time.Second)
	}
}
