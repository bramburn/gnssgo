# TOP708 Reader CLI

A command-line tool for reading data from TOPGNSS TOP708 GNSS receivers with RTK correction support.

## Overview

This tool allows you to connect to a TOPGNSS TOP708 GNSS receiver through a serial port and monitor the data streams in various formats. It supports:

- Raw data monitoring
- NMEA sentence parsing and display
- RTCM message monitoring
- UBX protocol message monitoring
- RTK correction using NTRIP

## Installation

```bash
# From the repository root
cd cmd/top708reader
go build
```

This will create a `top708reader.exe` executable in the current directory.

## Usage

### Basic Usage

```bash
# List available serial ports
top708reader -list

# Connect to a specific port and monitor raw data
top708reader -port COM3 -baud 38400

# Monitor and parse NMEA sentences
top708reader -port COM3 -mode nmea

# Monitor RTCM messages
top708reader -port COM3 -mode rtcm

# Monitor UBX protocol messages
top708reader -port COM3 -mode ubx

# Specify a different baud rate
top708reader -port COM3 -baud 115200 -mode nmea
```

### RTK Correction

The tool supports RTK correction using NTRIP. To use RTK correction, you need to provide NTRIP server details:

```bash
# Enable RTK correction
top708reader -port COM3 -mode rtk -ntrip-server example.com -ntrip-port 2101 -ntrip-user username -ntrip-password password -ntrip-mount MOUNTPOINT

# Show RTK status updates
top708reader -port COM3 -mode rtk -ntrip-server example.com -ntrip-mount MOUNTPOINT -show-rtk-status
```

## Command-Line Options

| Option | Description | Default |
|--------|-------------|---------|
| `-port` | Serial port name (e.g., COM1, /dev/ttyUSB0) | (prompt) |
| `-baud` | Baud rate | 38400 |
| `-timeout` | Connection verification timeout | 5s |
| `-mode` | Data mode: raw, nmea, rtcm, ubx, rtk | raw |
| `-list` | List available ports and exit | false |
| `-rtk` | Enable RTK correction | false |
| `-ntrip-server` | NTRIP server address | (none) |
| `-ntrip-port` | NTRIP server port | 2101 |
| `-ntrip-user` | NTRIP username | (none) |
| `-ntrip-password` | NTRIP password | (none) |
| `-ntrip-mount` | NTRIP mountpoint | (none) |
| `-show-rtk-status` | Show RTK status updates | false |

## Data Modes

### Raw Mode

Displays the raw data stream from the device without any parsing.

### NMEA Mode

Parses and displays NMEA sentences. For GGA sentences, it also displays position information in a more readable format.

Example output:
```
[GGA] $GPGGA,123519,4807.038,N,01131.000,E,1,08,0.9,545.4,M,46.9,M,,*47
  Position: 4807.038N, 01131.000E
  Quality: 1, Satellites: 08, HDOP: 0.9
  Altitude: 545.4 M
```

### RTCM Mode

Monitors RTCM3.3 messages by looking for the RTCM preamble (0xD3) and displays raw data in hexadecimal format.

Example output:
```
Potential RTCM data detected at offset 5
Raw data (128 bytes): D3 00 01 3C 20 C0 ...
```

### UBX Mode

Monitors UBX protocol messages by looking for the UBX header (0xB5 0x62) and displays the message class, ID, length, and payload in hexadecimal format.

Example output:
```
UBX Message - Class: 0x01, ID: 0x07, Length: 92 bytes
  Payload: 10 32 00 00 00 00 00 00 ...
```

### RTK Mode

Combines NMEA data from the GNSS receiver with RTCM correction data from an NTRIP server to achieve high-precision positioning. Displays position information along with RTK status.

Example output:
```
[GGA] $GPGGA,123519,4807.038,N,01131.000,E,4,12,0.6,545.4,M,46.9,M,,*47
  Position: 4807.038N, 01131.000E
  Quality: 4 (RTK Fixed), Satellites: 12, HDOP: 0.6
  Altitude: 545.4 M
[RTK Status] RTK Fixed, Satellites: 12, HDOP: 0.6, Last Update: 15:35:19
```

## RTK Status Information

When using RTK mode with `-show-rtk-status`, the tool will display the current RTK status:

| Quality | Description |
|---------|-------------|
| 0 | No Fix |
| 1 | GPS Fix (no RTK) |
| 2 | Differential GPS Fix |
| 4 | RTK Fixed (cm-level accuracy) |
| 5 | RTK Float (dm-level accuracy) |

## Stopping the Tool

Press Ctrl+C to stop monitoring and exit the tool.

## Troubleshooting

### No Serial Ports Found

- Make sure the device is properly connected to your computer
- Check if you have the necessary drivers installed
- On Linux, ensure you have the required permissions to access the serial port

### Connection Verification Failed

- Check if the device is powered on
- Verify that the baud rate is correct
- Try a different serial port

### Data Not Displaying

- Verify that the device is sending data in the expected format
- Try a different data mode
- Check the baud rate settings

### NTRIP Connection Issues

If you have trouble connecting to the NTRIP server:
- Verify the server address, port, username, password, and mountpoint
- Check your internet connection
- Ensure the NTRIP server is online and the mountpoint exists

### RTK Fix Not Achieved

If you can't get an RTK fixed solution:
- Ensure you have a clear view of the sky
- Check that the NTRIP corrections are appropriate for your location
- Verify that your GNSS receiver supports RTK
- Make sure the base station is within the recommended distance (typically <30km)
