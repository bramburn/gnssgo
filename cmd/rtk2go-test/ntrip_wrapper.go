package main

import (
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/bramburn/gnssgo/pkg/gnssgo"
	"github.com/bramburn/gnssgo/pkg/ntrip"
)

// NTRIPClientImpl implements the NTRIPClient interface
type NTRIPClientImpl struct {
	server     string
	port       string
	username   string
	password   string
	mountpoint string
	connected  bool
	client     *ntrip.Client
	mutex      sync.Mutex
	buffer     []byte
	bufferPos  int
}

// CreateNTRIPClient creates a new NTRIP client
func CreateNTRIPClient(server, port, username, password, mountpoint string) (NTRIPClient, error) {
	// Create the underlying NTRIP client
	client, err := ntrip.NewClient(server, port, username, password, mountpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to create NTRIP client: %v", err)
	}

	return &NTRIPClientImpl{
		server:     server,
		port:       port,
		username:   username,
		password:   password,
		mountpoint: mountpoint,
		connected:  false,
		client:     client,
		buffer:     make([]byte, 4096),
		bufferPos:  0,
	}, nil
}

// Connect connects to the NTRIP server
func (c *NTRIPClientImpl) Connect() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.connected {
		return fmt.Errorf("already connected")
	}

	// Connect to the NTRIP server
	err := c.client.Connect()
	if err != nil {
		return fmt.Errorf("failed to connect to NTRIP server: %v", err)
	}

	c.connected = true
	return nil
}

// Disconnect disconnects from the NTRIP server
func (c *NTRIPClientImpl) Disconnect() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if !c.connected {
		return nil
	}

	// Disconnect from the NTRIP server
	err := c.client.Disconnect()
	if err != nil {
		return fmt.Errorf("failed to disconnect from NTRIP server: %v", err)
	}

	c.connected = false
	return nil
}

// IsConnected returns whether the client is connected
func (c *NTRIPClientImpl) IsConnected() bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.connected
}

// Read reads data from the NTRIP server
func (c *NTRIPClientImpl) Read(p []byte) (n int, err error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if !c.connected {
		return 0, fmt.Errorf("not connected")
	}

	// Read data from the NTRIP server
	n, err = c.client.Read(p)
	if err != nil && err != io.EOF {
		return 0, fmt.Errorf("failed to read from NTRIP server: %v", err)
	}

	return n, err
}

// Write writes data to the NTRIP server
func (c *NTRIPClientImpl) Write(p []byte) (n int, err error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if !c.connected {
		return 0, fmt.Errorf("not connected")
	}

	// Write data to the NTRIP server
	n, err = c.client.Write(p)
	if err != nil {
		return 0, fmt.Errorf("failed to write to NTRIP server: %v", err)
	}

	return n, nil
}

// RTKProcessorImpl implements the RTKProcessor interface
type RTKProcessorImpl struct {
	receiver     GNSSDevice
	client       NTRIPClient
	running      bool
	stopChan     chan struct{}
	rtcmProc     *gnssgo.Rtcm
	solutions    int
	fixRatio     float64
	mutex        sync.Mutex
	currentSol   RTKSolution
	lastGGATime  time.Time
	ggaInterval  time.Duration
	rtcmBuffer   []byte
	rtcmBufferMu sync.Mutex
}

// CreateRTKProcessor creates a new RTK processor
func CreateRTKProcessor(receiver GNSSDevice, client NTRIPClient) (RTKProcessor, error) {
	if receiver == nil {
		return nil, fmt.Errorf("receiver is nil")
	}
	if client == nil {
		return nil, fmt.Errorf("client is nil")
	}

	// Create and initialize RTCM processor
	rtcmProc := &gnssgo.Rtcm{}
	rtcmProc.InitRtcm()

	return &RTKProcessorImpl{
		receiver:    receiver,
		client:      client,
		running:     false,
		stopChan:    make(chan struct{}),
		rtcmProc:    rtcmProc,
		solutions:   0,
		fixRatio:    0.0,
		ggaInterval: 5 * time.Second, // Send GGA every 5 seconds
		rtcmBuffer:  make([]byte, 4096),
	}, nil
}

// Start starts the RTK processing
func (p *RTKProcessorImpl) Start() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.running {
		return fmt.Errorf("already running")
	}

	// Start the RTK processing
	p.stopChan = make(chan struct{})
	p.running = true

	// Start goroutine to read RTCM data from the NTRIP server
	go p.processRTCM()

	// Start goroutine to send GGA data to the NTRIP server
	go p.sendGGA()

	return nil
}

// Stop stops the RTK processing
func (p *RTKProcessorImpl) Stop() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if !p.running {
		return nil
	}

	// Stop the RTK processing
	close(p.stopChan)
	p.running = false

	return nil
}

// processRTCM reads RTCM data from the NTRIP server and processes it
func (p *RTKProcessorImpl) processRTCM() {
	buffer := make([]byte, 1024)

	for {
		select {
		case <-p.stopChan:
			return
		default:
			// Read data from the NTRIP server
			n, err := p.client.Read(buffer)
			if err != nil && err != io.EOF {
				fmt.Printf("Error reading from NTRIP server: %v\n", err)
				time.Sleep(1 * time.Second)
				continue
			}

			if n > 0 {
				// Store RTCM data for debugging
				p.rtcmBufferMu.Lock()
				copy(p.rtcmBuffer, buffer[:n])
				p.rtcmBufferMu.Unlock()

				// Process each byte of RTCM data
				for i := 0; i < n; i++ {
					ret := p.rtcmProc.InputRtcm3(buffer[i])
					if ret > 0 {
						// RTCM message processed successfully
						p.solutions++
						// Update fix ratio based on message type
						if ret == 1 { // Observation data
							p.fixRatio = float64(p.solutions) / 100.0
							if p.fixRatio > 1.0 {
								p.fixRatio = 1.0
							}
						}
					}
				}
			}

			time.Sleep(10 * time.Millisecond)
		}
	}
}

// sendGGA sends GGA data to the NTRIP server at regular intervals
func (p *RTKProcessorImpl) sendGGA() {
	ticker := time.NewTicker(p.ggaInterval)
	defer ticker.Stop()

	for {
		select {
		case <-p.stopChan:
			return
		case <-ticker.C:
			// Read data from the GNSS device
			buffer := make([]byte, 4096)
			n, err := p.receiver.ReadRaw(buffer)
			if err != nil {
				fmt.Printf("Error reading from GNSS device: %v\n", err)
				continue
			}

			if n > 0 {
				// Look for GGA sentences
				data := string(buffer[:n])
				lines := strings.Split(data, "\r\n")

				for _, line := range lines {
					if strings.Contains(line, "GGA") {
						// Send GGA sentence to the NTRIP server
						_, err := p.client.Write([]byte(line + "\r\n"))
						if err != nil {
							fmt.Printf("Error sending GGA to NTRIP server: %v\n", err)
						} else {
							p.lastGGATime = time.Now()
							fmt.Printf("Sent GGA to NTRIP server: %s\n", line)
						}
						break
					}
				}
			}
		}
	}
}

// GetSolution returns the current RTK solution
func (p *RTKProcessorImpl) GetSolution() RTKSolution {
	// Read data from the GNSS device
	buffer := make([]byte, 4096)
	n, err := p.receiver.ReadRaw(buffer)

	// Default solution with no position
	solution := RTKSolution{
		Status:    rtkStatusNone,
		Latitude:  0.0,
		Longitude: 0.0,
		Altitude:  0.0,
		NSats:     0,
		HDOP:      0.0,
		Age:       0.0,
	}

	if err == nil && n > 0 {
		// Process the NMEA data to extract position
		data := string(buffer[:n])
		lines := strings.Split(data, "\r\n")

		for _, line := range lines {
			if strings.Contains(line, "GGA") {
				// Parse the GGA sentence
				ggaData, err := gnssgo.ParseGGA(line)
				if err == nil {
					// Update solution with actual position data
					solution.Status = gnssgo.GetFixQualityName(ggaData.Quality)
					solution.Latitude = ggaData.Latitude
					solution.Longitude = ggaData.Longitude
					solution.Altitude = ggaData.Altitude
					solution.NSats = ggaData.NumSats
					solution.HDOP = ggaData.HDOP
					solution.Age = ggaData.DGPSAge
					break
				}
			}
		}
	}

	p.mutex.Lock()
	p.currentSol = solution
	p.mutex.Unlock()

	return solution
}

// GetStats returns the current RTK statistics
func (p *RTKProcessorImpl) GetStats() RTKStats {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	return RTKStats{
		Solutions: p.solutions,
		FixRatio:  p.fixRatio,
	}
}
