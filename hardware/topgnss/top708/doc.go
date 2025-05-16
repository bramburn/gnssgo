/*
Package top708 provides functionality for working with TOPGNSS TOP708 GNSS receivers.

This package provides a complete interface for connecting to, configuring, and reading data from
TOPGNSS TOP708 GNSS receivers. It handles the details of serial communication, data parsing,
and device-specific commands.

# Main Components

## TOP708Device

The TOP708Device type provides a high-level interface for working with TOPGNSS TOP708 GNSS receivers.
It handles the details of establishing and maintaining the connection, sending commands, and reading data.

Example usage:

    // Create a new serial port
    serialPort := top708.NewGNSSSerialPort()
    
    // Create a new TOP708 device
    device := top708.NewTOP708Device(serialPort)
    
    // Connect to the device
    err := device.Connect("COM1", 38400)
    if err != nil {
        log.Fatalf("Failed to connect to device: %v", err)
    }
    defer device.Disconnect()
    
    // Verify the connection
    if !device.VerifyConnection(5 * time.Second) {
        log.Fatalf("Failed to verify connection")
    }
    
    // Read data from the device
    buffer := make([]byte, 1024)
    n, err := device.ReadRaw(buffer)
    if err != nil {
        log.Fatalf("Failed to read data: %v", err)
    }
    
    // Process the data
    fmt.Printf("Read %d bytes: %s\n", n, string(buffer[:n]))

## SerialPort

The SerialPort interface provides a generic interface for serial port operations.
The GNSSSerialPort type implements this interface for GNSS devices.

Example usage:

    // Create a new serial port
    serialPort := top708.NewGNSSSerialPort()
    
    // Open the port
    err := serialPort.Open("COM1", 38400)
    if err != nil {
        log.Fatalf("Failed to open port: %v", err)
    }
    defer serialPort.Close()
    
    // Read data from the port
    buffer := make([]byte, 1024)
    n, err := serialPort.Read(buffer)
    if err != nil {
        log.Fatalf("Failed to read data: %v", err)
    }
    
    // Process the data
    fmt.Printf("Read %d bytes: %s\n", n, string(buffer[:n]))

## Data Parsing

The package includes parsers for NMEA, RTCM, and UBX protocols.
These parsers can be used to extract structured data from the raw bytes received from the device.

Example usage:

    // Create a new NMEA parser
    parser := top708.NewNMEAParser()
    
    // Parse an NMEA sentence
    sentence := parser.Parse("$GPGGA,123519,4807.038,N,01131.000,E,1,08,0.9,545.4,M,46.9,M,,*47")
    
    // Check if the sentence is valid
    if sentence.Valid {
        fmt.Printf("Sentence type: %s\n", sentence.Type)
        fmt.Printf("Fields: %v\n", sentence.Fields)
    }

# Device Monitoring

The package provides functionality for monitoring the device and processing data in real-time.

Example usage:

    // Create a handler for NMEA data
    handler := &MyNMEAHandler{}
    
    // Create monitoring config
    config := top708.DefaultMonitorConfig(top708.ProtocolNMEA, handler)
    
    // Start monitoring
    err := device.MonitorNMEA(config)
    if err != nil {
        log.Fatalf("Failed to start monitoring: %v", err)
    }
    
    // Stop monitoring when done
    device.StopMonitoring()

Where MyNMEAHandler implements the DataHandler interface:

    type MyNMEAHandler struct{}
    
    func (h *MyNMEAHandler) HandleNMEA(sentence top708.NMEASentence) {
        fmt.Printf("Received NMEA sentence: %s\n", sentence.Raw)
    }
    
    func (h *MyNMEAHandler) HandleRTCM(message top708.RTCMMessage) {
        // Not used for NMEA monitoring
    }
    
    func (h *MyNMEAHandler) HandleUBX(message top708.UBXMessage) {
        // Not used for NMEA monitoring
    }
*/
package top708
