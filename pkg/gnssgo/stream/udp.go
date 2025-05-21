// Package stream provides stream input/output functionality for GNSS data
package stream

import (
	"fmt"
	"net"
	"strconv"
	"time"
)

// Default UDP settings
const (
	defaultUdpPort     = 8000
	defaultUdpBuffSize = 32768 // UDP buffer size (bytes)
	defaultUdpTimeout  = 100   // Read timeout (ms)
)

// OpenUdpSvr opens a UDP server
// path format: :port
func OpenUdpSvr(path string, msg *string) *UdpConn {
	var (
		sport string
		port  int
	)

	Tracet(3, "OpenUdpSvr: path=%s\n", path)

	// Decode TCP path (reusing TCP path decoder)
	DecodeTcpPath(path, nil, &sport, nil, nil, nil, nil)

	// Parse port
	if len(sport) == 0 {
		port = defaultUdpPort
	} else {
		port, _ = strconv.Atoi(sport)
	}

	// Create UDP server
	return GenUdp(0, port, "localhost", msg)
}

// OpenUdpClient opens a UDP client
// path format: address:port
func OpenUdpClient(path string, msg *string) *UdpConn {
	var (
		saddr, sport string
		port         int
		err          error
	)

	Tracet(3, "OpenUdpClient: path=%s\n", path)

	// Decode TCP path (reusing TCP path decoder)
	DecodeTcpPath(path, &saddr, &sport, nil, nil, nil, nil)

	// Parse port
	if len(sport) == 0 {
		port = defaultUdpPort
	} else {
		if port, err = strconv.Atoi(sport); err != nil {
			*msg = fmt.Sprintf("port error: %s", sport)
			Tracet(2, "OpenUdpClient: port error port=%s\n", sport)
			return nil
		}
	}

	// Create UDP client
	return GenUdp(1, port, saddr, msg)
}

// GenUdp generates a UDP socket
func GenUdp(ctype, port int, saddr string, msg *string) *UdpConn {
	var (
		udp *UdpConn = new(UdpConn)
		err error
	)

	Tracet(3, "GenUdp: type=%d port=%d saddr=%s\n", ctype, port, saddr)

	// Initialize UDP connection
	udp.state = 2
	udp.ctype = ctype
	udp.port = port
	udp.saddr = saddr

	// Create UDP address
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", saddr, port))
	if err != nil {
		*msg = fmt.Sprintf("address resolution error: %s", err.Error())
		Tracet(2, "GenUdp: address error: %s\n", err.Error())
		return nil
	}

	// Create UDP socket based on type
	if ctype == 0 { // UDP server
		// Create UDP server socket
		conn, err := net.ListenUDP("udp", addr)
		if err != nil {
			*msg = fmt.Sprintf("bind error: %s", err.Error())
			Tracet(2, "GenUdp: bind error port=%d err=%s\n", port, err.Error())
			return nil
		}

		// Set buffer size
		conn.SetReadBuffer(defaultUdpBuffSize)
		conn.SetWriteBuffer(defaultUdpBuffSize)

		// Set read deadline
		conn.SetReadDeadline(time.Now().Add(time.Duration(defaultUdpTimeout) * time.Millisecond))

		udp.sock = conn
	} else { // UDP client
		// Create UDP client socket
		conn, err := net.DialUDP("udp", nil, addr)
		if err != nil {
			*msg = fmt.Sprintf("connect error: %s", err.Error())
			Tracet(2, "GenUdp: connect error addr=%s:%d err=%s\n", saddr, port, err.Error())
			return nil
		}

		// Set buffer size
		conn.SetReadBuffer(defaultUdpBuffSize)
		conn.SetWriteBuffer(defaultUdpBuffSize)

		// Set read deadline
		conn.SetReadDeadline(time.Now().Add(time.Duration(defaultUdpTimeout) * time.Millisecond))

		udp.sock = conn
	}

	return udp
}

// CloseUdp closes a UDP connection
func (udp *UdpConn) CloseUdp() {
	Tracet(3, "CloseUdp:\n")

	if udp == nil || udp.sock == nil {
		return
	}

	// Close socket
	udp.sock.Close()
	udp.state = 0
	udp.sock = nil
}

// ReadUdpSvr reads data from a UDP server
func (udp *UdpConn) ReadUdpSvr(buff []byte, n int, msg *string) int {
	var (
		nr  int
		err error
	)

	Tracet(4, "ReadUdpSvr: n=%d\n", n)

	if udp == nil || udp.sock == nil {
		return 0
	}

	// Check if this is a server
	if udp.ctype != 0 {
		return 0
	}

	// Set read deadline
	udp.sock.SetReadDeadline(time.Now().Add(time.Duration(defaultUdpTimeout) * time.Millisecond))

	// Read data
	switch conn := udp.sock.(type) {
	case *net.UDPConn:
		nr, _, err = conn.ReadFromUDP(buff[:n])
	default:
		nr, err = conn.Read(buff[:n])
	}

	if err != nil {
		// Ignore timeout errors
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			return 0
		}

		// Handle other errors
		*msg = fmt.Sprintf("udp read error: %s", err.Error())
		Tracet(2, "ReadUdpSvr: read error: %s\n", err.Error())
		return 0
	}

	return nr
}

// ReadUdpClient reads data from a UDP client
func (udp *UdpConn) ReadUdpClient(buff []byte, n int, msg *string) int {
	var (
		nr  int
		err error
	)

	Tracet(4, "ReadUdpClient: n=%d\n", n)

	if udp == nil || udp.sock == nil {
		return 0
	}

	// Check if this is a client
	if udp.ctype != 1 {
		return 0
	}

	// Set read deadline
	udp.sock.SetReadDeadline(time.Now().Add(time.Duration(defaultUdpTimeout) * time.Millisecond))

	// Read data
	nr, err = udp.sock.Read(buff[:n])
	if err != nil {
		// Ignore timeout errors
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			return 0
		}

		// Handle other errors
		*msg = fmt.Sprintf("udp read error: %s", err.Error())
		Tracet(2, "ReadUdpClient: read error: %s\n", err.Error())
		return 0
	}

	return nr
}

// WriteUdpSvr writes data to a UDP server
func (udp *UdpConn) WriteUdpSvr(buff []byte, n int, msg *string) int {
	Tracet(4, "WriteUdpSvr: n=%d\n", n)

	// UDP server doesn't write data (it only responds to clients)
	// This is a placeholder for future implementation if needed
	return 0
}

// WriteUdpClient writes data to a UDP client
func (udp *UdpConn) WriteUdpClient(buff []byte, n int, msg *string) int {
	var (
		ns  int
		err error
	)

	Tracet(4, "WriteUdpClient: n=%d\n", n)

	if udp == nil || udp.sock == nil {
		return 0
	}

	// Check if this is a client
	if udp.ctype != 1 {
		return 0
	}

	// Write data
	ns, err = udp.sock.Write(buff[:n])
	if err != nil {
		*msg = fmt.Sprintf("udp write error: %s", err.Error())
		Tracet(2, "WriteUdpClient: write error: %s\n", err.Error())
		return 0
	}

	return ns
}

// StatExUdpSvr returns the state of a UDP server
func (udp *UdpConn) StatExUdpSvr(msg *string) int {
	if udp == nil {
		return 0
	}
	return udp.state
}

// StateXUdpClient returns the state of a UDP client
func (udp *UdpConn) StateXUdpClient(msg *string) int {
	if udp == nil {
		return 0
	}
	return udp.state
}

// Update stream_minimal.go to use the implemented functions
func init() {
	// This function will be called when the package is imported
	// It's a placeholder for any initialization needed
}
