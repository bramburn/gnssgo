// Package stream provides stream input/output functionality for GNSS data
package stream

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// ntripRegistry is a registry for mapping between legacy NTrip and enhanced EnhancedNTrip instances
var ntripRegistry = struct {
	sync.RWMutex
	registry map[*NTrip]*EnhancedNTrip
}{
	registry: make(map[*NTrip]*EnhancedNTrip),
}

// RegisterEnhancedNTrip registers an enhanced NTRIP instance with a legacy NTRIP instance
func RegisterEnhancedNTrip(ntrip *NTrip, enhancedNtrip *EnhancedNTrip) {
	ntripRegistry.Lock()
	defer ntripRegistry.Unlock()
	ntripRegistry.registry[ntrip] = enhancedNtrip
}

// GetEnhancedNTripFromRegistry returns the enhanced NTRIP instance for a legacy NTRIP instance
func GetEnhancedNTripFromRegistry(ntrip *NTrip) *EnhancedNTrip {
	ntripRegistry.RLock()
	defer ntripRegistry.RUnlock()
	return ntripRegistry.registry[ntrip]
}

// UnregisterEnhancedNTrip removes an enhanced NTRIP instance from the registry
func UnregisterEnhancedNTrip(ntrip *NTrip) {
	ntripRegistry.Lock()
	defer ntripRegistry.Unlock()
	delete(ntripRegistry.registry, ntrip)
}

// NTRIP-specific constants
const (
	ntripAgent   = "GNSSGO NTRIP Client/1.0"
	ntripSvrPort = 80
	ntripCliPort = 2101
	ntripMaxRsp  = 32768 // max response buffer size
	ntripMaxStr  = 256   // max mountpoint string length
)

// NTRIP error types
var (
	ErrNTRIPNotConnected      = errors.New("not connected to NTRIP server")
	ErrNTRIPAlreadyConnected  = errors.New("already connected to NTRIP server")
	ErrNTRIPAuthFailed        = errors.New("NTRIP authentication failed")
	ErrNTRIPMountpointInvalid = errors.New("invalid NTRIP mountpoint")
	ErrNTRIPServerError       = errors.New("NTRIP server error")
	ErrNTRIPNetworkError      = errors.New("NTRIP network error")
	ErrNTRIPTimeout           = errors.New("NTRIP connection timeout")
)

// RTCMMessageStats contains statistics for RTCM messages
type RTCMMessageStats struct {
	MessageType  int       // RTCM message type
	Count        int       // Number of messages received
	LastReceived time.Time // Time of last message
	TotalBytes   int       // Total bytes received for this message type
}

// CircularBuffer implements a fixed-size circular buffer for RTCM messages
type CircularBuffer struct {
	buffer [][]byte   // Buffer to store messages
	size   int        // Size of the buffer
	head   int        // Head index
	tail   int        // Tail index
	count  int        // Number of items in the buffer
	mutex  sync.Mutex // Mutex for thread safety
}

// NewCircularBuffer creates a new circular buffer with the given size
func NewCircularBuffer(size int) *CircularBuffer {
	return &CircularBuffer{
		buffer: make([][]byte, size),
		size:   size,
		head:   0,
		tail:   0,
		count:  0,
	}
}

// Add adds an item to the circular buffer
func (c *CircularBuffer) Add(data []byte) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Make a copy of the data
	dataCopy := make([]byte, len(data))
	copy(dataCopy, data)

	// Add the item to the buffer
	c.buffer[c.head] = dataCopy
	c.head = (c.head + 1) % c.size

	// If the buffer is full, move the tail
	if c.count == c.size {
		c.tail = (c.tail + 1) % c.size
	} else {
		c.count++
	}
}

// GetAll returns all items in the circular buffer
func (c *CircularBuffer) GetAll() [][]byte {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	result := make([][]byte, c.count)
	for i := 0; i < c.count; i++ {
		idx := (c.tail + i) % c.size
		result[i] = make([]byte, len(c.buffer[idx]))
		copy(result[i], c.buffer[idx])
	}
	return result
}

// NTripConfig contains configuration for an NTRIP connection
type NTripConfig struct {
	Server       string        // Server address
	Port         int           // Server port
	Mountpoint   string        // Mountpoint
	Username     string        // Username
	Password     string        // Password
	UserAgent    string        // User agent
	ConnTimeout  time.Duration // Connection timeout
	RetryTimeout time.Duration // Retry timeout
	MaxRetries   int           // Maximum number of retries
	Debug        bool          // Debug mode
}

// We're using the NTrip struct from types.go

// EnhancedNTrip represents an enhanced NTRIP connection
type EnhancedNTrip struct {
	config        NTripConfig               // Configuration
	state         int                       // State (0:close, 1:wait, 2:connect)
	ctype         int                       // Type (0:server, 1:client)
	url           string                    // URL for proxy
	buff          string                    // Response buffer
	tcp           *TcpClient                // TCP client
	client        *http.Client              // HTTP client
	lastError     error                     // Last error
	retryCount    int                       // Retry count
	nextRetry     time.Time                 // Next retry time
	mutex         sync.Mutex                // Mutex for thread safety
	messageStats  map[int]*RTCMMessageStats // Message statistics
	messageBuffer *CircularBuffer           // Message buffer
	dataRate      float64                   // Data rate in bytes per second
	lastDataTime  time.Time                 // Last data time
	totalBytes    int                       // Total bytes received
	ctx           context.Context           // Context for cancellation
	cancel        context.CancelFunc        // Cancel function
}

// DefaultNTripConfig returns a default NTRIP configuration
func DefaultNTripConfig() NTripConfig {
	return NTripConfig{
		Port:         ntripCliPort,
		UserAgent:    ntripAgent,
		ConnTimeout:  30 * time.Second,
		RetryTimeout: 5 * time.Second,
		MaxRetries:   5,
		Debug:        false,
	}
}

// NewEnhancedNTrip creates a new enhanced NTRIP connection with the given configuration
func NewEnhancedNTrip(config NTripConfig, ctype int) *EnhancedNTrip {
	// Create context with timeout
	ctx, cancel := context.WithCancel(context.Background())

	// Create HTTP client with appropriate timeouts
	client := &http.Client{
		Timeout: config.ConnTimeout,
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   10 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			TLSHandshakeTimeout:   10 * time.Second,
			ResponseHeaderTimeout: 10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			MaxIdleConns:          10,
			IdleConnTimeout:       90 * time.Second,
		},
	}

	return &EnhancedNTrip{
		config:        config,
		state:         0,
		ctype:         ctype,
		client:        client,
		messageStats:  make(map[int]*RTCMMessageStats),
		messageBuffer: NewCircularBuffer(100), // Store last 100 messages
		lastDataTime:  time.Now(),
		ctx:           ctx,
		cancel:        cancel,
	}
}

// Connect establishes a connection to the NTRIP server with retry logic
func (ntrip *EnhancedNTrip) Connect() error {
	ntrip.mutex.Lock()
	defer ntrip.mutex.Unlock()

	// Check if already connected
	if ntrip.state == 2 {
		return ErrNTRIPAlreadyConnected
	}

	// Reset retry count if it's been a while since the last retry
	if time.Since(ntrip.nextRetry) > ntrip.config.RetryTimeout*2 {
		ntrip.retryCount = 0
	}

	// Check if we've exceeded the maximum number of retries
	if ntrip.retryCount >= ntrip.config.MaxRetries {
		// Calculate next retry time with exponential backoff
		backoff := time.Duration(math.Pow(2, float64(ntrip.retryCount))) * time.Second
		if backoff > 5*time.Minute {
			backoff = 5 * time.Minute // Cap at 5 minutes
		}
		ntrip.nextRetry = time.Now().Add(backoff)
		return fmt.Errorf("%w: exceeded maximum retries, next retry at %s",
			ErrNTRIPNetworkError, ntrip.nextRetry.Format(time.RFC3339))
	}

	// Construct the URL
	scheme := "http"
	url := fmt.Sprintf("%s://%s:%d/%s", scheme, ntrip.config.Server, ntrip.config.Port, ntrip.config.Mountpoint)

	// Create the request
	req, err := http.NewRequestWithContext(ntrip.ctx, "GET", url, nil)
	if err != nil {
		ntrip.retryCount++
		ntrip.lastError = fmt.Errorf("%w: failed to create request: %v", ErrNTRIPNetworkError, err)
		return ntrip.lastError
	}

	// Set headers
	req.Header.Set("User-Agent", ntrip.config.UserAgent)

	// Set basic auth if credentials are provided
	if ntrip.config.Username != "" {
		req.SetBasicAuth(ntrip.config.Username, ntrip.config.Password)
	}

	// Send the request
	resp, err := ntrip.client.Do(req)
	if err != nil {
		ntrip.retryCount++
		ntrip.lastError = fmt.Errorf("%w: failed to connect: %v", ErrNTRIPNetworkError, err)
		return ntrip.lastError
	}

	// Check the response status
	if resp.StatusCode != http.StatusOK {
		ntrip.retryCount++
		switch resp.StatusCode {
		case http.StatusUnauthorized:
			ntrip.lastError = fmt.Errorf("%w: invalid credentials", ErrNTRIPAuthFailed)
		case http.StatusNotFound:
			ntrip.lastError = fmt.Errorf("%w: mountpoint '%s' not found", ErrNTRIPMountpointInvalid, ntrip.config.Mountpoint)
		default:
			ntrip.lastError = fmt.Errorf("%w: server returned status %d", ErrNTRIPServerError, resp.StatusCode)
		}
		resp.Body.Close()
		return ntrip.lastError
	}

	// Connection successful
	ntrip.state = 2
	ntrip.retryCount = 0
	ntrip.lastError = nil
	ntrip.lastDataTime = time.Now()

	// Start a goroutine to read from the response body
	go ntrip.readResponseBody(resp.Body)

	return nil
}

// readResponseBody reads data from the response body in a separate goroutine
func (ntrip *EnhancedNTrip) readResponseBody(body io.ReadCloser) {
	defer body.Close()

	buffer := make([]byte, 4096)
	for {
		select {
		case <-ntrip.ctx.Done():
			// Context cancelled, exit
			return
		default:
			// Read data from the response body
			n, err := body.Read(buffer)
			if err != nil {
				if err != io.EOF {
					ntrip.mutex.Lock()
					ntrip.lastError = fmt.Errorf("%w: read error: %v", ErrNTRIPNetworkError, err)
					ntrip.state = 0
					ntrip.mutex.Unlock()
				}
				return
			}

			if n > 0 {
				ntrip.mutex.Lock()
				// Process the data
				ntrip.processData(buffer[:n])
				ntrip.mutex.Unlock()
			}
		}
	}
}

// RTCMMessage represents an RTCM message
type RTCMMessage struct {
	Type      int       // Message type
	Length    int       // Message length
	Data      []byte    // Message data
	Timestamp time.Time // Timestamp when the message was received
}

// parseRTCMMessage parses an RTCM message from a byte slice
func parseRTCMMessage(data []byte) ([]RTCMMessage, []byte) {
	var messages []RTCMMessage
	remaining := data

	for len(remaining) >= 3 {
		// Check for RTCM preamble (0xD3)
		if remaining[0] != 0xD3 {
			// Find the next preamble
			idx := 1
			for idx < len(remaining) && remaining[idx] != 0xD3 {
				idx++
			}
			remaining = remaining[idx:]
			continue
		}

		// Check if we have enough data to read the length
		if len(remaining) < 3 {
			break
		}

		// Extract message length (10 bits)
		length := (int(remaining[1])<<8 | int(remaining[2])) & 0x03FF

		// Check if we have the complete message
		if len(remaining) < length+6 { // 3 bytes header + length + 3 bytes CRC
			break
		}

		// Extract message type (12 bits)
		msgType := (int(remaining[3])<<4 | int(remaining[4])>>4)

		// Extract message data
		msgData := make([]byte, length)
		copy(msgData, remaining[3:3+length])

		// Create the message
		message := RTCMMessage{
			Type:      msgType,
			Length:    length,
			Data:      msgData,
			Timestamp: time.Now(),
		}

		// Add the message to the list
		messages = append(messages, message)

		// Move to the next message
		remaining = remaining[3+length+3:]
	}

	return messages, remaining
}

// processData processes received data
func (ntrip *EnhancedNTrip) processData(data []byte) {
	// Update statistics
	now := time.Now()
	elapsed := now.Sub(ntrip.lastDataTime).Seconds()
	if elapsed > 0 {
		ntrip.dataRate = float64(ntrip.totalBytes) / elapsed
	}
	ntrip.totalBytes += len(data)
	ntrip.lastDataTime = now

	// Add to message buffer
	ntrip.messageBuffer.Add(data)

	// Parse RTCM messages
	messages, _ := parseRTCMMessage(data)

	// Update message statistics
	for _, msg := range messages {
		// Get or create statistics for this message type
		stats, ok := ntrip.messageStats[msg.Type]
		if !ok {
			stats = &RTCMMessageStats{
				MessageType: msg.Type,
				Count:       0,
				TotalBytes:  0,
			}
			ntrip.messageStats[msg.Type] = stats
		}

		// Update statistics
		stats.Count++
		stats.LastReceived = msg.Timestamp
		stats.TotalBytes += msg.Length

		// Log message if debug is enabled
		if ntrip.config.Debug {
			fmt.Printf("RTCM message: type=%d, length=%d, time=%s\n",
				msg.Type, msg.Length, msg.Timestamp.Format(time.RFC3339Nano))
		}
	}
}

// OpenEnhancedNtrip opens an NTRIP connection with enhanced error handling and retry logic
// This is a replacement for the OpenNtrip function in stream_minimal.go
func OpenEnhancedNtrip(path string, ctype int, msg *string) *NTrip {
	// Parse the path
	var server, port, username, password, mountpoint string

	// Split the path into components
	// Format: [username]:[password]@[server]:[port]/[mountpoint]
	parts := strings.Split(path, "@")
	if len(parts) > 1 {
		// Extract username and password
		auth := strings.Split(parts[0], ":")
		if len(auth) > 1 {
			username = auth[0]
			password = auth[1]
		} else if len(auth) == 1 {
			username = auth[0]
		}

		// Extract server, port, and mountpoint
		serverPart := parts[1]
		serverPortParts := strings.Split(serverPart, "/")
		if len(serverPortParts) > 1 {
			serverPort := serverPortParts[0]
			mountpoint = strings.Join(serverPortParts[1:], "/")

			// Extract server and port
			serverPortSplit := strings.Split(serverPort, ":")
			if len(serverPortSplit) > 1 {
				server = serverPortSplit[0]
				port = serverPortSplit[1]
			} else {
				server = serverPort
			}
		} else {
			server = serverPart
		}
	} else {
		// No authentication
		serverPart := parts[0]
		serverPortParts := strings.Split(serverPart, "/")
		if len(serverPortParts) > 1 {
			serverPort := serverPortParts[0]
			mountpoint = strings.Join(serverPortParts[1:], "/")

			// Extract server and port
			serverPortSplit := strings.Split(serverPort, ":")
			if len(serverPortSplit) > 1 {
				server = serverPortSplit[0]
				port = serverPortSplit[1]
			} else {
				server = serverPort
			}
		} else {
			server = serverPart
		}
	}

	// Use default port if not specified
	portNum := ntripCliPort
	if ctype == 0 {
		portNum = ntripSvrPort
	}
	if port != "" {
		var err error
		portNum, err = strconv.Atoi(port)
		if err != nil {
			if msg != nil {
				*msg = fmt.Sprintf("Invalid port number: %s", port)
			}
			return nil
		}
	}

	// Create the configuration
	config := DefaultNTripConfig()
	config.Server = server
	config.Port = portNum
	config.Mountpoint = mountpoint
	config.Username = username
	config.Password = password

	// Create the enhanced NTRIP connection
	enhancedNtrip := NewEnhancedNTrip(config, ctype)
	if enhancedNtrip == nil {
		if msg != nil {
			*msg = "Failed to create NTRIP connection"
		}
		return nil
	}

	// Connect to the server
	err := enhancedNtrip.Connect()
	if err != nil {
		if msg != nil {
			*msg = err.Error()
		}
		return nil
	}

	// Create a legacy NTrip object for compatibility
	ntrip := &NTrip{
		state:  enhancedNtrip.state,
		ctype:  enhancedNtrip.ctype,
		url:    enhancedNtrip.url,
		buff:   enhancedNtrip.buff,
		tcp:    enhancedNtrip.tcp,
		mntpnt: config.Mountpoint,
		user:   config.Username,
		passwd: config.Password,
	}

	// Register the enhanced NTRIP instance with the legacy NTRIP instance
	RegisterEnhancedNTrip(ntrip, enhancedNtrip)

	return ntrip
}

// CloseNtrip closes an NTRIP connection
func (ntrip *EnhancedNTrip) CloseNtrip() {
	ntrip.mutex.Lock()
	defer ntrip.mutex.Unlock()

	// Cancel the context to stop any ongoing operations
	ntrip.cancel()

	// Close the TCP connection if it exists
	if ntrip.tcp != nil {
		ntrip.tcp.CloseTcpClient()
	}

	// Reset the state
	ntrip.state = 0
}

// ReadNtrip reads data from an NTRIP connection
func (ntrip *EnhancedNTrip) ReadNtrip(buff []byte, n int, msg *string) int {
	ntrip.mutex.Lock()
	defer ntrip.mutex.Unlock()

	// Check if connected
	if ntrip.state != 2 {
		if msg != nil {
			*msg = "Not connected to NTRIP server"
		}
		return 0
	}

	// Create a context with timeout for this read operation
	ctx, cancel := context.WithTimeout(ntrip.ctx, 500*time.Millisecond)
	defer cancel()

	// If we have a TCP client, use it directly
	if ntrip.tcp != nil {
		return ntrip.tcp.ReadTcpClient(buff, n, msg)
	}

	// Otherwise, get data from the message buffer
	messages := ntrip.messageBuffer.GetAll()
	if len(messages) == 0 {
		// Try to read from the HTTP response if available
		select {
		case <-ctx.Done():
			// Timeout or cancelled
			if msg != nil {
				*msg = "Read timeout"
			}
			return 0
		default:
			// No data available yet
			if msg != nil {
				*msg = "No data available"
			}
			return 0
		}
	}

	// Use the most recent message
	latestMsg := messages[len(messages)-1]

	// Copy data to the output buffer
	bytesToCopy := len(latestMsg)
	if bytesToCopy > n {
		bytesToCopy = n
	}

	copy(buff, latestMsg[:bytesToCopy])

	// Log the read operation if debug is enabled
	if ntrip.config.Debug {
		Tracet(4, "ReadNtrip: read %d bytes\n", bytesToCopy)
	}

	return bytesToCopy
}

// WriteNtrip writes data to an NTRIP connection
func (ntrip *EnhancedNTrip) WriteNtrip(buff []byte, n int, msg *string) int {
	ntrip.mutex.Lock()
	defer ntrip.mutex.Unlock()

	// Check if connected
	if ntrip.state != 2 {
		if msg != nil {
			*msg = "Not connected to NTRIP server"
		}
		return 0
	}

	// Create a context with timeout for this write operation
	ctx, cancel := context.WithTimeout(ntrip.ctx, 5*time.Second)
	defer cancel()

	// If we have a TCP client, use it directly
	if ntrip.tcp != nil {
		// Write data to the TCP connection
		bytesWritten := ntrip.tcp.WriteTcpClient(buff, n, msg)
		if bytesWritten <= 0 {
			if msg != nil && *msg == "" {
				*msg = "Failed to write data to TCP connection"
			}
			return 0
		}

		// Log the write operation if debug is enabled
		if ntrip.config.Debug {
			Tracet(4, "WriteNtrip: sent %d bytes\n", bytesWritten)
		}

		return bytesWritten
	}

	// Check if the data is a NMEA GGA message
	isGGA := false
	if n > 6 && string(buff[:6]) == "$GPGGA" {
		isGGA = true
	}

	// For HTTP-based NTRIP connections, we need to send a POST request
	// This is typically used for position reporting (GGA messages) in NTRIP clients
	// or for sending data to an NTRIP server

	// Create a new HTTP request for sending data
	var req *http.Request
	var err error

	// Construct the URL
	url := fmt.Sprintf("http://%s:%d/%s", ntrip.config.Server, ntrip.config.Port, ntrip.config.Mountpoint)

	// Create a POST request with the data
	req, err = http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(buff[:n]))
	if err != nil {
		if msg != nil {
			*msg = fmt.Sprintf("Failed to create POST request: %v", err)
		}
		return 0
	}

	// Set headers for NTRIP
	req.Header.Set("User-Agent", ntrip.config.UserAgent)
	req.Header.Set("Content-Type", "text/plain")

	// Set basic auth if credentials are provided
	if ntrip.config.Username != "" {
		req.SetBasicAuth(ntrip.config.Username, ntrip.config.Password)
	}

	// Send the request
	resp, err := ntrip.client.Do(req)
	if err != nil {
		if msg != nil {
			*msg = fmt.Sprintf("Failed to send data: %v", err)
		}
		return 0
	}
	defer resp.Body.Close()

	// Check the response status
	if resp.StatusCode != http.StatusOK {
		if msg != nil {
			*msg = fmt.Sprintf("Server returned status %d", resp.StatusCode)
		}
		return 0
	}

	// Log the write operation if debug is enabled
	if ntrip.config.Debug {
		if isGGA {
			Tracet(4, "WriteNtrip: sent GGA message (%d bytes)\n", n)
		} else {
			Tracet(4, "WriteNtrip: sent data (%d bytes)\n", n)
		}
	}

	return n
}

// GetMessageStats returns statistics for all RTCM messages
func (ntrip *EnhancedNTrip) GetMessageStats() map[int]*RTCMMessageStats {
	ntrip.mutex.Lock()
	defer ntrip.mutex.Unlock()

	// Create a copy of the statistics
	stats := make(map[int]*RTCMMessageStats)
	for k, v := range ntrip.messageStats {
		statsCopy := *v
		stats[k] = &statsCopy
	}

	return stats
}

// GetDataRate returns the current data rate in bytes per second
func (ntrip *EnhancedNTrip) GetDataRate() float64 {
	ntrip.mutex.Lock()
	defer ntrip.mutex.Unlock()

	return ntrip.dataRate
}

// GetLastMessages returns the last N messages received
func (ntrip *EnhancedNTrip) GetLastMessages() [][]byte {
	ntrip.mutex.Lock()
	defer ntrip.mutex.Unlock()

	return ntrip.messageBuffer.GetAll()
}

// GetState returns the current state of the connection
func (ntrip *EnhancedNTrip) GetState() int {
	ntrip.mutex.Lock()
	defer ntrip.mutex.Unlock()

	return ntrip.state
}

// GetLastError returns the last error that occurred
func (ntrip *EnhancedNTrip) GetLastError() error {
	ntrip.mutex.Lock()
	defer ntrip.mutex.Unlock()

	return ntrip.lastError
}

// SetDebug sets the debug mode
func (ntrip *EnhancedNTrip) SetDebug(debug bool) {
	ntrip.mutex.Lock()
	defer ntrip.mutex.Unlock()

	ntrip.config.Debug = debug
}

// Close closes the NTRIP connection
func (ntrip *EnhancedNTrip) Close() {
	ntrip.mutex.Lock()
	defer ntrip.mutex.Unlock()

	// Cancel the context to stop any ongoing operations
	if ntrip.cancel != nil {
		ntrip.cancel()
	}

	// Close the TCP connection if it exists
	if ntrip.tcp != nil {
		ntrip.tcp.CloseTcpClient()
	}

	// Reset the state
	ntrip.state = 0

	// Remove from registry if it's registered
	for k, v := range ntripRegistry.registry {
		if v == ntrip {
			UnregisterEnhancedNTrip(k)
			break
		}
	}
}
