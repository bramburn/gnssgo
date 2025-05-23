package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/bramburn/gnssgo/hardware/topgnss/top708"
	"github.com/bramburn/gnssgo/pkg/ntrip"
)

// CLI flags
var (
	portName        string
	baudRate        int
	timeout         time.Duration
	mode            string
	showPorts       bool
	enableRTK       bool
	ntripServer     string
	ntripPort       string
	ntripUser       string
	ntripPassword   string
	ntripMountpoint string
	showRTKStatus   bool
)

// Supported modes
const (
	ModeRaw  = "raw"
	ModeNMEA = "nmea"
	ModeRTCM = "rtcm"
	ModeUBX  = "ubx"
	ModeRTK  = "rtk"
)

// RTK fix quality indicators
const (
	FixNone      = "0" // No fix
	FixGPS       = "1" // GPS fix
	FixDGPS      = "2" // Differential GPS fix
	FixPPS       = "3" // PPS fix
	FixRTKFixed  = "4" // Real Time Kinematic (fixed)
	FixRTKFloat  = "5" // Real Time Kinematic (float)
	FixEstimated = "6" // Estimated fix
	FixManual    = "7" // Manual input mode
	FixSimulated = "8" // Simulated mode
)

// RTKStatus represents the current RTK status
type RTKStatus struct {
	Quality     string
	Description string
	Satellites  string
	HDOP        string
	LastUpdate  time.Time
	mutex       sync.Mutex
}

// Global RTK status
var rtkStatus = RTKStatus{
	Quality:     FixNone,
	Description: "No Fix",
	LastUpdate:  time.Now(),
}

func init() {
	// Define command-line flags
	flag.StringVar(&portName, "port", "", "Serial port name (e.g., COM1, /dev/ttyUSB0)")
	flag.IntVar(&baudRate, "baud", 38400, "Baud rate (default: 38400)")
	flag.DurationVar(&timeout, "timeout", 5*time.Second, "Connection verification timeout")
	flag.StringVar(&mode, "mode", ModeRaw, "Data mode: raw, nmea, rtcm, ubx, rtk")
	flag.BoolVar(&showPorts, "list", false, "List available ports and exit")

	// RTK-related flags
	flag.BoolVar(&enableRTK, "rtk", false, "Enable RTK correction")
	flag.StringVar(&ntripServer, "ntrip-server", "", "NTRIP server address")
	flag.StringVar(&ntripPort, "ntrip-port", "2101", "NTRIP server port")
	flag.StringVar(&ntripUser, "ntrip-user", "", "NTRIP username")
	flag.StringVar(&ntripPassword, "ntrip-password", "", "NTRIP password")
	flag.StringVar(&ntripMountpoint, "ntrip-mount", "", "NTRIP mountpoint")
	flag.BoolVar(&showRTKStatus, "show-rtk-status", false, "Show RTK status updates")

	flag.Parse()
}

func main() {
	// Create a new serial port
	serialPort := top708.NewGNSSSerialPort()

	// Create a new TOP708 device
	device := top708.NewTOP708Device(serialPort)

	// List available ports if requested
	if showPorts {
		listAvailablePorts(device)
		return
	}

	// If no port specified, prompt user to select one
	if portName == "" {
		var err error
		portName, err = selectPort(device)
		if err != nil {
			log.Fatalf("Error selecting port: %v", err)
		}
		if portName == "" {
			log.Fatal("No port selected. Exiting.")
		}
	}

	// Connect to the device
	fmt.Printf("Opening port %s with baud rate %d...\n", portName, baudRate)
	err := device.Connect(portName, baudRate)
	if err != nil {
		log.Fatalf("Failed to connect to device: %v", err)
	}
	defer device.Disconnect()

	fmt.Println("Port opened successfully. Waiting for device to initialize...")
	time.Sleep(2 * time.Second) // Give the device time to initialize

	// Verify connection
	fmt.Println("Verifying connection...")
	if !device.VerifyConnection(timeout) {
		fmt.Println("Unable to verify GNSS data. The device may not be sending data.")
		fmt.Println("Do you want to continue anyway? (y/n)")
		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(response)
		if strings.ToLower(response) != "y" {
			fmt.Println("Exiting...")
			return
		}
	} else {
		fmt.Println("Connection verified successfully.")
	}

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Check if RTK mode is enabled
	if strings.ToLower(mode) == ModeRTK || enableRTK {
		// Validate NTRIP settings
		if ntripServer == "" {
			fmt.Println("NTRIP server address is required for RTK mode.")
			fmt.Println("Please provide a server address with -ntrip-server flag.")
			return
		}

		if ntripMountpoint == "" {
			fmt.Println("NTRIP mountpoint is required for RTK mode.")
			fmt.Println("Please provide a mountpoint with -ntrip-mount flag.")
			return
		}

		// Start RTK monitoring
		fmt.Printf("Starting RTK monitoring with NTRIP server %s:%s. Press Ctrl+C to stop.\n",
			ntripServer, ntripPort)
		monitorRTK(device, sigChan)
	} else {
		// Start regular monitoring based on selected mode
		fmt.Printf("Starting %s data monitoring. Press Ctrl+C to stop.\n", mode)
		switch strings.ToLower(mode) {
		case ModeRaw:
			monitorRawData(device, sigChan)
		case ModeNMEA:
			monitorNMEA(device, sigChan)
		case ModeRTCM:
			monitorRTCM(device, sigChan)
		case ModeUBX:
			monitorUBX(device, sigChan)
		default:
			log.Fatalf("Unsupported mode: %s", mode)
		}
	}
}

// getFixQualityDescription returns a human-readable description of the fix quality
func getFixQualityDescription(quality string) string {
	switch quality {
	case FixNone:
		return "No Fix"
	case FixGPS:
		return "GPS Fix"
	case FixDGPS:
		return "DGPS Fix"
	case FixPPS:
		return "PPS Fix"
	case FixRTKFixed:
		return "RTK Fixed"
	case FixRTKFloat:
		return "RTK Float"
	case FixEstimated:
		return "Estimated Fix"
	case FixManual:
		return "Manual Input"
	case FixSimulated:
		return "Simulated Mode"
	default:
		return "Unknown"
	}
}

// updateRTKStatus updates the global RTK status
func updateRTKStatus(quality, satellites, hdop string) {
	rtkStatus.mutex.Lock()
	defer rtkStatus.mutex.Unlock()

	rtkStatus.Quality = quality
	rtkStatus.Description = getFixQualityDescription(quality)
	rtkStatus.Satellites = satellites
	rtkStatus.HDOP = hdop
	rtkStatus.LastUpdate = time.Now()

	if showRTKStatus {
		fmt.Printf("[RTK Status] %s, Satellites: %s, HDOP: %s\n",
			rtkStatus.Description, rtkStatus.Satellites, rtkStatus.HDOP)
	}
}

// RTKHandler implements the DataHandler interface for RTK data
type RTKHandler struct {
	ntripClient *ntrip.Client
	processor   *ntrip.RTKProcessor
}

// HandleNMEA handles NMEA sentences in RTK mode
func (h *RTKHandler) HandleNMEA(sentence top708.NMEASentence) {
	if sentence.Valid {
		fmt.Printf("[%s] %s\n", sentence.Type, sentence.Raw)

		// For GGA sentences, display position information and update RTK status
		if sentence.Type == "GGA" && len(sentence.Fields) >= 10 {
			lat := sentence.Fields[1]
			latDir := sentence.Fields[2]
			lon := sentence.Fields[3]
			lonDir := sentence.Fields[4]
			quality := sentence.Fields[5]
			satellites := sentence.Fields[6]
			hdop := sentence.Fields[7]
			altitude := sentence.Fields[8]
			altUnit := sentence.Fields[9]

			fmt.Printf("  Position: %s%s, %s%s\n", lat, latDir, lon, lonDir)
			fmt.Printf("  Quality: %s (%s), Satellites: %s, HDOP: %s\n",
				quality, getFixQualityDescription(quality), satellites, hdop)
			fmt.Printf("  Altitude: %s %s\n", altitude, altUnit)

			// Update RTK status
			updateRTKStatus(quality, satellites, hdop)
		}
	}
}

// HandleRTCM handles RTCM messages in RTK mode
func (h *RTKHandler) HandleRTCM(message top708.RTCMMessage) {
	fmt.Printf("RTCM Message - ID: %d, Length: %d bytes\n", message.MessageID, message.Length)
}

// HandleUBX handles UBX messages in RTK mode
func (h *RTKHandler) HandleUBX(message top708.UBXMessage) {
	fmt.Printf("UBX Message - Class: 0x%02X, ID: 0x%02X, Length: %d bytes\n",
		message.Class, message.ID, len(message.Payload))
}

// monitorRTK monitors GNSS data with RTK correction
func monitorRTK(device *top708.TOP708Device, sigChan chan os.Signal) {
	// Create NTRIP client
	ntripClient, err := ntrip.NewClient(ntripServer, ntripPort, ntripUser, ntripPassword, ntripMountpoint)
	if err != nil {
		log.Fatalf("Failed to create NTRIP client: %v", err)
	}

	// Connect to NTRIP server
	fmt.Printf("Connecting to NTRIP server %s:%s...\n", ntripServer, ntripPort)
	err = ntripClient.Connect()
	if err != nil {
		log.Fatalf("Failed to connect to NTRIP server: %v", err)
	}
	// Use defer with a function to handle potential nil pointer
	defer func() {
		if ntripClient != nil {
			ntripClient.Disconnect()
		}
	}()
	fmt.Println("Connected to NTRIP server successfully.")

	// Create GNSS receiver
	gnssReceiver, err := ntrip.NewGNSSReceiver(portName)
	if err != nil {
		log.Fatalf("Failed to create GNSS receiver: %v", err)
	}
	// Use defer with a function to handle potential nil pointer
	defer func() {
		if gnssReceiver != nil {
			gnssReceiver.Close()
		}
	}()

	// Create RTK processor
	fmt.Println("Starting RTK processor...")
	processor, err := ntrip.NewRTKProcessor(gnssReceiver, ntripClient)
	if err != nil {
		log.Fatalf("Failed to create RTK processor: %v", err)
	}

	// Start RTK processing
	err = processor.Start()
	if err != nil {
		log.Fatalf("Failed to start RTK processing: %v", err)
	}
	// Use defer with a function to handle potential nil pointer
	defer func() {
		if processor != nil {
			processor.Stop()
		}
	}()
	fmt.Println("RTK processor started successfully.")

	// Create RTK handler
	handler := &RTKHandler{
		ntripClient: ntripClient,
		processor:   processor,
	}

	// Start NMEA monitoring to get position updates
	config := top708.DefaultMonitorConfig(top708.ProtocolNMEA, handler)
	err = device.MonitorNMEA(config)
	if err != nil {
		log.Fatalf("Failed to start NMEA monitoring: %v", err)
	}
	defer device.StopMonitoring()

	// Start a goroutine to display RTK status periodically
	if showRTKStatus {
		go func() {
			ticker := time.NewTicker(5 * time.Second)
			defer ticker.Stop()

			for {
				select {
				case <-ticker.C:
					rtkStatus.mutex.Lock()
					fmt.Printf("[RTK Status] %s, Satellites: %s, HDOP: %s, Last Update: %s\n",
						rtkStatus.Description, rtkStatus.Satellites, rtkStatus.HDOP,
						rtkStatus.LastUpdate.Format("15:04:05"))
					rtkStatus.mutex.Unlock()
				case <-sigChan:
					return
				}
			}
		}()
	}

	// Wait for signal
	<-sigChan
	fmt.Println("\nStopped RTK monitoring.")
}

// listAvailablePorts lists all available serial ports
func listAvailablePorts(device *top708.TOP708Device) {
	details, err := device.GetPortDetails()
	if err != nil {
		log.Fatalf("Error getting port details: %v", err)
	}

	if len(details) == 0 {
		fmt.Println("No serial ports found.")
		return
	}

	fmt.Println("Available serial ports:")
	fmt.Println("------------------------")
	for i, detail := range details {
		if detail.IsUSB {
			fmt.Printf("%d. %s - USB Device [VID:PID=%04X:%04X] %s\n", i+1, detail.Name, detail.VID, detail.PID, detail.Product)
		} else {
			fmt.Printf("%d. %s\n", i+1, detail.Name)
		}
	}
}

// selectPort prompts the user to select a serial port
func selectPort(device *top708.TOP708Device) (string, error) {
	details, err := device.GetPortDetails()
	if err != nil {
		return "", fmt.Errorf("error getting port details: %w", err)
	}

	if len(details) == 0 {
		return "", fmt.Errorf("no serial ports found")
	}

	fmt.Println("Available serial ports:")
	for i, detail := range details {
		if detail.IsUSB {
			fmt.Printf("%d. %s - USB Device [VID:PID=%04X:%04X] %s\n", i+1, detail.Name, detail.VID, detail.PID, detail.Product)
		} else {
			fmt.Printf("%d. %s\n", i+1, detail.Name)
		}
	}

	fmt.Print("Select a port (1-" + fmt.Sprint(len(details)) + "): ")
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	// Convert input to integer
	var index int
	_, err = fmt.Sscanf(input, "%d", &index)
	if err != nil || index < 1 || index > len(details) {
		return "", fmt.Errorf("invalid selection")
	}

	return details[index-1].Name, nil
}

// monitorRawData continuously reads and displays raw data from the device
func monitorRawData(device *top708.TOP708Device, sigChan chan os.Signal) {
	buffer := make([]byte, 1024)
	done := make(chan bool)

	go func() {
		for {
			select {
			case <-done:
				return
			default:
				n, err := device.ReadRaw(buffer)
				if err != nil {
					fmt.Printf("Error reading data: %v\n", err)
					time.Sleep(500 * time.Millisecond)
					continue
				}

				if n > 0 {
					fmt.Print(string(buffer[:n]))
				}
			}
		}
	}()

	// Wait for signal
	<-sigChan
	done <- true
	fmt.Println("\nStopped monitoring.")
}

// NMEAHandler implements the DataHandler interface for NMEA data
type NMEAHandler struct{}

// HandleNMEA handles NMEA sentences
func (h *NMEAHandler) HandleNMEA(sentence top708.NMEASentence) {
	if sentence.Valid {
		fmt.Printf("[%s] %s\n", sentence.Type, sentence.Raw)

		// For GGA sentences, display position information
		if sentence.Type == "GGA" && len(sentence.Fields) >= 10 {
			lat := sentence.Fields[1]
			latDir := sentence.Fields[2]
			lon := sentence.Fields[3]
			lonDir := sentence.Fields[4]
			quality := sentence.Fields[5]
			satellites := sentence.Fields[6]
			hdop := sentence.Fields[7]
			altitude := sentence.Fields[8]
			altUnit := sentence.Fields[9]

			fmt.Printf("  Position: %s%s, %s%s\n", lat, latDir, lon, lonDir)
			fmt.Printf("  Quality: %s (%s), Satellites: %s, HDOP: %s\n",
				quality, getFixQualityDescription(quality), satellites, hdop)
			fmt.Printf("  Altitude: %s %s\n", altitude, altUnit)
		}
	}
}

// HandleRTCM handles RTCM messages (not used in NMEA mode)
func (h *NMEAHandler) HandleRTCM(message top708.RTCMMessage) {}

// HandleUBX handles UBX messages (not used in NMEA mode)
func (h *NMEAHandler) HandleUBX(message top708.UBXMessage) {}

// monitorNMEA monitors and parses NMEA sentences
func monitorNMEA(device *top708.TOP708Device, sigChan chan os.Signal) {
	handler := &NMEAHandler{}
	config := top708.DefaultMonitorConfig(top708.ProtocolNMEA, handler)

	err := device.MonitorNMEA(config)
	if err != nil {
		log.Fatalf("Failed to start NMEA monitoring: %v", err)
	}

	// Wait for signal
	<-sigChan
	device.StopMonitoring()
	fmt.Println("\nStopped monitoring.")
}

// RTCMHandler implements the DataHandler interface for RTCM data
type RTCMHandler struct{}

// HandleNMEA handles NMEA sentences (not used in RTCM mode)
func (h *RTCMHandler) HandleNMEA(sentence top708.NMEASentence) {}

// HandleRTCM handles RTCM messages
func (h *RTCMHandler) HandleRTCM(message top708.RTCMMessage) {
	fmt.Printf("RTCM Message - ID: %d, Length: %d bytes\n", message.MessageID, message.Length)
}

// HandleUBX handles UBX messages (not used in RTCM mode)
func (h *RTCMHandler) HandleUBX(message top708.UBXMessage) {}

// monitorRTCM monitors RTCM messages
func monitorRTCM(device *top708.TOP708Device, sigChan chan os.Signal) {
	// Since MonitorRTCM is not implemented yet, we'll use a manual approach
	buffer := make([]byte, 2048) // RTCM messages can be larger
	done := make(chan bool)

	go func() {
		for {
			select {
			case <-done:
				return
			default:
				n, err := device.ReadRaw(buffer)
				if err != nil {
					fmt.Printf("Error reading data: %v\n", err)
					time.Sleep(500 * time.Millisecond)
					continue
				}

				if n > 0 {
					// Look for RTCM message preamble (0xD3)
					data := buffer[:n]
					for i := 0; i < len(data); i++ {
						if data[i] == 0xD3 && i+2 < len(data) {
							// Potential RTCM message
							// In a real implementation, we would parse the RTCM message here
							fmt.Printf("Potential RTCM data detected at offset %d\n", i)
							// Skip ahead to avoid duplicate detections
							i += 2
						}
					}

					// Also print raw data in hex format for debugging
					fmt.Printf("Raw data (%d bytes): ", n)
					for i := 0; i < n && i < 20; i++ { // Show first 20 bytes max
						fmt.Printf("%02X ", data[i])
					}
					if n > 20 {
						fmt.Print("...")
					}
					fmt.Println()
				}

				time.Sleep(100 * time.Millisecond)
			}
		}
	}()

	// Wait for signal
	<-sigChan
	done <- true
	fmt.Println("\nStopped monitoring.")
}

// UBXHandler implements the DataHandler interface for UBX data
type UBXHandler struct{}

// HandleNMEA handles NMEA sentences (not used in UBX mode)
func (h *UBXHandler) HandleNMEA(sentence top708.NMEASentence) {}

// HandleRTCM handles RTCM messages (not used in UBX mode)
func (h *UBXHandler) HandleRTCM(message top708.RTCMMessage) {}

// HandleUBX handles UBX messages
func (h *UBXHandler) HandleUBX(message top708.UBXMessage) {
	fmt.Printf("UBX Message - Class: 0x%02X, ID: 0x%02X, Length: %d bytes\n",
		message.Class, message.ID, len(message.Payload))
}

// monitorUBX monitors UBX protocol messages
func monitorUBX(device *top708.TOP708Device, sigChan chan os.Signal) {
	// Since MonitorUBX is not implemented yet, we'll use a manual approach
	buffer := make([]byte, 1024)
	done := make(chan bool)

	go func() {
		// UBX message starts with 0xB5 0x62
		ubxHeader := []byte{0xB5, 0x62}
		ubxBuffer := make([]byte, 0)

		for {
			select {
			case <-done:
				return
			default:
				n, err := device.ReadRaw(buffer)
				if err != nil {
					fmt.Printf("Error reading data: %v\n", err)
					time.Sleep(500 * time.Millisecond)
					continue
				}

				if n > 0 {
					// Add new data to buffer
					ubxBuffer = append(ubxBuffer, buffer[:n]...)

					// Look for UBX message header
					for len(ubxBuffer) >= 2 {
						// Find UBX header
						headerIndex := -1
						for i := 0; i < len(ubxBuffer)-1; i++ {
							if ubxBuffer[i] == ubxHeader[0] && ubxBuffer[i+1] == ubxHeader[1] {
								headerIndex = i
								break
							}
						}

						if headerIndex == -1 {
							// No header found, keep the last byte in case it's the first byte of a header
							if len(ubxBuffer) > 1 {
								ubxBuffer = ubxBuffer[len(ubxBuffer)-1:]
							} else {
								ubxBuffer = make([]byte, 0)
							}
							break
						}

						// Remove data before header
						ubxBuffer = ubxBuffer[headerIndex:]

						// Check if we have enough data for a complete message
						if len(ubxBuffer) < 8 {
							// Not enough data for header + class + id + length, wait for more
							break
						}

						// Extract class, id, and length
						class := ubxBuffer[2]
						id := ubxBuffer[3]
						length := int(ubxBuffer[4]) | (int(ubxBuffer[5]) << 8)

						// Check if we have the complete message
						if len(ubxBuffer) < 8+length {
							// Not enough data for complete message, wait for more
							break
						}

						// We have a complete message
						fmt.Printf("UBX Message - Class: 0x%02X, ID: 0x%02X, Length: %d bytes\n",
							class, id, length)

						// Print payload in hex format (first 20 bytes max)
						fmt.Print("  Payload: ")
						for i := 0; i < length && i < 20; i++ {
							fmt.Printf("%02X ", ubxBuffer[6+i])
						}
						if length > 20 {
							fmt.Print("...")
						}
						fmt.Println()

						// Remove processed message from buffer
						ubxBuffer = ubxBuffer[8+length:]
					}
				}

				time.Sleep(100 * time.Millisecond)
			}
		}
	}()

	// Wait for signal
	<-sigChan
	done <- true
	fmt.Println("\nStopped monitoring.")
}
