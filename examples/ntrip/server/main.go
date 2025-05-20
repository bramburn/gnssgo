package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bramburn/gnssgo/pkg/caster"
	"github.com/bramburn/gnssgo/pkg/gnssgo"
	"github.com/bramburn/gnssgo/pkg/server"
	"github.com/sirupsen/logrus"
)

// GNSSDataSource is a data source that provides RTCM data from a GNSS receiver
type GNSSDataSource struct {
	stream     gnssgo.Stream
	dataChan   chan []byte
	running    bool
	port       string
	bufferSize int
	interval   time.Duration
	logger     *logrus.Logger
}

// NewGNSSDataSource creates a new GNSS data source
func NewGNSSDataSource(port string, bufferSize int, interval time.Duration, logger *logrus.Logger) *GNSSDataSource {
	return &GNSSDataSource{
		port:       port,
		dataChan:   make(chan []byte, 10),
		bufferSize: bufferSize,
		interval:   interval,
		logger:     logger,
	}
}

// Start starts the data source
func (ds *GNSSDataSource) Start() error {
	if ds.running {
		return nil
	}

	// Initialize the stream
	ds.stream.InitStream()

	// Open the serial port
	result := ds.stream.OpenStream(gnssgo.STR_SERIAL, gnssgo.STR_MODE_R, ds.port)

	// Check if the stream was opened successfully
	if result <= 0 || ds.stream.State <= 0 {
		return fmt.Errorf("failed to connect to GNSS receiver on %s", ds.port)
	}

	ds.logger.Infof("Connected to GNSS receiver on %s", ds.port)

	// Start reading data in a goroutine
	go ds.readData()

	ds.running = true
	return nil
}

// Stop stops the data source
func (ds *GNSSDataSource) Stop() error {
	if !ds.running {
		return nil
	}

	// Close the stream
	ds.stream.StreamClose()

	// Close the data channel
	close(ds.dataChan)

	ds.running = false
	return nil
}

// Data returns the data channel
func (ds *GNSSDataSource) Data() <-chan []byte {
	return ds.dataChan
}

// readData reads data from the GNSS receiver
func (ds *GNSSDataSource) readData() {
	buffer := make([]byte, ds.bufferSize)

	for ds.running {
		// Read data from the stream
		n := ds.stream.StreamRead(buffer, ds.bufferSize)
		if n <= 0 {
			// Wait before retrying
			time.Sleep(ds.interval)
			continue
		}

		// Copy the data to avoid race conditions
		data := make([]byte, n)
		copy(data, buffer[:n])

		// Log the data
		ds.logger.Debugf("Read %d bytes from GNSS receiver", n)

		// Send the data to the channel
		select {
		case ds.dataChan <- data:
		default:
			// Skip if the channel is full
		}

		// Wait before reading again
		time.Sleep(ds.interval)
	}
}

func main() {
	// Parse command-line flags
	casterPort := flag.Int("caster-port", 2101, "Port for the NTRIP caster")
	gnssPort := flag.String("gnss-port", "COM3:115200:8:N:1", "Serial port for the GNSS receiver")
	mountpoint := flag.String("mountpoint", "RTCM33", "Mountpoint for the NTRIP server")
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

	// Create a source service for the caster
	svc := caster.NewInMemorySourceService()
	svc.Sourcetable = caster.Sourcetable{
		Casters: []caster.CasterEntry{
			{
				Host:       "localhost",
				Port:       *casterPort,
				Identifier: "GNSSGO NTRIP Caster",
				Operator:   "GNSSGO",
				NMEA:       true,
				Country:    "USA",
				Latitude:   37.7749,
				Longitude:  -122.4194,
			},
		},
		Networks: []caster.NetworkEntry{
			{
				Identifier:          "GNSSGO",
				Operator:            "GNSSGO",
				Authentication:      "B",
				Fee:                 false,
				NetworkInfoURL:      "https://github.com/bramburn/gnssgo",
				StreamInfoURL:       "https://github.com/bramburn/gnssgo",
				RegistrationAddress: "admin@example.com",
			},
		},
		Mounts: []caster.StreamEntry{
			{
				Name:           *mountpoint,
				Identifier:     *mountpoint,
				Format:         "RTCM 3.3",
				FormatDetails:  "1004(1),1005/1006(5),1008(5),1012(1),1019(5),1020(5),1033(5),1042(5),1044(5),1045(5),1046(5)",
				Carrier:        "2",
				NavSystem:      "GPS+GLO+GAL+BDS+QZSS",
				Network:        "GNSSGO",
				CountryCode:    "USA",
				Latitude:       37.7749,
				Longitude:      -122.4194,
				NMEA:           true,
				Solution:       false,
				Generator:      "GNSSGO",
				Compression:    "none",
				Authentication: "B",
				Fee:            false,
				Bitrate:        9600,
			},
		},
	}

	// Create the caster
	caster := caster.NewCaster(fmt.Sprintf(":%d", *casterPort), svc, logger)

	// Start the caster in a goroutine
	go func() {
		logger.Infof("Starting NTRIP caster on port %d", *casterPort)
		if err := caster.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("Caster error: %v", err)
		}
	}()

	// Create a data source for the server
	dataSource := NewGNSSDataSource(*gnssPort, 1024, 100*time.Millisecond, logger)

	// Create the server
	server := server.NewServer("localhost", fmt.Sprintf("%d", *casterPort), "admin", "password", *mountpoint, logger)

	// Set the data source
	server.SetDataSource(dataSource)

	// Start the server
	if err := server.Start(); err != nil {
		logger.Fatalf("Failed to start server: %v", err)
	}
	defer server.Stop()

	logger.Infof("NTRIP server started and connected to caster at localhost:%d/%s", *casterPort, *mountpoint)
	logger.Infof("GNSS receiver connected on %s", *gnssPort)
	logger.Infof("Press Ctrl+C to exit")

	// Wait for interrupt signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	// Shutdown the caster
	logger.Info("Shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := caster.Shutdown(ctx); err != nil {
		logger.Errorf("Error shutting down caster: %v", err)
	}
}
