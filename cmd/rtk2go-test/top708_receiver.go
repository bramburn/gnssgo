package main

import (
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/bramburn/gnssgo/hardware/topgnss/top708"
)

// TOP708Receiver implements the io.ReadWriter interface for use with the RTK2go client
type TOP708Receiver struct {
	device    *top708.TOP708Device
	mutex     sync.Mutex
	connected bool
	portName  string
	baudRate  int
}

// NewTOP708Receiver creates a new TOP708Receiver
func NewTOP708Receiver(portName string, baudRate int) (*TOP708Receiver, error) {
	// Create a new serial port
	serialPort := top708.NewGNSSSerialPort()

	// Create a new TOP708 device
	device := top708.NewTOP708Device(serialPort)

	receiver := &TOP708Receiver{
		device:    device,
		connected: false,
		portName:  portName,
		baudRate:  baudRate,
	}

	// Connect to the device
	err := receiver.Connect()
	if err != nil {
		return nil, err
	}

	return receiver, nil
}

// Connect connects to the GNSS receiver
func (r *TOP708Receiver) Connect() error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if r.connected {
		return nil
	}

	// Connect to the device
	err := r.device.Connect(r.portName, r.baudRate)
	if err != nil {
		return fmt.Errorf("failed to connect to device: %w", err)
	}

	// Verify connection
	if !r.device.VerifyConnection(5 * time.Second) {
		r.device.Disconnect()
		return fmt.Errorf("unable to verify GNSS data")
	}

	r.connected = true
	return nil
}

// Close closes the connection to the GNSS receiver
func (r *TOP708Receiver) Close() error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if !r.connected {
		return nil
	}

	err := r.device.Disconnect()
	if err != nil {
		return fmt.Errorf("failed to disconnect from device: %w", err)
	}

	r.connected = false
	return nil
}

// Read reads data from the GNSS receiver
func (r *TOP708Receiver) Read(p []byte) (int, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if !r.connected {
		return 0, fmt.Errorf("receiver not connected")
	}

	n, err := r.device.ReadRaw(p)
	if err != nil {
		return 0, err
	}

	if n <= 0 {
		return 0, io.EOF
	}

	return n, nil
}

// Write writes data to the GNSS receiver (not used in this implementation)
func (r *TOP708Receiver) Write(p []byte) (int, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if !r.connected {
		return 0, fmt.Errorf("receiver not connected")
	}

	return r.device.WriteRaw(p)
}

// IsConnected returns true if the receiver is connected
func (r *TOP708Receiver) IsConnected() bool {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	return r.connected
}
