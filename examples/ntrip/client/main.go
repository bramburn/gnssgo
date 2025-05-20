package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
)

func main() {
	// Parse command-line flags
	host := flag.String("host", "localhost", "NTRIP caster host")
	port := flag.Int("port", 2101, "NTRIP caster port")
	mountpoint := flag.String("mountpoint", "RTCM33", "NTRIP caster mountpoint")
	username := flag.String("username", "user", "NTRIP caster username")
	password := flag.String("password", "password", "NTRIP caster password")
	outputFile := flag.String("output", "", "Output file for RTCM data (if not specified, data will be printed to console)")
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

	// Create a client
	client := &http.Client{}

	// Create a request
	url := fmt.Sprintf("http://%s:%d/%s", *host, *port, *mountpoint)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		logger.Fatalf("Failed to create request: %v", err)
	}

	// Set basic auth
	req.SetBasicAuth(*username, *password)

	// Set headers
	req.Header.Set("User-Agent", "GNSSGO NTRIP Client/1.0")

	// Send the request
	logger.Infof("Connecting to NTRIP caster at %s", url)
	resp, err := client.Do(req)
	if err != nil {
		logger.Fatalf("Failed to connect to NTRIP caster: %v", err)
	}
	defer resp.Body.Close()

	// Check the response
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		logger.Fatalf("Unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	logger.Infof("Connected to NTRIP caster at %s", url)

	// Open the output file if specified
	var output io.Writer = os.Stdout
	if *outputFile != "" {
		file, err := os.Create(*outputFile)
		if err != nil {
			logger.Fatalf("Failed to create output file: %v", err)
		}
		defer file.Close()
		output = file
		logger.Infof("Writing RTCM data to %s", *outputFile)
	} else {
		logger.Info("Writing RTCM data to console")
	}

	// Create a channel to signal when to stop
	stopCh := make(chan struct{})

	// Start reading data in a goroutine
	go func() {
		buffer := make([]byte, 1024)
		totalBytes := 0
		startTime := time.Now()

		for {
			select {
			case <-stopCh:
				return
			default:
				// Read data from the response
				n, err := resp.Body.Read(buffer)
				if err != nil {
					if err != io.EOF {
						logger.Errorf("Failed to read data: %v", err)
					}
					return
				}

				// Write data to the output
				if n > 0 {
					totalBytes += n
					
					// Print data as hex if going to console
					if *outputFile == "" {
						fmt.Printf("Received %d bytes: ", n)
						for i := 0; i < n && i < 16; i++ {
							fmt.Printf("%02X ", buffer[i])
						}
						if n > 16 {
							fmt.Print("...")
						}
						fmt.Println()
					} else {
						// Write binary data to file
						_, err := output.Write(buffer[:n])
						if err != nil {
							logger.Errorf("Failed to write data: %v", err)
							return
						}
					}

					// Log statistics every second
					if time.Since(startTime) >= time.Second {
						logger.Infof("Received %d bytes in the last second", totalBytes)
						totalBytes = 0
						startTime = time.Now()
					}
				}
			}
		}
	}()

	// Wait for interrupt signal
	logger.Info("Press Ctrl+C to exit")
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	// Stop the reader goroutine
	close(stopCh)
	logger.Info("Disconnected from NTRIP caster")
}
