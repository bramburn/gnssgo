# Enhanced NTRIP Client for GNSSGO

This package provides an enhanced NTRIP (Networked Transport of RTCM via Internet Protocol) client implementation for the GNSSGO library. The implementation includes improved error handling, connection retry logic, and RTCM message statistics.

## Features

### Enhanced Error Handling

- Detailed error classification (network errors, authentication errors, server errors)
- Meaningful error messages that include the specific failure point
- Context-aware error handling using Go's errors.Wrap pattern

### Connection Retry Logic

- Exponential backoff for connection retries
- Configurable retry parameters (max retries, retry timeout)
- Detailed status information about connection attempts

### RTCM Message Logging and Statistics

- Structured logging of RTCM message types received
- Statistics collection (message counts by type, data rates)
- Circular buffer to capture the last N messages before an error
- Debug mode for detailed message logging

### RTCM Message Type Support

- Support for RTCM 3.3 message types (1074-1077, 1084-1087, 1094-1097, 1124-1127)
- MSM7 message parsing for high-precision applications
- Support for RTCM 3.3 SSR messages (1057-1068) for PPP applications

## Usage

### Basic Usage

```go
// Create a configuration
config := stream.DefaultNTripConfig()
config.Server = "rtk2go.com"
config.Port = 2101
config.Mountpoint = "EXAMPLE"
config.Username = "user"
config.Password = "pass"
config.Debug = true

// Create the NTRIP client
ntrip := stream.NewEnhancedNTrip(config, 1) // 1 = client mode

// Connect to the server
err := ntrip.Connect()
if err != nil {
    log.Fatalf("Failed to connect: %v", err)
}

// Read data
buffer := make([]byte, 4096)
for {
    // Get message statistics
    stats := ntrip.GetMessageStats()
    for msgType, stat := range stats {
        fmt.Printf("Message type %d: count=%d, last=%s, bytes=%d\n",
            msgType, stat.Count, stat.LastReceived.Format(time.RFC3339), stat.TotalBytes)
    }

    // Get data rate
    fmt.Printf("Data rate: %.2f bytes/sec\n", ntrip.GetDataRate())

    // Sleep for a while
    time.Sleep(5 * time.Second)
}

// Close the connection
ntrip.CloseNtrip()
```

### Error Handling

The enhanced NTRIP client provides detailed error information:

```go
err := ntrip.Connect()
if err != nil {
    if errors.Is(err, stream.ErrNTRIPAuthFailed) {
        fmt.Println("Authentication failed. Check your username and password.")
    } else if errors.Is(err, stream.ErrNTRIPMountpointInvalid) {
        fmt.Println("Invalid mountpoint. Check the mountpoint name.")
    } else if errors.Is(err, stream.ErrNTRIPNetworkError) {
        fmt.Println("Network error. Check your internet connection.")
    } else if errors.Is(err, stream.ErrNTRIPServerError) {
        fmt.Println("Server error. The NTRIP server might be down.")
    } else {
        fmt.Printf("Unknown error: %v\n", err)
    }
}
```

## RTCM Message Types

The enhanced NTRIP client supports the following RTCM message types:

| Message Type | Description |
|--------------|-------------|
| 1001-1004    | GPS RTK Observables |
| 1005-1006    | Stationary RTK Reference Station ARP |
| 1007-1008    | Antenna Descriptor |
| 1009-1012    | GLONASS RTK Observables |
| 1019         | GPS Ephemerides |
| 1020         | GLONASS Ephemerides |
| 1033         | Receiver and Antenna Descriptors |
| 1057-1068    | SSR Messages |
| 1071-1077    | MSM1-MSM7 for GPS |
| 1081-1087    | MSM1-MSM7 for GLONASS |
| 1091-1097    | MSM1-MSM7 for Galileo |
| 1101-1107    | MSM1-MSM7 for SBAS |
| 1111-1117    | MSM1-MSM7 for QZSS |
| 1121-1127    | MSM1-MSM7 for BeiDou |

## License

This package is part of the GNSSGO library and is licensed under the same terms.
