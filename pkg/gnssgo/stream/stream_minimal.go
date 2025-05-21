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

// OpenSerial is implemented in serial.go

// OpenStreamFile is implemented in file.go

// OpenTcpSvr is implemented in tcp.go

// OpenTcpClient is implemented in tcp.go

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

// OpenUdpSvr is implemented in udp.go

// OpenUdpClient is implemented in udp.go

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

// SetBrate is implemented in serial.go

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

// SerialComm methods are implemented in serial.go

// FileType methods are implemented in file.go

// TcpSvr methods are implemented in tcp.go

// TcpClient methods are implemented in tcp.go

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

// UdpConn methods are implemented in udp.go

// MemBuf methods
func (m *MemBuf) CloseMemBuf()                                    {}
func (m *MemBuf) ReadMemBuf(buff []byte, n int, msg *string) int  { return 0 }
func (m *MemBuf) WriteMemBuf(buff []byte, n int, msg *string) int { return 0 }
func (m *MemBuf) StateXMemBuf(msg *string) int                    { return 0 }

// FtpConn methods
func (f *FtpConn) CloseFtp()                                   {}
func (f *FtpConn) ReadFtp(buff []byte, n int, msg *string) int { return 0 }
func (f *FtpConn) StateXFtp(msg *string) int                   { return 0 }
