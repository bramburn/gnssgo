espackage main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/bramburn/gnssgo/hardware/topgnss/top708"
	"github.com/bramburn/gnssgo/pkg/gnssgo"
	"github.com/bramburn/gnssgo/pkg/ntrip"
)

// RTK status constants
const (
	rtkStatusNone   = "NONE"   // No position
	rtkStatusSingle = "SINGLE" // Single solution
	rtkStatusDGPS   = "DGPS"   // DGPS solution
	rtkStatusFloat  = "FLOAT"  // Float solution
	rtkStatusFix    = "FIX"    // Fixed solution
)

// ANSI color codes for terminal output
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorPurple = "\033[35m"
	colorCyan   = "\033[36m"
	colorWhite  = "\033[37m"
	colorBold   = "\033[1m"
)

// RTKStatus represents the current RTK status
type RTKStatus struct {
	Status    string    // Current RTK status (NONE, SINGLE, FLOAT, FIX)
	Latitude  float64   // Latitude in degrees
	Longitude float64   // Longitude in degrees
	Altitude  float64   // Altitude in meters
	NSats     int       // Number of satellites
	HDOP      float64   // Horizontal dilution of precision
	Age       float64   // Age of differential corrections in seconds
	Time      time.Time // Time of the last update
}

// RTKApp represents the RTK application
type RTKApp struct {
	ntripClient       *ntrip.Client
	gnssDevice        *top708.TOP708Device // Use TOP708Device directly
	rtkProcessor      *ntrip.RTKProcessor
	status            RTKStatus
	statusMutex       sync.Mutex
	stopChan          chan struct{}
	colorOutput       bool
	reconnect         bool
	reconnectInterval int
	nmeaBuffer        string // Buffer to accumulate NMEA data across multiple reads
}

func main() {
	// Parse command line flags
	ntripServer := flag.String("server", "rtk2go.com", "NTRIP server address")
	ntripPort := flag.String("port", "2101", "NTRIP server port")
	ntripUser := flag.String("user", "nitrogen@gmail.com", "NTRIP username (email address)")
	ntripPassword := flag.String("password", "password", "NTRIP password (any value for RTK2go)")
	ntripMountpoint := flag.String("mountpoint", "OCF-RH55LS-Capel", "NTRIP mountpoint (OCF-RH55LS-Capel, MEDW, ozzy1)")
	gnssPort := flag.String("gnss", "COM3", "GNSS receiver port")
	baudRate := flag.Int("baud", 38400, "GNSS receiver baud rate")
	duration := flag.Int("duration", 0, "Duration to run in seconds (0 for indefinite)")
	verbose := flag.Bool("verbose", false, "Enable verbose output")
	colorOutput := flag.Bool("color", true, "Enable colored output for RTK status")
	reconnect := flag.Bool("reconnect", true, "Automatically reconnect on connection loss")
	reconnectInterval := flag.Int("reconnect-interval", 5, "Reconnection interval in seconds")
	flag.Parse()

	// List available mountpoints
	availableMountpoints := []string{"OCF-RH55LS-Capel", "MEDW", "ozzy1"}

	// Validate mountpoint
	validMountpoint := false
	for _, m := range availableMountpoints {
		if *ntripMountpoint == m {
			validMountpoint = true
			break
		}
	}

	if !validMountpoint {
		fmt.Printf("Warning: Mountpoint '%s' is not in the list of known mountpoints: %v\n",
			*ntripMountpoint, availableMountpoints)
		fmt.Println("Continuing anyway, but connection might fail...")
	}

	// Set up logging
	log.SetFlags(log.Ltime | log.Ldate | log.Lshortfile)

	// Create a console logger
	consoleLogger := log.New(os.Stdout, "", log.Ltime)

	// Format the GNSS port with baud rate
	gnssPortWithBaud := fmt.Sprintf("%s:%d:8:N:1", *gnssPort, *baudRate)

	// Print connection information
	consoleLogger.Printf("RTK2go Test Client")
	consoleLogger.Printf("NTRIP Server: %s:%s", *ntripServer, *ntripPort)
	consoleLogger.Printf("NTRIP Mountpoint: %s", *ntripMountpoint)
	consoleLogger.Printf("GNSS Receiver: %s", gnssPortWithBaud)

	// Create the RTK application
	app := &RTKApp{
		stopChan: make(chan struct{}),
		status: RTKStatus{
			Status: rtkStatusNone,
			Time:   time.Now(),
		},
		colorOutput:       *colorOutput,
		reconnect:         *reconnect,
		reconnectInterval: *reconnectInterval,
	}

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		consoleLogger.Println("\nReceived shutdown signal")
		close(app.stopChan)
	}()

	// Connect to the GNSS receiver first
	consoleLogger.Printf("Connecting to GNSS receiver on port %s...\n", *gnssPort)

	// Create a new serial port
	serialPort := top708.NewGNSSSerialPort()

	// Create a new TOP708 device
	app.gnssDevice = top708.NewTOP708Device(serialPort)

	// Connect to the device
	err := app.gnssDevice.Connect(*gnssPort, *baudRate)
	if err != nil {
		consoleLogger.Fatalf("Failed to connect to device: %v", err)
	}
	defer app.gnssDevice.Disconnect()

	// Verify the connection
	if !app.gnssDevice.VerifyConnection(5 * time.Second) {
		consoleLogger.Fatalf("Failed to verify connection")
	}

	consoleLogger.Println("Connected to GNSS receiver successfully.")

	// Connect to the NTRIP server
	consoleLogger.Printf("Connecting to NTRIP server %s:%s...\n", *ntripServer, *ntripPort)
	app.ntripClient, err = ntrip.NewClient(*ntripServer, *ntripPort, *ntripUser, *ntripPassword, *ntripMountpoint)
	if err != nil {
		consoleLogger.Fatalf("Failed to create NTRIP client: %v", err)
	}

	// Try to connect to NTRIP server
	err = app.ntripClient.Connect()
	if err != nil {
		if app.reconnect {
			consoleLogger.Printf("Failed to connect to NTRIP server: %v", err)
			consoleLogger.Printf("Will retry connection every %d seconds...", app.reconnectInterval)
		} else {
			consoleLogger.Fatalf("Failed to connect to NTRIP server: %v", err)
		}
	} else {
		consoleLogger.Println("Connected to NTRIP server successfully.")
	}
	// We'll handle disconnection manually to avoid panic
	// DO NOT call Disconnect() on the NTRIP client because it causes a panic
	// The OS will clean up the resources when the process exits

	// Start the RTK processor
	consoleLogger.Println("Starting RTK processor...")

	// Create a ntrip.GNSSReceiver for the RTK processor
	ntripReceiver, err := ntrip.NewGNSSReceiver(*gnssPort)
	if err != nil {
		consoleLogger.Fatalf("Failed to create NTRIP receiver: %v", err)
	}

	app.rtkProcessor, err = ntrip.NewRTKProcessor(ntripReceiver, app.ntripClient)
	if err != nil {
		consoleLogger.Fatalf("Failed to create RTK processor: %v", err)
	}

	// Start the RTK processor but don't panic if it fails
	err = app.rtkProcessor.Start()
	if err != nil {
		consoleLogger.Printf("Warning: Failed to start RTK processing: %v", err)
		consoleLogger.Println("Continuing without RTK processing...")
	} else {
		consoleLogger.Println("RTK processor started successfully.")
	}
	// Don't use defer here to avoid potential panics during shutdown
	// We'll manually stop the processor in the shutdown sequence

	// Start connection monitoring if reconnection is enabled
	if app.reconnect {
		go app.monitorConnection(consoleLogger, *ntripServer, *ntripPort, *ntripUser, *ntripPassword, *ntripMountpoint)
	}

	// Start monitoring solutions
	go app.monitorSolutions(consoleLogger, *verbose)

	// Start a goroutine to directly read from the GNSS receiver for debugging
	go func() {
		buffer := make([]byte, 1024)
		for {
			select {
			case <-app.stopChan:
				return
			default:
				n, err := app.gnssDevice.ReadRaw(buffer)
				if err == nil && n > 0 {
					consoleLogger.Printf("Raw GNSS data: %s", string(buffer[:n]))
				}
				time.Sleep(5 * time.Second)
			}
		}
	}()

	// Run for the specified duration or until interrupted
	if *duration > 0 {
		select {
		case <-time.After(time.Duration(*duration) * time.Second):
			consoleLogger.Printf("Duration of %d seconds reached, shutting down...", *duration)
		case <-app.stopChan:
			// Shutdown signal received
		}
	} else {
		<-app.stopChan
	}

	consoleLogger.Println("Shutting down...")

	// Safely stop the RTK processor
	if app.rtkProcessor != nil {
		app.rtkProcessor.Stop()
	}

	// We're not calling Disconnect() on the NTRIP client because it causes a panic
	// The OS will clean up the resources when the process exits
	// If we need to explicitly clean up resources, we should implement a proper shutdown sequence
}

// monitorConnection monitors the NTRIP connection and attempts to reconnect if it fails
func (app *RTKApp) monitorConnection(logger *log.Logger, server, port, username, password, mountpoint string) {
	ticker := time.NewTicker(time.Duration(app.reconnectInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-app.stopChan:
			return
		case <-ticker.C:
			// Check if the client is connected
			if !app.ntripClient.IsConnected() {
				logger.Printf("NTRIP connection lost. Attempting to reconnect...")

				// Try to reconnect
				err := app.ntripClient.Connect()
				if err != nil {
					logger.Printf("Failed to reconnect to NTRIP server: %v", err)
					logger.Printf("Will retry in %d seconds...", app.reconnectInterval)
				} else {
					logger.Println("Successfully reconnected to NTRIP server.")

					// Restart the RTK processor if needed
					if app.rtkProcessor != nil {
						app.rtkProcessor.Stop()
						err = app.rtkProcessor.Start()
						if err != nil {
							logger.Printf("Failed to restart RTK processing: %v", err)
						} else {
							logger.Println("RTK processor restarted successfully.")
						}
					}
				}
			}
		}
	}
}

// monitorSolutions monitors RTK solutions and updates the status
func (app *RTKApp) monitorSolutions(logger *log.Logger, verbose bool) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	// Track status changes for notification
	var lastStatus string

	for {
		select {
		case <-app.stopChan:
			return
		case <-ticker.C:
			// Get the current solution from the RTK processor
			sol := app.rtkProcessor.GetSolution()
			now := time.Now()

			// Update the status
			app.statusMutex.Lock()

			// Update status based on solution status
			switch sol.Stat {
			case gnssgo.SOLQ_NONE:
				app.status.Status = rtkStatusNone
			case gnssgo.SOLQ_SINGLE:
				app.status.Status = rtkStatusSingle
			case gnssgo.SOLQ_FLOAT:
				app.status.Status = rtkStatusFloat
			case gnssgo.SOLQ_FIX:
				app.status.Status = rtkStatusFix
			default:
				app.status.Status = rtkStatusNone
			}

			// Read NMEA data directly from our TOP708Device
			buffer := make([]byte, 4096) // Larger buffer to capture more data
			n, err := app.gnssDevice.ReadRaw(buffer)

			foundValidPosition := false

			if err != nil {
				if verbose {
					logger.Printf("Error reading from GNSS device: %v", err)
				}
			} else if n > 0 {
				if verbose {
					logger.Printf("Read %d bytes from GNSS device", n)
				}

				// Append new data to the NMEA buffer
				app.nmeaBuffer += string(buffer[:n])

				// Limit the buffer size to prevent memory issues
				if len(app.nmeaBuffer) > 16384 { // 16KB limit
					app.nmeaBuffer = app.nmeaBuffer[len(app.nmeaBuffer)-16384:]
				}

				if verbose {
					logger.Printf("NMEA buffer size: %d bytes", len(app.nmeaBuffer))
				}

				// Split the buffer into lines and look for complete NMEA sentences
				lines := strings.Split(app.nmeaBuffer, "\r\n")

				for _, line := range lines {
					if strings.HasPrefix(line, "$") && strings.Contains(line, "GGA") {
						if verbose {
							logger.Printf("Found GGA sentence: %s", line)
						}

						// Parse GGA sentence
						fields := strings.Split(line, ",")
						if len(fields) >= 15 {
							// Extract fix quality to determine RTK status
							fixQuality := 0
							if fields[6] != "" {
								fixQuality, _ = strconv.Atoi(fields[6])
							}

							// Update RTK status based on fix quality
							switch fixQuality {
							case 0:
								app.status.Status = rtkStatusNone
							case 1:
								app.status.Status = rtkStatusSingle
							case 2:
								app.status.Status = rtkStatusDGPS
							case 4:
								app.status.Status = rtkStatusFix
							case 5:
								app.status.Status = rtkStatusFloat
							default:
								app.status.Status = rtkStatusNone
							}

							// Extract position
							if fields[2] != "" && fields[4] != "" {
								// Parse latitude
								lat, _ := strconv.ParseFloat(fields[2], 64)
								latDir := fields[3]
								if latDir == "S" {
									lat = -lat
								}

								// Parse longitude
								lon, _ := strconv.ParseFloat(fields[4], 64)
								lonDir := fields[5]
								if lonDir == "W" {
									lon = -lon
								}

								// Convert NMEA format (DDMM.MMMM) to decimal degrees
								latDeg := math.Floor(lat / 100.0)
								latMin := lat - latDeg*100.0
								app.status.Latitude = latDeg + latMin/60.0

								lonDeg := math.Floor(lon / 100.0)
								lonMin := lon - lonDeg*100.0
								app.status.Longitude = lonDeg + lonMin/60.0

								if verbose {
									logger.Printf("Calculated position: Lat: %f, Lon: %f", app.status.Latitude, app.status.Longitude)
								}

								// Parse altitude
								if fields[9] != "" {
									alt, _ := strconv.ParseFloat(fields[9], 64)
									app.status.Altitude = alt
								}

								// Extract number of satellites
								if fields[7] != "" {
									sats, _ := strconv.Atoi(fields[7])
									app.status.NSats = sats
								}

								// Extract HDOP
								if fields[8] != "" {
									hdop, _ := strconv.ParseFloat(fields[8], 64)
									app.status.HDOP = hdop
								}

								// Extract age of differential
								if fields[13] != "" {
									age, _ := strconv.ParseFloat(fields[13], 64)
									app.status.Age = age
								}

								foundValidPosition = true
								// Found a valid GGA sentence, no need to continue
								break
							}
						}
					}
				}
			}

			if !foundValidPosition {
				// If we couldn't parse any GGA sentences, use the RTK solution
				// but only if we don't already have a valid position
				if app.status.Latitude == 0 && app.status.Longitude == 0 {
					app.status.Latitude = sol.Pos[0]
					app.status.Longitude = sol.Pos[1]
					app.status.Altitude = sol.Pos[2]
					app.status.NSats = int(sol.Ns)
					app.status.HDOP = 0.8 + (float64(sol.Ns) * 0.02) // Estimate HDOP based on satellite count
					app.status.Age = float64(sol.Age)

					if verbose {
						logger.Printf("Using RTK solution as fallback: Lat: %f, Lon: %f", app.status.Latitude, app.status.Longitude)
					}
				} else {
					// Keep the last known position
					if verbose {
						logger.Printf("Keeping last known position: Lat: %f, Lon: %f", app.status.Latitude, app.status.Longitude)
					}
				}
			}

			app.status.Time = now

			// Check for status change
			statusChanged := lastStatus != app.status.Status
			lastStatus = app.status.Status

			app.statusMutex.Unlock()

			// Get status color
			statusColor := ""
			if app.colorOutput {
				switch app.status.Status {
				case rtkStatusNone:
					statusColor = colorRed
				case rtkStatusSingle:
					statusColor = colorYellow
				case rtkStatusFloat:
					statusColor = colorCyan
				case rtkStatusFix:
					statusColor = colorGreen
				}
			}

			// Format status with color if enabled
			var statusDisplay string
			if app.colorOutput {
				statusDisplay = fmt.Sprintf("%s%s%s%s%s",
					colorBold, statusColor, app.status.Status, colorReset, colorReset)
			} else {
				statusDisplay = app.status.Status
			}

			// Print status change notification
			if statusChanged {
				logger.Printf("RTK Status changed to: %s", statusDisplay)
			}

			// Print status
			statusStr := fmt.Sprintf("Status: %s | Lat: %.6f, Lon: %.6f, Alt: %.2fm | Sats: %d | Age: %.1fs",
				statusDisplay,
				app.status.Latitude,
				app.status.Longitude,
				app.status.Altitude,
				app.status.NSats,
				app.status.Age)

			// Print additional details if verbose
			if verbose {
				stats := app.rtkProcessor.GetStats()
				statusStr += fmt.Sprintf(" | Solutions: %d, Fix Ratio: %.2f%%",
					stats.Solutions,
					stats.FixRatio*100.0)
			}

			logger.Println(statusStr)
		}
	}
}
