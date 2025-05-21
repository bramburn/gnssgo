// Package stream provides stream input/output functionality for GNSS data
package stream

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

// Default TCP settings
const (
	defaultTcpPort     = 8000
	defaultTcpInactTO  = 10000 // Inactive timeout (ms)
	defaultTcpReconTO  = 10000 // Reconnect timeout (ms)
	defaultTcpBuffSize = 32768 // TCP buffer size (bytes)
	defaultConnTimeout = 10    // Connection timeout (seconds)
	maxBacklog         = 5     // Maximum connection backlog
)

// DecodeTcpPath decodes TCP path
// path format: [address]:[port][#port]
func DecodeTcpPath(path string, addr, port, user, passwd, mntpnt, str *string) {
	var (
		buff string
		p    string
	)

	Tracet(4, "DecodeTcpPath: path=%s\n", path)

	buff = path

	// Parse address and port
	if i := strings.Index(buff, "@"); i >= 0 {
		// Extract user and password
		p = buff[:i]
		buff = buff[i+1:]

		// Extract user and password
		if j := strings.Index(p, ":"); j >= 0 {
			if user != nil {
				*user = p[:j]
			}
			if passwd != nil {
				*passwd = p[j+1:]
			}
		} else {
			if user != nil {
				*user = p
			}
		}
	}

	// Extract mountpoint
	if i := strings.Index(buff, "/"); i >= 0 {
		if mntpnt != nil {
			*mntpnt = buff[i+1:]
		}
		buff = buff[:i]
	}

	// Extract address and port
	if i := strings.LastIndex(buff, ":"); i >= 0 {
		if addr != nil {
			*addr = buff[:i]
		}
		if port != nil {
			*port = buff[i+1:]
		}
	} else {
		if addr != nil {
			*addr = buff
		}
	}

	// Extract mountpoint string
	if str != nil && mntpnt != nil {
		*str = *mntpnt
		if i := strings.Index(*str, ":"); i >= 0 {
			*str = (*str)[:i]
		}
	}
}

// OpenTcpClient opens a TCP client
// path format: address:port
func OpenTcpClient(path string, msg *string) *TcpClient {
	var (
		tcpcli       *TcpClient = new(TcpClient)
		saddr, sport string
	)

	Tracet(3, "OpenTcpClient: path=%s\n", path)

	// Decode TCP path
	DecodeTcpPath(path, &saddr, &sport, nil, nil, nil, nil)

	// Set default values
	if len(saddr) == 0 {
		saddr = "localhost"
	}
	if len(sport) == 0 {
		sport = strconv.Itoa(defaultTcpPort)
	}

	// Initialize TCP client
	tcpcli.svr.state = 0
	tcpcli.svr.saddr = saddr
	tcpcli.svr.port, _ = strconv.Atoi(sport)
	tcpcli.svr.tcon = 0
	tcpcli.toinact = defaultTcpInactTO
	tcpcli.tirecon = defaultTcpReconTO

	// Connect to server
	if tcpcli.svr.GenTcp(1, msg) == 0 {
		return nil
	}

	return tcpcli
}

// CloseTcpClient closes a TCP client
func (tcpcli *TcpClient) CloseTcpClient() {
	Tracet(3, "CloseTcpClient:\n")

	if tcpcli == nil {
		return
	}

	// Close socket
	if tcpcli.svr.sock != nil {
		if conn, ok := tcpcli.svr.sock.(net.Conn); ok {
			conn.Close()
		}
		tcpcli.svr.sock = nil
	}
	tcpcli.svr.state = 0
}

// ReadTcpClient reads data from a TCP client
func (tcpcli *TcpClient) ReadTcpClient(buff []byte, n int, msg *string) int {
	var (
		nr   int
		err  error
		tick uint32
	)

	Tracet(4, "ReadTcpClient: n=%d\n", n)

	if tcpcli == nil {
		return 0
	}

	// Check connection status
	if tcpcli.svr.state == 0 {
		// Try to reconnect if not connected
		if tcpcli.tirecon > 0 &&
			(int64(TickGet())-tcpcli.svr.tdis) > int64(tcpcli.tirecon) {
			tcpcli.svr.tcon = 0
		}
		if tcpcli.svr.tcon == 0 {
			// Try to connect
			if tcpcli.svr.GenTcp(1, msg) == 0 {
				tcpcli.svr.tcon = 1000
				return 0
			}
		} else {
			// Wait for reconnect timeout
			tcpcli.svr.tcon -= 100
			if tcpcli.svr.tcon < 0 {
				tcpcli.svr.tcon = 0
			}
			return 0
		}
	}

	// Check socket
	if tcpcli.svr.sock == nil {
		return 0
	}

	// Get the connection
	conn, ok := tcpcli.svr.sock.(net.Conn)
	if !ok {
		if msg != nil {
			*msg = "tcp socket is not a valid connection"
		}
		tcpcli.svr.sock = nil
		tcpcli.svr.state = 0
		return 0
	}

	// Set read deadline
	conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))

	// Read data
	nr, err = conn.Read(buff[:n])
	if err != nil {
		// Handle connection error
		if msg != nil {
			*msg = fmt.Sprintf("tcp read error: %s", err.Error())
		}
		conn.Close()
		tcpcli.svr.sock = nil
		tcpcli.svr.state = 0
		tcpcli.svr.tcon = tcpcli.tirecon
		tcpcli.svr.tdis = int64(TickGet())
		return 0
	}

	// Update activity time
	if nr > 0 {
		tcpcli.svr.tact = int64(TickGet())
	} else {
		// Check for inactive timeout
		tick = TickGet()
		if tcpcli.toinact > 0 &&
			(int64(tick)-tcpcli.svr.tact) > int64(tcpcli.toinact) {
			// Connection inactive for too long, close it
			if msg != nil {
				*msg = "tcp timeout"
			}
			if conn, ok := tcpcli.svr.sock.(net.Conn); ok {
				conn.Close()
			}
			tcpcli.svr.sock = nil
			tcpcli.svr.state = 0
			tcpcli.svr.tcon = tcpcli.tirecon
			tcpcli.svr.tdis = int64(tick)
			return 0
		}
	}

	return nr
}

// WriteTcpClient writes data to a TCP client
func (tcpcli *TcpClient) WriteTcpClient(buff []byte, n int, msg *string) int {
	var (
		ns  int
		err error
	)

	Tracet(4, "WriteTcpClient: n=%d\n", n)

	if tcpcli == nil {
		return 0
	}

	// Check connection status
	if tcpcli.svr.state == 0 {
		// Try to reconnect if not connected
		if tcpcli.tirecon > 0 &&
			(int64(TickGet())-tcpcli.svr.tdis) > int64(tcpcli.tirecon) {
			tcpcli.svr.tcon = 0
		}
		if tcpcli.svr.tcon == 0 {
			// Try to connect
			if tcpcli.svr.GenTcp(1, msg) == 0 {
				tcpcli.svr.tcon = 1000
				return 0
			}
		} else {
			// Wait for reconnect timeout
			tcpcli.svr.tcon -= 100
			if tcpcli.svr.tcon < 0 {
				tcpcli.svr.tcon = 0
			}
			return 0
		}
	}

	// Check socket
	if tcpcli.svr.sock == nil {
		return 0
	}

	// Get the connection
	conn, ok := tcpcli.svr.sock.(net.Conn)
	if !ok {
		if msg != nil {
			*msg = "tcp socket is not a valid connection"
		}
		tcpcli.svr.sock = nil
		tcpcli.svr.state = 0
		return 0
	}

	// Set write deadline
	conn.SetWriteDeadline(time.Now().Add(1 * time.Second))

	// Write data
	ns, err = conn.Write(buff[:n])
	if err != nil {
		// Handle connection error
		if msg != nil {
			*msg = fmt.Sprintf("tcp write error: %s", err.Error())
		}
		conn.Close()
		tcpcli.svr.sock = nil
		tcpcli.svr.state = 0
		tcpcli.svr.tcon = tcpcli.tirecon
		tcpcli.svr.tdis = int64(TickGet())
		return 0
	}

	// Update activity time
	if ns > 0 {
		tcpcli.svr.tact = int64(TickGet())
	}

	return ns
}

// StateXTcpClient returns the state of a TCP client
func (tcpcli *TcpClient) StateXTcpClient(msg *string) int {
	return tcpcli.svr.state
}

// OpenTcpSvr opens a TCP server
// path format: :port
func OpenTcpSvr(path string, msg *string) *TcpSvr {
	var (
		tcpsvr *TcpSvr = new(TcpSvr)
		sport  string
		port   int
	)

	Tracet(3, "OpenTcpSvr: path=%s\n", path)

	// Decode TCP path
	DecodeTcpPath(path, nil, &sport, nil, nil, nil, nil)

	// Parse port
	if len(sport) == 0 {
		port = defaultTcpPort
	} else {
		port, _ = strconv.Atoi(sport)
	}

	// Initialize TCP server
	tcpsvr.svr.state = 0
	tcpsvr.svr.port = port
	tcpsvr.svr.saddr = ""
	tcpsvr.svr.tcon = 0

	// Initialize client connections
	for i := 0; i < MAXCLI; i++ {
		tcpsvr.cli[i].state = 0
		tcpsvr.cli[i].sock = nil
	}

	// Create server socket
	addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		if msg != nil {
			*msg = fmt.Sprintf("tcp address error: %s", err.Error())
		}
		return nil
	}

	// Create listener
	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		if msg != nil {
			*msg = fmt.Sprintf("tcp listen error: %s", err.Error())
		}
		return nil
	}

	// Store the listener
	tcpsvr.svr.sock = listener
	tcpsvr.svr.state = 1

	return tcpsvr
}

// CloseTcpSvr closes a TCP server
func (tcpsvr *TcpSvr) CloseTcpSvr() {
	Tracet(3, "CloseTcpSvr:\n")

	if tcpsvr == nil {
		return
	}

	// Close client connections
	for i := 0; i < MAXCLI; i++ {
		if tcpsvr.cli[i].state > 0 && tcpsvr.cli[i].sock != nil {
			if conn, ok := tcpsvr.cli[i].sock.(net.Conn); ok {
				conn.Close()
			}
			tcpsvr.cli[i].sock = nil
			tcpsvr.cli[i].state = 0
		}
	}

	// Close server socket
	if tcpsvr.svr.sock != nil {
		if listener, ok := tcpsvr.svr.sock.(*net.TCPListener); ok {
			listener.Close()
		}
		tcpsvr.svr.sock = nil
	}
	tcpsvr.svr.state = 0
}

// Accept_nb accepts a non-blocking connection
func Accept_nb(listener *net.TCPListener) net.Conn {
	// Set a short deadline to make the accept non-blocking
	listener.SetDeadline(time.Now().Add(10 * time.Millisecond))

	// Try to accept a connection
	conn, err := listener.Accept()
	if err != nil {
		// Timeout or other error
		return nil
	}

	return conn
}

// ReadTcpSvr reads data from a TCP server
func (tcpsvr *TcpSvr) ReadTcpSvr(buff []byte, n int, msg *string) int {
	var (
		nr  int
		err error
		i   int
	)

	Tracet(4, "ReadTcpSvr: n=%d\n", n)

	if tcpsvr == nil {
		return 0
	}

	// Accept new client connections
	if tcpsvr.svr.state > 0 {
		// Find free client slot
		for i = 0; i < MAXCLI; i++ {
			if tcpsvr.cli[i].state == 0 {
				break
			}
		}
		if i < MAXCLI {
			// Accept connection
			if listener, ok := tcpsvr.svr.sock.(*net.TCPListener); ok {
				conn := Accept_nb(listener)
				if conn != nil {
					// Connection accepted
					tcpsvr.cli[i].sock = conn
					tcpsvr.cli[i].state = 1
					tcpsvr.cli[i].tact = int64(TickGet())
				}
			}
		}
	}

	// Read data from clients
	for i = 0; i < MAXCLI; i++ {
		if tcpsvr.cli[i].state == 0 || tcpsvr.cli[i].sock == nil {
			continue
		}

		// Get the connection
		conn, ok := tcpsvr.cli[i].sock.(net.Conn)
		if !ok {
			tcpsvr.cli[i].sock = nil
			tcpsvr.cli[i].state = 0
			continue
		}

		// Set read deadline
		conn.SetReadDeadline(time.Now().Add(10 * time.Millisecond))

		// Read data
		nr, err = conn.Read(buff[:n])
		if err != nil {
			// Handle connection error
			conn.Close()
			tcpsvr.cli[i].sock = nil
			tcpsvr.cli[i].state = 0
			continue
		}

		// Update activity time
		if nr > 0 {
			tcpsvr.cli[i].tact = int64(TickGet())
			return nr
		}
	}

	return 0
}

// WriteTcpSvr writes data to a TCP server
func (tcpsvr *TcpSvr) WriteTcpSvr(buff []byte, n int, msg *string) int {
	var (
		i, ns int
		err   error
	)

	Tracet(4, "WriteTcpSvr: n=%d\n", n)

	if tcpsvr == nil {
		return 0
	}

	// Write data to all clients
	for i = 0; i < MAXCLI; i++ {
		if tcpsvr.cli[i].state == 0 || tcpsvr.cli[i].sock == nil {
			continue
		}

		// Get the connection
		conn, ok := tcpsvr.cli[i].sock.(net.Conn)
		if !ok {
			tcpsvr.cli[i].sock = nil
			tcpsvr.cli[i].state = 0
			continue
		}

		// Set write deadline
		conn.SetWriteDeadline(time.Now().Add(1 * time.Second))

		// Write data
		ns, err = conn.Write(buff[:n])
		if err != nil {
			// Handle connection error
			conn.Close()
			tcpsvr.cli[i].sock = nil
			tcpsvr.cli[i].state = 0
			continue
		}

		// Update activity time
		if ns > 0 {
			tcpsvr.cli[i].tact = int64(TickGet())
		}
	}

	return n
}

// StateXTcpSvr returns the state of a TCP server
func (tcpsvr *TcpSvr) StateXTcpSvr(msg *string) int {
	var (
		state int
		i     int
	)

	if tcpsvr == nil {
		return 0
	}

	// Count active client connections
	state = tcpsvr.svr.state
	for i = 0; i < MAXCLI; i++ {
		if tcpsvr.cli[i].state > 0 {
			state++
		}
	}
	return state
}

// ResolveAddr resolves a TCP address
func (tcp *TcpConn) ResolveAddr() string {
	if tcp.port <= 0 {
		return tcp.saddr
	}
	return fmt.Sprintf("%s:%d", tcp.saddr, tcp.port)
}

// GenTcp generates a TCP socket
func (tcp *TcpConn) GenTcp(ctype int, msg *string) int {
	if ctype == 0 { // Server socket
		// Server sockets are created in OpenTcpSvr
		tcp.state = 1
	} else { // Client socket
		var err error

		// Resolve address
		addr := tcp.ResolveAddr()

		// Connect to server with timeout
		dialer := net.Dialer{Timeout: defaultConnTimeout * time.Second}
		tcp.sock, err = dialer.Dial("tcp", addr)
		if err != nil {
			*msg = fmt.Sprintf("connect error: %s", err)
			tcp.state = -1
			return 0
		}
		tcp.state = 1
	}

	tcp.tact = int64(TickGet())
	Tracet(5, "GenTcp: exit sock=%v\n", tcp.sock)

	return 1
}
