package stream

import (
	"fmt"
	"strings"

	"github.com/bramburn/gnssgo/pkg/gnssgo/util"
)

// StreamLock locks a stream
func (stream *Stream) StreamLock() {
	stream.Lock.Lock()
}

// StreamUnlock unlocks a stream
func (stream *Stream) StreamUnlock() {
	stream.Lock.Unlock()
}

// InitStream initializes a stream
func (stream *Stream) InitStream() {
	util.Tracet(3, "strinit:\n")

	stream.Type = 0
	stream.Mode = 0
	stream.State = 0
	stream.InBytes, stream.InRate, stream.OutBytes, stream.OutRate = 0, 0, 0, 0
	stream.TickInput, stream.TickOutput, stream.TickActive, stream.InByeTick, stream.OutByteTick = 0, 0, 0, 0, 0

	stream.Port = nil
	stream.Path = ""
	stream.Msg = ""
}

// OpenStream opens a stream
func (stream *Stream) OpenStream(ctype, mode int, path string) int {
	Tracet(3, "stropen: type=%d mode=%d path=%s\n", ctype, mode, path)

	stream.Type = ctype
	stream.Mode = mode
	stream.Path = path
	stream.InBytes, stream.InRate, stream.OutBytes, stream.OutRate = 0, 0, 0, 0
	stream.TickInput = TickGet()
	stream.TickOutput = stream.TickInput
	stream.InByeTick, stream.OutByteTick = 0, 0
	stream.Msg = ""
	stream.Port = nil

	switch byte(ctype) {
	case STR_SERIAL:
		stream.Port = OpenSerial(path, mode, &stream.Msg)

	case STR_FILE:
		stream.Port = OpenStreamFile(path, mode, &stream.Msg)

	case STR_TCPSVR:
		stream.Port = OpenTcpSvr(path, &stream.Msg)

	case STR_TCPCLI:
		stream.Port = OpenTcpClient(path, &stream.Msg)

	case STR_NTRIPSVR:
		stream.Port = OpenNtrip(path, 0, &stream.Msg)

	case STR_NTRIPCLI:
		stream.Port = OpenNtrip(path, 1, &stream.Msg)

	case STR_NTRIPCAS:
		stream.Port = OpenNtripc(path, &stream.Msg)

	case STR_UDPSVR:
		stream.Port = OpenUdpSvr(path, &stream.Msg)

	case STR_UDPCLI:
		stream.Port = OpenUdpClient(path, &stream.Msg)

	case STR_MEMBUF:
		stream.Port = OpenMemBuf(path, &stream.Msg)

	case STR_FTP:
		stream.Port = OpenFtp(path, 0, &stream.Msg)

	case STR_HTTP:
		stream.Port = OpenFtp(path, 1, &stream.Msg)

	default:
		stream.State = 0
		return 1
	}

	if stream.Port == nil {
		stream.State = -1
		return 0
	}

	stream.State = 1
	return 1
}

// StreamClose closes a stream
func (stream *Stream) StreamClose() {
	Tracet(3, "strclose: type=%d\n", stream.Type)

	if stream.Port == nil {
		return
	}

	switch byte(stream.Type) {
	case STR_SERIAL:
		stream.Port.(*SerialComm).CloseSerial()

	case STR_FILE:
		stream.Port.(*FileType).CloseFile()

	case STR_TCPSVR:
		stream.Port.(*TcpSvr).CloseTcpSvr()

	case STR_TCPCLI:
		stream.Port.(*TcpClient).CloseTcpClient()

	case STR_NTRIPSVR, STR_NTRIPCLI:
		stream.Port.(*NTrip).CloseNtrip()

	case STR_NTRIPCAS:
		stream.Port.(*NTripc).CloseNtripc()

	case STR_UDPSVR, STR_UDPCLI:
		stream.Port.(*UdpConn).CloseUdp()

	case STR_MEMBUF:
		stream.Port.(*MemBuf).CloseMemBuf()

	case STR_FTP, STR_HTTP:
		stream.Port.(*FtpConn).CloseFtp()
	}

	stream.Type = 0
	stream.Mode = 0
	stream.State = 0
	stream.InBytes = 0
	stream.InRate = 0
	stream.OutBytes = 0
	stream.OutRate = 0
	stream.Port = nil
	stream.Path = ""
	stream.Msg = ""
}

// StreamRead reads data from a stream
func (stream *Stream) StreamRead(buff []byte, n int) int {
	var (
		tick uint32
		nr   int
		msg  string
	)

	Tracet(4, "strread: n=%d\n", n)

	if stream.Port == nil {
		return 0
	}

	stream.StreamLock()

	switch byte(stream.Type) {
	case STR_SERIAL:
		nr = stream.Port.(*SerialComm).ReadSerial(buff, n, &msg)

	case STR_FILE:
		nr = stream.Port.(*FileType).ReadFile(buff, int64(n), &msg)

	case STR_TCPSVR:
		nr = stream.Port.(*TcpSvr).ReadTcpSvr(buff, n, &msg)

	case STR_TCPCLI:
		nr = stream.Port.(*TcpClient).ReadTcpClient(buff, n, &msg)

	case STR_NTRIPSVR, STR_NTRIPCLI:
		nr = stream.Port.(*NTrip).ReadNtrip(buff, n, &msg)

	case STR_NTRIPCAS:
		nr = stream.Port.(*NTripc).ReadNtripc(buff, n, &msg)

	case STR_UDPSVR:
		nr = stream.Port.(*UdpConn).ReadUdpSvr(buff, n, &msg)

	case STR_MEMBUF:
		nr = stream.Port.(*MemBuf).ReadMemBuf(buff, n, &msg)

	case STR_FTP:
		nr = stream.Port.(*FtpConn).ReadFtp(buff, n, &msg)

	case STR_HTTP:
		nr = stream.Port.(*FtpConn).ReadFtp(buff, n, &msg)

	default:
		stream.StreamUnlock()
		return 0
	}

	stream.Msg = msg

	if nr > 0 {
		stream.InBytes += uint32(nr)
		tick = TickGet()
		if tick-stream.TickInput >= uint32(tirate) {
			stream.InRate = (stream.InBytes - stream.InByeTick) * 1000 / uint32(tick-stream.TickInput)
			stream.TickInput = tick
			stream.InByeTick = stream.InBytes
		}
		stream.TickActive = tick
	}

	stream.StreamUnlock()
	return nr
}

// StreamWrite writes data to a stream
func (stream *Stream) StreamWrite(buff []byte, n int) int {
	var (
		tick uint32
		ns   int
		msg  string
	)

	Tracet(4, "strwrite: n=%d\n", n)

	if stream.Port == nil {
		return 0
	}

	stream.StreamLock()
	tick = TickGet()

	switch byte(stream.Type) {
	case STR_SERIAL:
		ns = stream.Port.(*SerialComm).WriteSerial(buff, n, &msg)

	case STR_FILE:
		ns = stream.Port.(*FileType).WriteFile(buff, n, &msg)

	case STR_TCPSVR:
		ns = stream.Port.(*TcpSvr).WriteTcpSvr(buff, n, &msg)

	case STR_TCPCLI:
		ns = stream.Port.(*TcpClient).WriteTcpClient(buff, n, &msg)

	case STR_NTRIPSVR, STR_NTRIPCLI:
		ns = stream.Port.(*NTrip).WriteNtrip(buff, n, &msg)

	case STR_NTRIPCAS:
		ns = stream.Port.(*NTripc).WriteNtripc(buff, n, &msg)

	case STR_UDPCLI:
		ns = stream.Port.(*UdpConn).WriteUdpClient(buff, n, &msg)

	case STR_MEMBUF:
		ns = stream.Port.(*MemBuf).WriteMemBuf(buff, n, &msg)

	case STR_FTP, STR_HTTP:
		stream.StreamUnlock()
		return 0

	default:
		stream.StreamUnlock()
		return 0
	}

	stream.Msg = msg

	if ns > 0 {
		stream.OutBytes += uint32(ns)
		if tick-stream.TickOutput >= uint32(tirate) {
			stream.OutRate = (stream.OutBytes - stream.OutByteTick) * 1000 / uint32(tick-stream.TickOutput)
			stream.TickOutput = tick
			stream.OutByteTick = stream.OutBytes
		}
		stream.TickActive = tick
	}

	stream.StreamUnlock()
	return ns
}

// StreamSendNmea sends NMEA GGA message to stream
func (stream *Stream) StreamSendNmea(sol interface{}) {
	var (
		buff string
		n    int
	)

	Tracet(3, "strsendnmea: sending NMEA GGA message\n")

	// This is a placeholder implementation
	// In a real implementation, we would call sol.OutSolNmeaGga(&buff)
	buff = "$GPGGA,123519,4807.038,N,01131.000,E,1,08,0.9,545.4,M,46.9,M,,*47\r\n"
	n = len(buff)
	stream.StreamWrite([]byte(buff), n)
}

// StreamSendCmd sends a command to a stream
func (stream *Stream) StreamSendCmd(cmd string) {
	var (
		buff         []byte = make([]byte, 1024)
		m, ms, brate int
		str          *Stream = stream
	)

	Tracet(3, "strsendcmd: cmd=%s\n", cmd)

	// For binary commands, convert to byte array
	if strings.HasPrefix(cmd, "!") {
		switch {
		case strings.HasPrefix(cmd[1:], "WAIT"):
			if n, _ := fmt.Sscanf(cmd[5:], "%d", &ms); n < 1 {
				ms = 100
			}
			if ms > 3000 {
				ms = 3000 // max 3 s
			}
			Sleepms(ms)

		case strings.HasPrefix(cmd[1:], "BRATE"):
			if n, _ := fmt.Sscanf(cmd[6:], "%d", &brate); n < 1 {
				brate = 9600
			}
			SetBrate(str, brate)
			Sleepms(500)

		case strings.HasPrefix(cmd[1:], "UBX"):
			if m = GenUbx(cmd[4:], buff); m > 0 {
				str.StreamWrite(buff, m)
			}

		case strings.HasPrefix(cmd[1:], "STQ"):
			if m = GenStq(cmd[4:], buff); m > 0 {
				str.StreamWrite(buff, m)
			}

		case strings.HasPrefix(cmd[1:], "NVS"):
			if m = GenNvs(cmd[4:], buff); m > 0 {
				str.StreamWrite(buff, m)
			}

		case strings.HasPrefix(cmd[1:], "HEX"):
			if m = GenHex(cmd[4:], buff); m > 0 {
				str.StreamWrite(buff, m)
			}
		}
		return
	}

	// For regular commands, send as string
	stream.StreamWrite([]byte(cmd), len(cmd))
}

// StreamGetState gets stream state
func (stream *Stream) StreamGetState() int {
	if stream.Port == nil {
		return STR_STAT_NONE
	}
	return stream.State
}

// StreamGetStatEx gets extended stream state
func (stream *Stream) StreamGetStatEx(msg *string) int {
	var state int

	if stream.Port == nil {
		return 0
	}

	stream.StreamLock()

	switch byte(stream.Type) {
	case STR_SERIAL:
		state = stream.Port.(*SerialComm).StateXSerial(msg)
		*msg = "serial:\n"
		*msg += fmt.Sprintf("  state   = %d\n", state)

	case STR_FILE:
		state = stream.Port.(*FileType).StateXFile(msg)
		*msg = "file:\n"
		*msg += fmt.Sprintf("  state   = %d\n", state)

	case STR_TCPSVR:
		state = stream.Port.(*TcpSvr).StateXTcpSvr(msg)

	case STR_TCPCLI:
		state = stream.Port.(*TcpClient).StateXTcpClient(msg)

	case STR_NTRIPSVR, STR_NTRIPCLI:
		state = stream.Port.(*NTrip).StatExNtrip(msg)

	case STR_NTRIPCAS:
		state = stream.Port.(*NTripc).StatExNtripc(msg)

	case STR_UDPSVR:
		state = stream.Port.(*UdpConn).StatExUdpSvr(msg)

	case STR_UDPCLI:
		state = stream.Port.(*UdpConn).StateXUdpClient(msg)

	case STR_MEMBUF:
		state = stream.Port.(*MemBuf).StateXMemBuf(msg)

	case STR_FTP:
		state = stream.Port.(*FtpConn).StateXFtp(msg)

	case STR_HTTP:
		state = stream.Port.(*FtpConn).StateXFtp(msg)

	default:
		*msg = ""
		stream.StreamUnlock()
		return 0
	}

	if state == 2 && int(TickGet()-stream.TickActive) <= TINTACT {
		state = 3
	}

	stream.StreamUnlock()
	return state
}
