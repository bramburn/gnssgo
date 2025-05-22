package stream

import (
	"github.com/bramburn/gnssgo/pkg/gnssgo/gtime"
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

// StreamGetTime gets stream time
func StreamGetTime(stream *Stream) gtime.Gtime {
	var time gtime.Gtime

	if stream == nil || stream.Port == nil {
		return time
	}

	stream.StreamLock()
	defer stream.StreamUnlock()

	switch byte(stream.Type) {
	case STR_FILE:
		time = stream.Port.(*FileType).time
	default:
		time = gtime.Utc2GpsT(gtime.TimeGet())
	}

	return time
}

// strinitcom initializes stream communication
func strinitcom() {
	// This is a placeholder function that would initialize any global
	// communication resources. In the original RTKLIB, this would initialize
	// the Windows socket library.
	Tracet(3, "strinitcom:\n")
}

// strsync synchronizes streams
func strsync(stream1, stream2 *Stream) {
	// This is a placeholder function that would synchronize two streams.
	// In the original RTKLIB, this would synchronize file streams.
	Tracet(3, "strsync:\n")

	if stream1 == nil || stream2 == nil {
		return
	}

	// If both streams are file streams, synchronize them
	if stream1.Type == STR_FILE && stream2.Type == STR_FILE {
		file1 := stream1.Port.(*FileType)
		file2 := stream2.Port.(*FileType)

		// Synchronize file positions
		if file1.time.Time != 0 && file2.time.Time != 0 {
			if gtime.TimeDiff(file1.time, file2.time) > 0.0 {
				// If file1 is ahead of file2, rewind file1
				file1.repmode = 1 // slave mode
				file1.offset = 0
			} else {
				// If file2 is ahead of file1, rewind file2
				file2.repmode = 1 // slave mode
				file2.offset = 0
			}
		}
	}
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
func (ntrip *NTrip) CloseNtrip() {
	// Get the enhanced implementation from the registry
	if enhancedNtrip := GetEnhancedNTripFromRegistry(ntrip); enhancedNtrip != nil {
		enhancedNtrip.Close()
	} else if ntrip.tcp != nil {
		ntrip.tcp.CloseTcpClient()
	}
}

func (ntrip *NTrip) ReadNtrip(buff []byte, size int, msg *string) int {
	// Get the enhanced implementation from the registry
	if enhancedNtrip := GetEnhancedNTripFromRegistry(ntrip); enhancedNtrip != nil {
		return enhancedNtrip.ReadNtrip(buff, size, msg)
	} else if ntrip.tcp != nil {
		// Fall back to the legacy implementation
		if ntrip.nb > 0 { // read response buffer first
			nb := size
			if ntrip.nb <= size {
				nb = ntrip.nb
			}
			copy(buff, []byte(ntrip.buff)[ntrip.nb-nb:ntrip.nb])
			ntrip.nb = 0
			ntrip.buff = ""
			return nb
		}
		return ntrip.tcp.ReadTcpClient(buff, size, msg)
	}
	return 0
}

func (ntrip *NTrip) WriteNtrip(buff []byte, size int, msg *string) int {
	// Get the enhanced implementation from the registry
	if enhancedNtrip := GetEnhancedNTripFromRegistry(ntrip); enhancedNtrip != nil {
		return enhancedNtrip.WriteNtrip(buff, size, msg)
	} else if ntrip.tcp != nil {
		// Fall back to the legacy implementation
		return ntrip.tcp.WriteTcpClient(buff, size, msg)
	}
	return 0
}

func (ntrip *NTrip) StatExNtrip(msg *string) int {
	// Get the enhanced implementation from the registry
	if enhancedNtrip := GetEnhancedNTripFromRegistry(ntrip); enhancedNtrip != nil {
		return enhancedNtrip.GetState()
	}
	return ntrip.state
}

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
