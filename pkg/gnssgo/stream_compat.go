// Package gnssgo provides GNSS-related functionality
package gnssgo

import (
	"github.com/bramburn/gnssgo/pkg/gnssgo/stream"
)

// DEPRECATED: This file is maintained for backward compatibility only.
// Please use the pkg/gnssgo/stream package for new code.

// Stream type constants
const (
	STR_NONE     = stream.STR_NONE     // No stream
	STR_SERIAL   = stream.STR_SERIAL   // Serial
	STR_FILE     = stream.STR_FILE     // File
	STR_TCPSVR   = stream.STR_TCPSVR   // TCP server
	STR_TCPCLI   = stream.STR_TCPCLI   // TCP client
	STR_NTRIPSVR = stream.STR_NTRIPSVR // NTRIP server
	STR_NTRIPCLI = stream.STR_NTRIPCLI // NTRIP client
	STR_NTRIPCAS = stream.STR_NTRIPCAS // NTRIP caster
	STR_UDPSVR   = stream.STR_UDPSVR   // UDP server
	STR_UDPCLI   = stream.STR_UDPCLI   // UDP client
	STR_MEMBUF   = stream.STR_MEMBUF   // Memory buffer
	STR_FTP      = stream.STR_FTP      // FTP
	STR_HTTP     = stream.STR_HTTP     // HTTP
)

// Stream mode constants
const (
	STR_MODE_R  = stream.STR_MODE_R  // Read
	STR_MODE_W  = stream.STR_MODE_W  // Write
	STR_MODE_RW = stream.STR_MODE_RW // Read/Write
)

// Stream status constants
const (
	STR_STAT_NONE   = stream.STR_STAT_NONE   // No stream
	STR_STAT_WAIT   = stream.STR_STAT_WAIT   // Waiting for connection
	STR_STAT_CONN   = stream.STR_STAT_CONN   // Connected
	STR_STAT_ACTIVE = stream.STR_STAT_ACTIVE // Active
)

// Stream represents a generic stream (compatibility wrapper)
type Stream = stream.Stream

// FileType represents a file stream (compatibility wrapper)
type FileType = stream.FileType

// TcpConn represents a TCP connection (compatibility wrapper)
type TcpConn = stream.TcpConn

// TcpSvr represents a TCP server (compatibility wrapper)
type TcpSvr = stream.TcpSvr

// TcpClient represents a TCP client (compatibility wrapper)
type TcpClient = stream.TcpClient

// SerialComm represents a serial connection (compatibility wrapper)
type SerialComm = stream.SerialComm

// NTrip represents an NTRIP connection (compatibility wrapper)
type NTrip = stream.NTrip

// NTripc_con represents an NTRIP client/server connection (compatibility wrapper)
type NTripc_con = stream.NTripc_con

// NTripc represents an NTRIP caster (compatibility wrapper)
type NTripc = stream.NTripc

// UdpConn represents a UDP connection (compatibility wrapper)
type UdpConn = stream.UdpConn

// FtpConn represents an FTP/HTTP connection (compatibility wrapper)
type FtpConn = stream.FtpConn

// MemBuf represents a memory buffer (compatibility wrapper)
type MemBuf = stream.MemBuf

// ListSerialPorts lists available serial ports (compatibility wrapper)
func ListSerialPorts() ([]string, error) {
	return stream.ListSerialPorts()
}

// OpenSerial opens a serial port (compatibility wrapper)
func OpenSerial(path string, modeFlag int, msg *string) *SerialComm {
	return stream.OpenSerial(path, modeFlag, msg)
}

// OpenStreamFile opens a file stream (compatibility wrapper)
func OpenStreamFile(path string, mode int, msg *string) *FileType {
	return stream.OpenStreamFile(path, mode, msg)
}

// OpenTcpSvr opens a TCP server (compatibility wrapper)
func OpenTcpSvr(path string, msg *string) *TcpSvr {
	return stream.OpenTcpSvr(path, msg)
}

// OpenTcpClient opens a TCP client (compatibility wrapper)
func OpenTcpClient(path string, msg *string) *TcpClient {
	return stream.OpenTcpClient(path, msg)
}

// OpenNtrip opens an NTRIP connection (compatibility wrapper)
func OpenNtrip(path string, ctype int, msg *string) *NTrip {
	return stream.OpenNtrip(path, ctype, msg)
}

// OpenNtripc opens an NTRIP caster (compatibility wrapper)
func OpenNtripc(path string, msg *string) *NTripc {
	return stream.OpenNtripc(path, msg)
}

// OpenUdpSvr opens a UDP server (compatibility wrapper)
func OpenUdpSvr(path string, msg *string) *UdpConn {
	return stream.OpenUdpSvr(path, msg)
}

// OpenUdpClient opens a UDP client (compatibility wrapper)
func OpenUdpClient(path string, msg *string) *UdpConn {
	return stream.OpenUdpClient(path, msg)
}

// OpenMemBuf opens a memory buffer (compatibility wrapper)
func OpenMemBuf(path string, msg *string) *MemBuf {
	return stream.OpenMemBuf(path, msg)
}

// OpenFtp opens an FTP/HTTP connection (compatibility wrapper)
func OpenFtp(path string, proto int, msg *string) *FtpConn {
	return stream.OpenFtp(path, proto, msg)
}

// SetBrate sets the baud rate for a serial connection (compatibility wrapper)
func SetBrate(str *Stream, brate int) {
	stream.SetBrate(str, brate)
}

// GenUbx generates a UBX message (compatibility wrapper)
func GenUbx(cmd string, buff []byte) int {
	return stream.GenUbx(cmd, buff)
}

// GenStq generates a STQ message (compatibility wrapper)
func GenStq(cmd string, buff []byte) int {
	return stream.GenStq(cmd, buff)
}

// GenNvs generates an NVS message (compatibility wrapper)
func GenNvs(cmd string, buff []byte) int {
	return stream.GenNvs(cmd, buff)
}

// GenHex generates a HEX message (compatibility wrapper)
func GenHex(cmd string, buff []byte) int {
	return stream.GenHex(cmd, buff)
}
