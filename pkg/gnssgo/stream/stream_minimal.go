package stream

import (
	"github.com/bramburn/gnssgo/pkg/gnssgo/util"
)

// TickGet returns the current tick count in milliseconds
func TickGet() uint32 {
	return util.TickGet()
}

// Sleepms sleeps for the specified number of milliseconds
func Sleepms(ms int) {
	util.Sleepms(ms)
}

// Tracet prints a trace message
func Tracet(level int, format string, args ...interface{}) {
	util.Tracet(level, format, args...)
}

// OpenSerial opens a serial port
func OpenSerial(path string, modeFlag int, msg *string) *SerialComm {
	// Placeholder implementation
	return nil
}

// OpenStreamFile opens a file stream
func OpenStreamFile(path string, mode int, msg *string) *FileType {
	// Placeholder implementation
	return nil
}

// OpenTcpSvr opens a TCP server
func OpenTcpSvr(path string, msg *string) *TcpSvr {
	// Placeholder implementation
	return nil
}

// OpenTcpClient opens a TCP client
func OpenTcpClient(path string, msg *string) *TcpClient {
	// Placeholder implementation
	return nil
}

// OpenNtrip opens an NTRIP connection
func OpenNtrip(path string, ctype int, msg *string) *NTrip {
	// Use the enhanced implementation
	return OpenEnhancedNtrip(path, ctype, msg)
}

// OpenNtripc opens an NTRIP caster
func OpenNtripc(path string, msg *string) *NTripc {
	// Placeholder implementation
	return nil
}

// OpenUdpSvr opens a UDP server
func OpenUdpSvr(path string, msg *string) *UdpConn {
	// Placeholder implementation
	return nil
}

// OpenUdpClient opens a UDP client
func OpenUdpClient(path string, msg *string) *UdpConn {
	// Placeholder implementation
	return nil
}

// OpenMemBuf opens a memory buffer
func OpenMemBuf(path string, msg *string) *MemBuf {
	// Placeholder implementation
	return nil
}

// OpenFtp opens an FTP/HTTP connection
func OpenFtp(path string, proto int, msg *string) *FtpConn {
	// Placeholder implementation
	return nil
}

// SetBrate sets the baud rate for a serial connection
func SetBrate(str *Stream, brate int) {
	// Placeholder implementation
}

// GenUbx generates a UBX message
func GenUbx(cmd string, buff []byte) int {
	// Placeholder implementation
	return 0
}

// GenStq generates a STQ message
func GenStq(cmd string, buff []byte) int {
	// Placeholder implementation
	return 0
}

// GenNvs generates an NVS message
func GenNvs(cmd string, buff []byte) int {
	// Placeholder implementation
	return 0
}

// GenHex generates a HEX message
func GenHex(cmd string, buff []byte) int {
	// Placeholder implementation
	return 0
}

// SerialComm methods
func (s *SerialComm) CloseSerial()                                    {}
func (s *SerialComm) ReadSerial(buff []byte, n int, msg *string) int  { return 0 }
func (s *SerialComm) WriteSerial(buff []byte, n int, msg *string) int { return 0 }
func (s *SerialComm) StateSerial() int                                { return 0 }

// FileType methods
func (f *FileType) CloseFile()                                     {}
func (f *FileType) ReadFile(buff []byte, n int64, msg *string) int { return 0 }
func (f *FileType) WriteFile(buff []byte, n int, msg *string) int  { return 0 }
func (f *FileType) StateFile() int                                 { return 0 }

// TcpSvr methods
func (t *TcpSvr) CloseTcpSvr()                                    {}
func (t *TcpSvr) ReadTcpSvr(buff []byte, n int, msg *string) int  { return 0 }
func (t *TcpSvr) WriteTcpSvr(buff []byte, n int, msg *string) int { return 0 }
func (t *TcpSvr) StatExTcpSvr(msg *string) int                    { return 0 }

// TcpClient methods
func (t *TcpClient) CloseTcpClient()                                    {}
func (t *TcpClient) ReadTcpClient(buff []byte, n int, msg *string) int  { return 0 }
func (t *TcpClient) WriteTcpClient(buff []byte, n int, msg *string) int { return 0 }
func (t *TcpClient) StatExTcpClient(msg *string) int                    { return 0 }

// NTrip methods
func (ntrip *NTrip) CloseNtrip()                                       {}
func (ntrip *NTrip) ReadNtrip(buff []byte, size int, msg *string) int  { return 0 }
func (ntrip *NTrip) WriteNtrip(buff []byte, size int, msg *string) int { return 0 }
func (ntrip *NTrip) StatExNtrip(msg *string) int                       { return 0 }

// NTripc methods
func (ntripc *NTripc) CloseNtripc()                                       {}
func (ntripc *NTripc) ReadNtripc(buff []byte, size int, msg *string) int  { return 0 }
func (ntripc *NTripc) WriteNtripc(buff []byte, size int, msg *string) int { return 0 }
func (ntripc *NTripc) StatExNtripc(msg *string) int                       { return 0 }

// UdpConn methods
func (u *UdpConn) CloseUdp()                                          {}
func (u *UdpConn) ReadUdpSvr(buff []byte, n int, msg *string) int     { return 0 }
func (u *UdpConn) WriteUdpClient(buff []byte, n int, msg *string) int { return 0 }
func (u *UdpConn) StatExUdpSvr(msg *string) int                       { return 0 }
func (u *UdpConn) StateXUdpClient(msg *string) int                    { return 0 }

// MemBuf methods
func (m *MemBuf) CloseMemBuf()                                    {}
func (m *MemBuf) ReadMemBuf(buff []byte, n int, msg *string) int  { return 0 }
func (m *MemBuf) WriteMemBuf(buff []byte, n int, msg *string) int { return 0 }
func (m *MemBuf) StateXMemBuf(msg *string) int                    { return 0 }

// FtpConn methods
func (f *FtpConn) CloseFtp()                                   {}
func (f *FtpConn) ReadFtp(buff []byte, n int, msg *string) int { return 0 }
func (f *FtpConn) StateXFtp(msg *string) int                   { return 0 }
