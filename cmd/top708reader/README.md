# TOP708 Reader CLI

A command-line tool for reading data from TOPGNSS TOP708 GNSS receivers.

## Overview

This tool allows you to connect to a TOPGNSS TOP708 GNSS receiver through a serial port and monitor the data streams in various formats. It supports:

- Raw data monitoring
- NMEA sentence parsing and display
- RTCM message monitoring
- UBX protocol message monitoring

## Installation

```bash
# From the repository root
cd cmd/top708reader
go build
```

This will create a `top708reader.exe` executable in the current directory.

## Usage

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

## Command-Line Options

| Option    | Description                                   | Default   |
|-----------|-----------------------------------------------|-----------|
| `-port`   | Serial port name (e.g., COM1, /dev/ttyUSB0)   | (prompt)  |
| `-baud`   | Baud rate                                     | 38400     |
| `-timeout`| Connection verification timeout               | 5s        |
| `-mode`   | Data mode: raw, nmea, rtcm, ubx               | raw       |
| `-list`   | List available ports and exit                 | false     |

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
