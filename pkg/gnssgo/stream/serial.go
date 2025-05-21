// Package stream provides stream input/output functionality for GNSS data
package stream

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"go.bug.st/serial"
)

// Default serial port settings
const (
	defaultBaudRate = 9600
	defaultDataBits = 8
	defaultStopBits = 1
	defaultParity   = "N"
	defaultTimeout  = 100 * time.Millisecond
)

// OpenSerial opens a serial port
// path format: port[:brate[:bsize[:parity[:stopb[:fctr[#port]]]]]]
func OpenSerial(path string, modeFlag int, msg *string) *SerialComm {
	var (
		seri                          *SerialComm = new(SerialComm)
		brate, bsize, stopb, tcp_port int         = defaultBaudRate, defaultDataBits, defaultStopBits, 0
		parity                        rune        = 'N'
		port, fctr, path_tcp, msg_tcp string
		flowControl                   bool = false
	)

	Tracet(3, "OpenSerial: path=%s mode=%d\n", path, modeFlag)

	// Parse path format: port[:brate[:bsize[:parity[:stopb[:fctr[#port]]]]]]
	index := strings.Index(path, ":")
	if index > 0 {
		port = path[:index]
		parts := strings.Split(path[index+1:], ":")
		
		if len(parts) > 0 && parts[0] != "" {
			fmt.Sscanf(parts[0], "%d", &brate)
		}
		if len(parts) > 1 && parts[1] != "" {
			fmt.Sscanf(parts[1], "%d", &bsize)
		}
		if len(parts) > 2 && parts[2] != "" {
			fmt.Sscanf(parts[2], "%c", &parity)
		}
		if len(parts) > 3 && parts[3] != "" {
			fmt.Sscanf(parts[3], "%d", &stopb)
		}
		if len(parts) > 4 && parts[4] != "" {
			fctr = parts[4]
			
			// Check for TCP port
			hashIndex := strings.Index(fctr, "#")
			if hashIndex > 0 {
				fmt.Sscanf(fctr[hashIndex+1:], "%d", &tcp_port)
				fctr = fctr[:hashIndex]
			}
		}
	} else {
		port = path
	}

	// Validate parameters
	if brate <= 0 {
		brate = defaultBaudRate
	}
	if bsize <= 0 {
		bsize = defaultDataBits
	}
	if stopb <= 0 {
		stopb = defaultStopBits
	}

	// Set flow control
	if strings.Contains(strings.ToLower(fctr), "rts") {
		flowControl = true
	}

	// Create serial port mode
	serialMode := &serial.Mode{
		BaudRate: brate,
		DataBits: bsize,
		StopBits: serial.OneStopBit,
		Parity:   serial.NoParity,
	}

	// Set stop bits
	switch stopb {
	case 1:
		serialMode.StopBits = serial.OneStopBit
	case 2:
		serialMode.StopBits = serial.TwoStopBits
	default:
		serialMode.StopBits = serial.OneStopBit
	}

	// Set parity
	switch parity {
	case 'E', 'e':
		serialMode.Parity = serial.EvenParity
	case 'O', 'o':
		serialMode.Parity = serial.OddParity
	default:
		serialMode.Parity = serial.NoParity
	}

	// Store mode in SerialComm struct
	seri.mode = serialMode
	seri.timeout = defaultTimeout

	// Open the serial port
	s, err := serial.Open(port, serialMode)
	if err != nil {
		*msg = fmt.Sprintf("serial port open error: %s", err.Error())
		Tracet(1, "OpenSerial: %s path=%s\n", *msg, path)
		seri.err = 1
		return nil
	}

	// Set read timeout
	s.SetReadTimeout(seri.timeout)

	seri.serialio = s
	seri.err = 0
	seri.tcpsvr = nil

	// Open TCP server to output received stream if requested
	if tcp_port > 0 {
		path_tcp = fmt.Sprintf(":%d", tcp_port)
		seri.tcpsvr = OpenTcpSvr(path_tcp, &msg_tcp)
	}

	Tracet(3, "OpenSerial: port=%s baud=%d data=%d parity=%c stop=%d flow=%v\n",
		port, brate, bsize, parity, stopb, flowControl)
	return seri
}

// CloseSerial closes a serial port
func (seri *SerialComm) CloseSerial() {
	Tracet(3, "CloseSerial:\n")

	if seri == nil {
		return
	}

	// Close TCP server if open
	if seri.tcpsvr != nil {
		seri.tcpsvr.CloseTcpSvr()
		seri.tcpsvr = nil
	}

	// Close serial port
	if seri.serialio != nil {
		seri.serialio.Close()
		seri.serialio = nil
	}
}

// ReadSerial reads data from a serial port
func (seri *SerialComm) ReadSerial(buff []byte, n int, msg *string) int {
	var msg_tcp string

	Tracet(4, "ReadSerial: n=%d\n", n)

	if seri == nil || seri.serialio == nil {
		return 0
	}

	// Use mutex to ensure thread safety
	seri.lock.Lock()
	defer seri.lock.Unlock()

	// Read data from serial port
	nr, err := seri.serialio.Read(buff[:n])
	if err != nil {
		*msg = fmt.Sprintf("serial read error: %s", err.Error())
		seri.err = 1
		Tracet(2, "ReadSerial: error: %s\n", err.Error())
		return 0
	} else {
		seri.err = 0
	}

	// Forward data to TCP server if available
	if seri.tcpsvr != nil && nr > 0 {
		seri.tcpsvr.WriteTcpSvr(buff[:nr], nr, &msg_tcp)
	}

	Tracet(5, "ReadSerial: exit nr=%d\n", nr)
	return nr
}

// WriteSerial writes data to a serial port
func (seri *SerialComm) WriteSerial(buff []byte, n int, msg *string) int {
	Tracet(3, "WriteSerial: n=%d\n", n)

	if seri == nil || seri.serialio == nil {
		return 0
	}

	if n <= 0 {
		return 0
	}

	// Use mutex to ensure thread safety
	seri.lock.Lock()
	defer seri.lock.Unlock()

	// Write data to serial port
	ns, err := seri.serialio.Write(buff[:n])
	if err != nil {
		*msg = fmt.Sprintf("serial write error: %s", err.Error())
		seri.err = 1
		Tracet(2, "WriteSerial: error: %s\n", err.Error())
		return 0
	} else {
		seri.err = 0
	}

	Tracet(5, "WriteSerial: exit ns=%d\n", ns)
	return ns
}

// StateXSerial returns the state of a serial port
func (seri *SerialComm) StateXSerial(msg *string) int {
	return seri.err
}

// SetBrate sets the baud rate for a serial connection
func SetBrate(str *Stream, brate int) {
	var seri *SerialComm

	Tracet(3, "SetBrate: brate=%d\n", brate)

	if str == nil || str.Type != STR_SERIAL || str.Port == nil {
		return
	}

	seri = str.Port.(*SerialComm)
	if seri.serialio == nil {
		return
	}

	// Close and reopen the serial port with the new baud rate
	seri.serialio.Close()
	seri.mode.BaudRate = brate
	s, err := serial.Open(str.Path, seri.mode)
	if err != nil {
		Tracet(1, "SetBrate: serial port open error: %s\n", err.Error())
		seri.err = 1
		return
	}

	// Set read timeout
	s.SetReadTimeout(seri.timeout)
	seri.serialio = s
	seri.err = 0
}
