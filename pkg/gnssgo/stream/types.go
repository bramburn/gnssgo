// Package stream provides stream input/output functionality for GNSS data
package stream

import (
	"net"
	"os"
	"sync"
	"time"

	"github.com/bramburn/gnssgo/pkg/gnssgo/gtime"
	"go.bug.st/serial"
)

// Stream types
const (
	STR_NONE     = 0  // No stream
	STR_SERIAL   = 1  // Serial
	STR_FILE     = 2  // File
	STR_TCPSVR   = 3  // TCP server
	STR_TCPCLI   = 4  // TCP client
	STR_NTRIPSVR = 5  // NTRIP server
	STR_NTRIPCLI = 6  // NTRIP client
	STR_NTRIPCAS = 7  // NTRIP caster
	STR_UDPSVR   = 8  // UDP server
	STR_UDPCLI   = 9  // UDP client
	STR_MEMBUF   = 10 // Memory buffer
	STR_FTP      = 11 // FTP
	STR_HTTP     = 12 // HTTP
)

// Stream modes
const (
	STR_MODE_R  = 0x1 // Read
	STR_MODE_W  = 0x2 // Write
	STR_MODE_RW = 0x3 // Read/Write
)

// Stream status
const (
	STR_STAT_NONE   = 0 // No stream
	STR_STAT_WAIT   = 1 // Waiting for connection
	STR_STAT_CONN   = 2 // Connected
	STR_STAT_ACTIVE = 3 // Active
)

// Stream constants
const (
	TINTACT             = 200                      // Period for stream active (ms)
	SERIBUFFSIZE        = 4096                     // Serial buffer size (bytes)
	TIMETAGH_LEN        = 64                       // Time tag file header length
	MAXCLI              = 32                       // Max client connection for tcp svr
	MAXSTATMSG          = 32                       // Max length of status message
	DEFAULT_MEMBUF_SIZE = 4096                     // Default memory buffer size (bytes)
	NTRIP_AGENT         = "RTKLIB/3.0.0"           // Version hardcoded for now
	NTRIP_CLI_PORT      = 2101                     // Default ntrip-client connection port
	NTRIP_SVR_PORT      = 80                       // Default ntrip-server connection port
	NTRIP_MAXRSP        = 32768                    // Max size of ntrip response
	NTRIP_MAXSTR        = 256                      // Max length of mountpoint string
	NTRIP_RSP_OK_CLI    = "ICY 200 OK\r\n"         // Ntrip response: client
	NTRIP_RSP_OK_SVR    = "OK\r\n"                 // Ntrip response: server
	NTRIP_RSP_SRCTBL    = "SOURCETABLE 200 OK\r\n" // Ntrip response: source table
	NTRIP_RSP_TBLEND    = "ENDSOURCETABLE"
	NTRIP_RSP_HTTP      = "HTTP/" // Ntrip response: http
	NTRIP_RSP_ERROR     = "ERROR" // Ntrip response: error
	NTRIP_RSP_UNAUTH    = "HTTP/1.0 401 Unauthorized\r\n"
	NTRIP_RSP_ERR_PWD   = "ERROR - Bad Pasword\r\n"
	NTRIP_RSP_ERR_MNTP  = "ERROR - Bad Mountpoint\r\n"
)

// Global options
var (
	toinact     int    = 10000 // Inactive timeout (ms)
	ticonnect   int    = 1000  // Interval to re-connect (ms)
	tirate      int    = 1000  // Averaging time for data rate (ms)
	localdir    string = ""    // Local directory for ftp/http
	proxyaddr   string = ""    // Http/ntrip/ftp proxy address
	tick_master uint32 = 0     // Time tick master for replay
	fswapmargin int    = 30    // File swap margin (s)
)

// Stream represents a generic stream
type Stream struct {
	Type        int        // Stream type
	Mode        int        // Stream mode
	State       int        // Stream state
	InBytes     uint32     // Bytes of input data
	InRate      uint32     // Input rate (bytes/sec)
	OutBytes    uint32     // Bytes of output data
	OutRate     uint32     // Output rate (bytes/sec)
	TickInput   uint32     // Tick of input
	TickOutput  uint32     // Tick of output
	TickActive  uint32     // Tick of active
	InByeTick   uint32     // Input bytes at tick
	OutByteTick uint32     // Output bytes at tick
	Path        string     // Stream path
	Msg         string     // Stream message
	Port        any        // Stream port
	Lock        sync.Mutex // Lock for thread safety
}

// FileType represents a file stream
type FileType struct {
	fp         *os.File    // File pointer
	fp_tag     *os.File    // File pointer of tag file
	fp_tmp     *os.File    // Temporary file pointer for swap
	fp_tag_tmp *os.File    // Temporary file pointer of tag file for swap
	path       string      // File path
	openpath   string      // Open file path
	mode       int         // File mode
	timetag    int         // Time tag flag (0:off,1:on)
	repmode    int         // Replay mode (0:master,1:slave)
	offset     int         // Time offset (ms) for slave
	size_fpos  int         // File position size (bytes)
	time       gtime.Gtime // Start time
	wtime      gtime.Gtime // Write time
	tick       uint32      // Start tick
	tick_f     uint32      // Start tick in file
	fpos_n     int64       // Next file position
	tick_n     uint32      // Next tick
	start      float64     // Start offset (s)
	speed      float64     // Replay speed (time factor)
	swapintv   float64     // Swap interval (hr) (0: no swap)
	lock       sync.Mutex  // Lock flag
}

// TcpConn represents a TCP connection
type TcpConn struct {
	state int         // State (0:close,1:wait,2:connect)
	saddr string      // Address string
	port  int         // Port
	addr  net.Addr    // Address resolved
	sock  interface{} // Socket descriptor (net.Conn or *net.TCPListener)
	tcon  int         // Reconnect time (ms) (-1:never,0:now)
	tact  int64       // Data active tick
	tdis  int64       // Disconnect tick
}

// TcpSvr represents a TCP server
type TcpSvr struct {
	svr TcpConn         // TCP server control
	cli [MAXCLI]TcpConn // TCP client controls
}

// TcpClient represents a TCP client
type TcpClient struct {
	svr     TcpConn // TCP server control
	toinact int     // Inactive timeout (ms) (0:no timeout)
	tirecon int     // Reconnect interval (ms) (0:no reconnect)
}

// SerialComm represents a serial connection
type SerialComm struct {
	dev      int           // Serial device
	serialio serial.Port   // Serial port interface
	err      int           // Error state
	lock     sync.Mutex    // Lock flag for thread safety
	tcpsvr   *TcpSvr       // TCP server for received stream
	mode     *serial.Mode  // Serial port mode
	timeout  time.Duration // Read timeout
}

// NTrip represents an NTRIP connection
type NTrip struct {
	state  int        // State (0:close,1:wait,2:connect)
	ctype  int        // Type (0:server,1:client)
	nb     int        // Response buffer size
	url    string     // URL for proxy
	mntpnt string     // Mountpoint
	user   string     // User
	passwd string     // Password
	str    string     // Mountpoint string for server
	buff   string     // Response buffer
	tcp    *TcpClient // TCP client
}

// NTripc_con represents an NTRIP client/server connection
type NTripc_con struct {
	state  int    // State (0:close,1:connect)
	mntpnt string // Mountpoint
	nb     int    // Request buffer size
	buff   string // Request buffer
}

// NTripc represents an NTRIP caster
type NTripc struct {
	state  int          // State (0:close,1:wait,2:connect)
	ctype  int          // Type (0:server,1:client)
	mntpnt string       // Mountpoint
	user   string       // User
	passwd string       // Password
	srctbl string       // Source table
	tcp    *TcpSvr      // TCP server
	con    []NTripc_con // NTRIP client/server connections
}

// UdpConn represents a UDP connection
type UdpConn struct {
	state int      // State (0:close,1:open)
	ctype int      // Type (0:server,1:client)
	port  int      // Port
	saddr string   // Address (server:filter,client:server)
	sock  net.Conn // Socket descriptor
}

// FtpConn represents an FTP/HTTP connection
type FtpConn struct {
	state  int         // State (0:close,1:download,2:complete,3:error)
	proto  int         // Protocol (0:ftp,1:http)
	error  int         // Error code
	addr   string      // Download address
	file   string      // Download file path
	user   string      // User for ftp
	passwd string      // Password for ftp
	local  string      // Local file path
	topts  [4]int      // Time options {poff,tint,toff,tretry} (s)
	tnext  gtime.Gtime // Next retry time (gpst)
	thread int         // Download thread
}

// MemBuf represents a memory buffer
type MemBuf struct {
	state, wp, rp int        // State,write/read pointer
	bufsize       int        // Buffer size (bytes)
	lock          sync.Mutex // Lock flag
	buf           []byte     // Write buffer
}
