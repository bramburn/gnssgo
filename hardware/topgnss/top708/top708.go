package top708

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Logger defines a simple logging interface
type Logger interface {
	Printf(format string, v ...interface{})
	Debugf(format string, v ...interface{})
	Infof(format string, v ...interface{})
	Warnf(format string, v ...interface{})
	Errorf(format string, v ...interface{})
}

// DefaultLogger is a simple implementation of the Logger interface
type DefaultLogger struct{}

// Printf prints a formatted message
func (l *DefaultLogger) Printf(format string, v ...interface{}) {
	fmt.Printf(format, v...)
}

// Debugf prints a debug level formatted message
func (l *DefaultLogger) Debugf(format string, v ...interface{}) {
	fmt.Printf("[DEBUG] "+format, v...)
}

// Infof prints an info level formatted message
func (l *DefaultLogger) Infof(format string, v ...interface{}) {
	fmt.Printf("[INFO] "+format, v...)
}

// Warnf prints a warning level formatted message
func (l *DefaultLogger) Warnf(format string, v ...interface{}) {
	fmt.Printf("[WARN] "+format, v...)
}

// Errorf prints an error level formatted message
func (l *DefaultLogger) Errorf(format string, v ...interface{}) {
	fmt.Printf("[ERROR] "+format, v...)
}

// TOP708Device implements GNSSDevice interface for TOPGNSS TOP708
type TOP708Device struct {
	serialPort SerialPort
	connected  bool
	mutex      sync.Mutex
	stopChan   chan bool
	logger     Logger
	portName   string
	baudRate   int
	retryCount int
	retryDelay time.Duration
}

// NewTOP708Device creates a new TOPGNSS TOP708 device
func NewTOP708Device(serialPort SerialPort) *TOP708Device {
	return &TOP708Device{
		serialPort: serialPort,
		connected:  false,
		stopChan:   make(chan bool),
		logger:     &DefaultLogger{},
		retryCount: 3,
		retryDelay: 1 * time.Second,
	}
}

// SetLogger sets a custom logger for the device
func (d *TOP708Device) SetLogger(logger Logger) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	d.logger = logger
}

// SetRetryOptions sets the retry options for connection attempts
func (d *TOP708Device) SetRetryOptions(retryCount int, retryDelay time.Duration) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	d.retryCount = retryCount
	d.retryDelay = retryDelay
}

// Connect establishes a connection to the device
func (d *TOP708Device) Connect(portName string, baudRate int) error {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	if d.connected {
		d.logger.Debugf("Device already connected\n")
		return fmt.Errorf("device already connected")
	}

	// Use default baud rate if not specified
	if baudRate <= 0 {
		baudRate = 38400 // Default for TOPGNSS TOP708
		d.logger.Debugf("Using default baud rate: %d\n", baudRate)
	}

	// Store port name and baud rate for reconnection
	d.portName = portName
	d.baudRate = baudRate

	d.logger.Infof("Connecting to %s at %d baud...\n", portName, baudRate)

	// Try to connect with retry logic
	var err error
	for attempt := 0; attempt <= d.retryCount; attempt++ {
		if attempt > 0 {
			d.logger.Infof("Retrying connection (attempt %d/%d)...\n", attempt, d.retryCount)
			time.Sleep(d.retryDelay)
		}

		// Open the port
		err = d.serialPort.Open(portName, baudRate)
		if err == nil {
			d.connected = true
			d.logger.Infof("Successfully connected to %s\n", portName)
			return nil
		}

		d.logger.Warnf("Connection attempt %d failed: %v\n", attempt+1, err)
	}

	return fmt.Errorf("failed to connect to device after %d attempts: %w", d.retryCount+1, err)
}

// ConnectWithContext establishes a connection to the device with context for cancellation
func (d *TOP708Device) ConnectWithContext(ctx context.Context, portName string, baudRate int) error {
	// Create a channel to communicate the result
	resultCh := make(chan error, 1)

	// Start the connection process in a goroutine
	go func() {
		resultCh <- d.Connect(portName, baudRate)
	}()

	// Wait for either the context to be canceled or the connection to complete
	select {
	case <-ctx.Done():
		// Context was canceled, try to disconnect if needed
		d.Disconnect()
		return fmt.Errorf("connection canceled: %w", ctx.Err())
	case err := <-resultCh:
		return err
	}
}

// Disconnect closes the connection to the device
func (d *TOP708Device) Disconnect() error {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	if !d.connected {
		d.logger.Debugf("Device already disconnected\n")
		return nil
	}

	d.logger.Infof("Disconnecting from device...\n")

	// Stop any ongoing monitoring
	select {
	case d.stopChan <- true:
		d.logger.Debugf("Stopped monitoring\n")
	default:
		// No monitoring is active
	}

	err := d.serialPort.Close()
	if err != nil {
		d.logger.Errorf("Error disconnecting device: %v\n", err)
		return fmt.Errorf("error disconnecting device: %w", err)
	}

	d.connected = false
	d.logger.Infof("Successfully disconnected from device\n")
	return nil
}

// IsConnected returns whether the device is connected
func (d *TOP708Device) IsConnected() bool {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	return d.connected
}

// VerifyConnection checks if the device is sending valid GNSS data
func (d *TOP708Device) VerifyConnection(timeout time.Duration) bool {
	if !d.IsConnected() {
		d.logger.Warnf("Cannot verify connection: device not connected\n")
		return false
	}

	d.logger.Infof("Verifying connection with timeout of %v...\n", timeout)

	buffer := make([]byte, 1024)
	endTime := time.Now().Add(timeout)
	attempts := 0

	for time.Now().Before(endTime) {
		attempts++
		n, err := d.serialPort.Read(buffer)
		if err != nil {
			d.logger.Debugf("Read attempt %d failed: %v\n", attempts, err)
			time.Sleep(500 * time.Millisecond)
			continue
		}

		if n > 0 {
			data := string(buffer[:n])
			d.logger.Debugf("Read %d bytes\n", n)

			// Check for NMEA sentences
			if strings.Contains(data, "$GN") || strings.Contains(data, "$GP") {
				d.logger.Infof("Connection verified: valid NMEA data received\n")
				return true
			}

			d.logger.Debugf("Data received but no valid NMEA sentences found\n")
		}

		time.Sleep(500 * time.Millisecond)
	}

	d.logger.Warnf("Connection verification failed: no valid NMEA data received within timeout\n")
	return false
}

// VerifyConnectionWithContext checks if the device is sending valid GNSS data with context for cancellation
func (d *TOP708Device) VerifyConnectionWithContext(ctx context.Context, timeout time.Duration) bool {
	// Create a channel to communicate the result
	resultCh := make(chan bool, 1)

	// Start the verification process in a goroutine
	go func() {
		resultCh <- d.VerifyConnection(timeout)
	}()

	// Wait for either the context to be canceled or the verification to complete
	select {
	case <-ctx.Done():
		d.logger.Warnf("Connection verification canceled: %v\n", ctx.Err())
		return false
	case result := <-resultCh:
		return result
	}
}

// ReadRaw reads raw data from the device
func (d *TOP708Device) ReadRaw(buffer []byte) (int, error) {
	if !d.IsConnected() {
		err := errors.New("device not connected")
		d.logger.Errorf("ReadRaw failed: %v\n", err)
		return 0, err
	}

	n, err := d.serialPort.Read(buffer)
	if err != nil {
		d.logger.Debugf("ReadRaw: read %d bytes with error: %v\n", n, err)
	} else {
		d.logger.Debugf("ReadRaw: read %d bytes\n", n)
	}
	return n, err
}

// ReadRawWithTimeout reads raw data from the device with a timeout
func (d *TOP708Device) ReadRawWithTimeout(buffer []byte, timeout time.Duration) (int, error) {
	if !d.IsConnected() {
		err := errors.New("device not connected")
		d.logger.Errorf("ReadRawWithTimeout failed: %v\n", err)
		return 0, err
	}

	// Save the current timeout
	d.mutex.Lock()
	currentTimeout := d.serialPort.(interface{ GetTimeout() time.Duration }).GetTimeout()
	d.mutex.Unlock()

	// Set the new timeout
	err := d.serialPort.SetReadTimeout(timeout)
	if err != nil {
		d.logger.Errorf("Failed to set read timeout: %v\n", err)
		return 0, fmt.Errorf("failed to set read timeout: %w", err)
	}

	// Read data
	n, err := d.serialPort.Read(buffer)

	// Restore the original timeout
	restoreErr := d.serialPort.SetReadTimeout(currentTimeout)
	if restoreErr != nil {
		d.logger.Warnf("Failed to restore read timeout: %v\n", restoreErr)
	}

	if err != nil {
		d.logger.Debugf("ReadRawWithTimeout: read %d bytes with error: %v\n", n, err)
	} else {
		d.logger.Debugf("ReadRawWithTimeout: read %d bytes\n", n)
	}

	return n, err
}

// WriteRaw writes raw data to the device
func (d *TOP708Device) WriteRaw(data []byte) (int, error) {
	if !d.IsConnected() {
		err := errors.New("device not connected")
		d.logger.Errorf("WriteRaw failed: %v\n", err)
		return 0, err
	}

	n, err := d.serialPort.Write(data)
	if err != nil {
		d.logger.Errorf("WriteRaw: wrote %d bytes with error: %v\n", n, err)
	} else {
		d.logger.Debugf("WriteRaw: wrote %d bytes\n", n)
	}
	return n, err
}

// WriteCommand sends a command to the device
func (d *TOP708Device) WriteCommand(command string) error {
	if !d.IsConnected() {
		err := errors.New("device not connected")
		d.logger.Errorf("WriteCommand failed: %v\n", err)
		return err
	}

	// Add newline if not present
	if !strings.HasSuffix(command, "\r\n") {
		command += "\r\n"
	}

	d.logger.Debugf("Sending command: %s", command)
	n, err := d.serialPort.Write([]byte(command))
	if err != nil {
		d.logger.Errorf("Failed to send command: %v\n", err)
		return fmt.Errorf("failed to send command: %w", err)
	}

	d.logger.Debugf("Sent %d bytes\n", n)
	return nil
}

// WriteCommandWithResponse sends a command to the device and waits for a response
func (d *TOP708Device) WriteCommandWithResponse(command string, timeout time.Duration) (string, error) {
	if !d.IsConnected() {
		err := errors.New("device not connected")
		d.logger.Errorf("WriteCommandWithResponse failed: %v\n", err)
		return "", err
	}

	// Send the command
	err := d.WriteCommand(command)
	if err != nil {
		return "", err
	}

	// Read the response
	buffer := make([]byte, 1024)
	n, err := d.ReadRawWithTimeout(buffer, timeout)
	if err != nil {
		d.logger.Errorf("Failed to read response: %v\n", err)
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	response := string(buffer[:n])
	d.logger.Debugf("Received response: %s\n", response)
	return response, nil
}

// ChangeBaudRate changes the baud rate of the connection
func (d *TOP708Device) ChangeBaudRate(baudRate int) error {
	if !d.IsConnected() {
		err := errors.New("device not connected")
		d.logger.Errorf("ChangeBaudRate failed: %v\n", err)
		return err
	}

	d.logger.Infof("Changing baud rate to %d...\n", baudRate)

	// For TOPGNSS TOP708, we need to send a specific command to change the baud rate
	// The command format is: $PMTK251,<baud_rate>*<checksum>
	// For example: $PMTK251,115200*1F

	// First, try to send the command to the device
	baudCommand := fmt.Sprintf("$PMTK251,%d", baudRate)
	// Calculate checksum (NMEA checksum is XOR of all characters between $ and *)
	var checksum byte
	for i := 1; i < len(baudCommand); i++ {
		checksum ^= baudCommand[i]
	}
	baudCommand = fmt.Sprintf("%s*%02X", baudCommand, checksum)

	d.logger.Debugf("Sending baud rate command: %s\n", baudCommand)
	err := d.WriteCommand(baudCommand)
	if err != nil {
		d.logger.Warnf("Failed to send baud rate command: %v\n", err)
		// Continue with the port change anyway
	} else {
		// Wait a moment for the command to take effect
		time.Sleep(500 * time.Millisecond)
	}

	// Close and reopen the port with the new baud rate
	if d.portName == "" {
		err := errors.New("port name not available for reconnection")
		d.logger.Errorf("ChangeBaudRate failed: %v\n", err)
		return err
	}

	// Disconnect
	err = d.Disconnect()
	if err != nil {
		d.logger.Errorf("Failed to disconnect: %v\n", err)
		return fmt.Errorf("failed to disconnect: %w", err)
	}

	// Reconnect with new baud rate
	d.logger.Infof("Reconnecting with new baud rate %d...\n", baudRate)
	return d.Connect(d.portName, baudRate)
}

// GetAvailablePorts returns a list of available serial ports
func (d *TOP708Device) GetAvailablePorts() ([]string, error) {
	d.logger.Debugf("Getting available ports...\n")
	ports, err := d.serialPort.ListPorts()
	if err != nil {
		d.logger.Errorf("Failed to get available ports: %v\n", err)
		return nil, fmt.Errorf("failed to get available ports: %w", err)
	}

	d.logger.Debugf("Found %d available ports\n", len(ports))
	return ports, nil
}

// GetPortDetails returns detailed information about available ports
func (d *TOP708Device) GetPortDetails() ([]PortDetail, error) {
	d.logger.Debugf("Getting port details...\n")
	details, err := d.serialPort.GetPortDetails()
	if err != nil {
		d.logger.Errorf("Failed to get port details: %v\n", err)
		return nil, fmt.Errorf("failed to get port details: %w", err)
	}

	var result []PortDetail
	for _, detail := range details {
		// Convert string VID/PID to uint16 if they are USB devices
		vid := uint16(0)
		pid := uint16(0)

		if detail.IsUSB {
			// Parse hexadecimal VID/PID strings to uint16
			if vidVal, err := parseHexToUint16(detail.VID); err == nil {
				vid = vidVal
			} else {
				d.logger.Debugf("Failed to parse VID '%s': %v\n", detail.VID, err)
			}

			if pidVal, err := parseHexToUint16(detail.PID); err == nil {
				pid = pidVal
			} else {
				d.logger.Debugf("Failed to parse PID '%s': %v\n", detail.PID, err)
			}
		}

		result = append(result, PortDetail{
			Name:    detail.Name,
			IsUSB:   detail.IsUSB,
			VID:     vid,
			PID:     pid,
			Product: detail.Product,
		})

		d.logger.Debugf("Found port: %s, USB: %v, VID: %04X, PID: %04X, Product: %s\n",
			detail.Name, detail.IsUSB, vid, pid, detail.Product)
	}

	d.logger.Debugf("Found %d ports\n", len(result))
	return result, nil
}

// GetCurrentPortName returns the current port name
func (d *TOP708Device) GetCurrentPortName() string {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	return d.portName
}

// MonitorNMEA starts monitoring NMEA data
func (d *TOP708Device) MonitorNMEA(config MonitorConfig) error {
	if !d.IsConnected() {
		err := errors.New("device not connected")
		d.logger.Errorf("MonitorNMEA failed: %v\n", err)
		return err
	}

	d.logger.Infof("Starting NMEA monitoring with poll interval %v...\n", config.PollInterval)

	// Create NMEA parser
	nmeaParser := NewNMEAParser()
	buffer := make([]byte, config.BufferSize)
	dataBuffer := ""
	sentenceCount := 0
	errorCount := 0
	lastErrorTime := time.Time{}

	// Start monitoring in a goroutine
	go func() {
		d.logger.Debugf("NMEA monitoring goroutine started\n")

		for {
			select {
			case <-d.stopChan:
				d.logger.Infof("NMEA monitoring stopped\n")
				return
			default:
				n, err := d.serialPort.Read(buffer)
				if err != nil {
					// Only log errors if they're not too frequent (avoid flooding logs)
					if time.Since(lastErrorTime) > 5*time.Second {
						d.logger.Debugf("Read error: %v (suppressing similar errors for 5s)\n", err)
						lastErrorTime = time.Now()
						errorCount++
					}
					time.Sleep(config.PollInterval)
					continue
				}

				if n > 0 {
					// Add new data to buffer
					dataBuffer += string(buffer[:n])

					// Process complete NMEA sentences
					for {
						// Find start and end of NMEA sentence
						startIdx := strings.Index(dataBuffer, "$")
						if startIdx == -1 {
							break
						}

						endIdx := strings.Index(dataBuffer[startIdx:], "\r\n")
						if endIdx == -1 {
							break
						}
						endIdx += startIdx

						// Extract and parse the sentence
						sentence := dataBuffer[startIdx:endIdx]
						parsedSentence := nmeaParser.Parse(sentence)

						// Handle parsed data
						if parsedSentence.Valid && config.Handler != nil {
							sentenceCount++
							if sentenceCount%100 == 0 {
								d.logger.Debugf("Processed %d NMEA sentences, last type: %s\n",
									sentenceCount, parsedSentence.Type)
							}
							config.Handler.HandleNMEA(parsedSentence)
						} else if !parsedSentence.Valid {
							d.logger.Debugf("Invalid NMEA sentence: %s\n", sentence)
						}

						// Remove processed data from buffer
						if endIdx+2 <= len(dataBuffer) {
							dataBuffer = dataBuffer[endIdx+2:]
						} else {
							dataBuffer = ""
						}
					}
				}

				// If the buffer gets too large without finding complete sentences, trim it
				if len(dataBuffer) > config.BufferSize*2 {
					d.logger.Warnf("NMEA buffer overflow, trimming %d bytes\n", len(dataBuffer)-config.BufferSize)
					dataBuffer = dataBuffer[len(dataBuffer)-config.BufferSize:]
				}

				time.Sleep(config.PollInterval)
			}
		}
	}()

	d.logger.Infof("NMEA monitoring started successfully\n")
	return nil
}

// StopMonitoring stops all monitoring activities
func (d *TOP708Device) StopMonitoring() {
	d.logger.Infof("Stopping monitoring...\n")

	// Send stop signal with timeout to avoid blocking
	select {
	case d.stopChan <- true:
		d.logger.Debugf("Stop signal sent\n")
	case <-time.After(500 * time.Millisecond):
		d.logger.Warnf("Timed out sending stop signal, monitoring may already be stopped\n")
	}
}

// ConfigureOutputMessages configures which NMEA messages are output by the device
func (d *TOP708Device) ConfigureOutputMessages(messages map[string]bool) error {
	if !d.IsConnected() {
		err := errors.New("device not connected")
		d.logger.Errorf("ConfigureOutputMessages failed: %v\n", err)
		return err
	}

	d.logger.Infof("Configuring output messages...\n")

	// PMTK314 command format: $PMTK314,<GLL>,<RMC>,<VTG>,<GGA>,<GSA>,<GSV>,<GRS>,<GST>*<checksum>
	// 0 = disable, 1 = output once, 2 = output at interval
	// Example: $PMTK314,0,1,0,1,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0*28

	// Default values (all disabled)
	msgValues := map[string]int{
		"GLL": 0, "RMC": 0, "VTG": 0, "GGA": 0,
		"GSA": 0, "GSV": 0, "GRS": 0, "GST": 0,
	}

	// Update values based on input
	for msg, enabled := range messages {
		if _, exists := msgValues[strings.ToUpper(msg)]; exists {
			if enabled {
				msgValues[strings.ToUpper(msg)] = 1 // Output at interval
			} else {
				msgValues[strings.ToUpper(msg)] = 0 // Disable
			}
		} else {
			d.logger.Warnf("Unknown message type: %s\n", msg)
		}
	}

	// Build the command
	cmd := fmt.Sprintf("$PMTK314,%d,%d,%d,%d,%d,%d,%d,%d,0,0,0,0,0,0,0,0,0,0,0",
		msgValues["GLL"], msgValues["RMC"], msgValues["VTG"], msgValues["GGA"],
		msgValues["GSA"], msgValues["GSV"], msgValues["GRS"], msgValues["GST"])

	// Calculate checksum
	var checksum byte
	for i := 1; i < len(cmd); i++ {
		checksum ^= cmd[i]
	}
	cmd = fmt.Sprintf("%s*%02X", cmd, checksum)

	// Send the command
	d.logger.Debugf("Sending message configuration command: %s\n", cmd)
	response, err := d.WriteCommandWithResponse(cmd, 1*time.Second)
	if err != nil {
		d.logger.Errorf("Failed to configure output messages: %v\n", err)
		return fmt.Errorf("failed to configure output messages: %w", err)
	}

	// Check response
	if strings.Contains(response, "$PMTK001,314,3") {
		d.logger.Infof("Output messages configured successfully\n")
		return nil
	} else {
		err := fmt.Errorf("unexpected response: %s", response)
		d.logger.Errorf("Failed to configure output messages: %v\n", err)
		return err
	}
}

// ConfigureUpdateRate configures the update rate of the device in milliseconds
func (d *TOP708Device) ConfigureUpdateRate(rateMs int) error {
	if !d.IsConnected() {
		err := errors.New("device not connected")
		d.logger.Errorf("ConfigureUpdateRate failed: %v\n", err)
		return err
	}

	d.logger.Infof("Configuring update rate to %d ms...\n", rateMs)

	// PMTK220 command format: $PMTK220,<rate_ms>*<checksum>
	// Example: $PMTK220,1000*1F (1Hz)

	// Validate rate
	if rateMs < 100 || rateMs > 10000 {
		err := fmt.Errorf("invalid update rate: %d ms (must be between 100 and 10000 ms)", rateMs)
		d.logger.Errorf("ConfigureUpdateRate failed: %v\n", err)
		return err
	}

	// Build the command
	cmd := fmt.Sprintf("$PMTK220,%d", rateMs)

	// Calculate checksum
	var checksum byte
	for i := 1; i < len(cmd); i++ {
		checksum ^= cmd[i]
	}
	cmd = fmt.Sprintf("%s*%02X", cmd, checksum)

	// Send the command
	d.logger.Debugf("Sending update rate command: %s\n", cmd)
	response, err := d.WriteCommandWithResponse(cmd, 1*time.Second)
	if err != nil {
		d.logger.Errorf("Failed to configure update rate: %v\n", err)
		return fmt.Errorf("failed to configure update rate: %w", err)
	}

	// Check response
	if strings.Contains(response, "$PMTK001,220,3") {
		d.logger.Infof("Update rate configured successfully\n")
		return nil
	} else {
		err := fmt.Errorf("unexpected response: %s", response)
		d.logger.Errorf("Failed to configure update rate: %v\n", err)
		return err
	}
}

// PositioningMode defines the positioning mode for the device
type PositioningMode int

const (
	// PositioningModeNormal is the normal positioning mode
	PositioningModeNormal PositioningMode = 0
	// PositioningModeStationary is for stationary applications
	PositioningModeStationary PositioningMode = 1
	// PositioningModeWalking is optimized for pedestrian applications
	PositioningModeWalking PositioningMode = 2
	// PositioningModeVehicle is optimized for vehicle applications
	PositioningModeVehicle PositioningMode = 3
	// PositioningModeSea is optimized for sea applications
	PositioningModeSea PositioningMode = 4
	// PositioningModeAirborne is optimized for airborne applications
	PositioningModeAirborne PositioningMode = 5
)

// ConfigurePositioningMode configures the positioning mode of the device
func (d *TOP708Device) ConfigurePositioningMode(mode PositioningMode) error {
	if !d.IsConnected() {
		err := errors.New("device not connected")
		d.logger.Errorf("ConfigurePositioningMode failed: %v\n", err)
		return err
	}

	d.logger.Infof("Configuring positioning mode to %d...\n", mode)

	// PMTK886 command format: $PMTK886,<mode>*<checksum>
	// Example: $PMTK886,0*28 (Normal mode)

	// Validate mode
	if mode < 0 || mode > 5 {
		err := fmt.Errorf("invalid positioning mode: %d (must be between 0 and 5)", mode)
		d.logger.Errorf("ConfigurePositioningMode failed: %v\n", err)
		return err
	}

	// Build the command
	cmd := fmt.Sprintf("$PMTK886,%d", mode)

	// Calculate checksum
	var checksum byte
	for i := 1; i < len(cmd); i++ {
		checksum ^= cmd[i]
	}
	cmd = fmt.Sprintf("%s*%02X", cmd, checksum)

	// Send the command
	d.logger.Debugf("Sending positioning mode command: %s\n", cmd)
	response, err := d.WriteCommandWithResponse(cmd, 1*time.Second)
	if err != nil {
		d.logger.Errorf("Failed to configure positioning mode: %v\n", err)
		return fmt.Errorf("failed to configure positioning mode: %w", err)
	}

	// Check response
	if strings.Contains(response, "$PMTK001,886,3") {
		d.logger.Infof("Positioning mode configured successfully\n")
		return nil
	} else {
		err := fmt.Errorf("unexpected response: %s", response)
		d.logger.Errorf("Failed to configure positioning mode: %v\n", err)
		return err
	}
}

// SatelliteSystem represents a GNSS satellite system
type SatelliteSystem int

const (
	// SatelliteSystemGPS is the GPS satellite system
	SatelliteSystemGPS SatelliteSystem = 1
	// SatelliteSystemGLONASS is the GLONASS satellite system
	SatelliteSystemGLONASS SatelliteSystem = 2
	// SatelliteSystemGalileo is the Galileo satellite system
	SatelliteSystemGalileo SatelliteSystem = 4
	// SatelliteSystemBeiDou is the BeiDou satellite system
	SatelliteSystemBeiDou SatelliteSystem = 8
	// SatelliteSystemQZSS is the QZSS satellite system
	SatelliteSystemQZSS SatelliteSystem = 16
	// SatelliteSystemAll enables all satellite systems
	SatelliteSystemAll SatelliteSystem = 31
)

// ConfigureSatelliteSystems configures which satellite systems are used by the device
func (d *TOP708Device) ConfigureSatelliteSystems(systems SatelliteSystem) error {
	if !d.IsConnected() {
		err := errors.New("device not connected")
		d.logger.Errorf("ConfigureSatelliteSystems failed: %v\n", err)
		return err
	}

	d.logger.Infof("Configuring satellite systems to %d...\n", systems)

	// PMTK353 command format: $PMTK353,<GPS>,<GLONASS>,<Galileo>,<BeiDou>,<QZSS>*<checksum>
	// Example: $PMTK353,1,1,0,0,0*2A (GPS + GLONASS)

	// Extract individual system settings
	gps := 0
	if systems&SatelliteSystemGPS != 0 {
		gps = 1
	}

	glonass := 0
	if systems&SatelliteSystemGLONASS != 0 {
		glonass = 1
	}

	galileo := 0
	if systems&SatelliteSystemGalileo != 0 {
		galileo = 1
	}

	beidou := 0
	if systems&SatelliteSystemBeiDou != 0 {
		beidou = 1
	}

	qzss := 0
	if systems&SatelliteSystemQZSS != 0 {
		qzss = 1
	}

	// Build the command
	cmd := fmt.Sprintf("$PMTK353,%d,%d,%d,%d,%d", gps, glonass, galileo, beidou, qzss)

	// Calculate checksum
	var checksum byte
	for i := 1; i < len(cmd); i++ {
		checksum ^= cmd[i]
	}
	cmd = fmt.Sprintf("%s*%02X", cmd, checksum)

	// Send the command
	d.logger.Debugf("Sending satellite systems command: %s\n", cmd)
	response, err := d.WriteCommandWithResponse(cmd, 1*time.Second)
	if err != nil {
		d.logger.Errorf("Failed to configure satellite systems: %v\n", err)
		return fmt.Errorf("failed to configure satellite systems: %w", err)
	}

	// Check response
	if strings.Contains(response, "$PMTK001,353,3") {
		d.logger.Infof("Satellite systems configured successfully\n")
		return nil
	} else {
		err := fmt.Errorf("unexpected response: %s", response)
		d.logger.Errorf("Failed to configure satellite systems: %v\n", err)
		return err
	}
}

// parseHexToUint16 converts a hexadecimal string to uint16
func parseHexToUint16(hexStr string) (uint16, error) {
	// Remove 0x prefix if present
	hexStr = strings.TrimPrefix(hexStr, "0x")

	// Parse the hex string
	val, err := strconv.ParseUint(hexStr, 16, 16)
	if err != nil {
		return 0, err
	}

	return uint16(val), nil
}
