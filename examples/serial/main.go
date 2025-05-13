package main

import (
	"fmt"
	"time"

	"github.com/bramburn/gnssgo"
)

// Example demonstrating serial communication with GNSS receivers
func main() {
	fmt.Println("GNSSGO Serial Communication Example")
	fmt.Println("----------------------------------")

	// Create a stream object
	var stream gnssgo.Stream

	// Serial port configuration
	// Format: port[:brate[:bsize[:parity[:stopb[:fctr[#port]]]]]]
	// Example: COM1:115200:8:N:1:off
	serialConfig := "COM1:115200:8:N:1:off"

	fmt.Printf("Opening serial port with configuration: %s\n", serialConfig)
	fmt.Println("(This is a demonstration - port should be replaced with an actual port)")

	// In a real application, you would:
	// status := stream.OpenStream(gnssgo.STR_SERIAL, gnssgo.STR_MODE_RW, serialConfig)
	// if status == 0 {
	//     fmt.Println("Failed to open serial port")
	//     return
	// }
	// fmt.Println("Serial port opened successfully")

	// Read data from the stream
	fmt.Println("\nReading data from serial port:")
	fmt.Println("1. Create a buffer to store the data")
	fmt.Println("2. Call stream.StreamRead() to read data into the buffer")
	fmt.Println("3. Process the data according to your needs")

	// Example code for reading data:
	// buff := make([]byte, 1024)
	// for {
	//     n := stream.StreamRead(buff, 1024)
	//     if n > 0 {
	//         fmt.Printf("Read %d bytes from serial port\n", n)
	//         // Process the data...
	//     }
	//     time.Sleep(100 * time.Millisecond)
	// }

	// Write data to the stream
	fmt.Println("\nWriting data to serial port:")
	fmt.Println("1. Prepare the data to be sent")
	fmt.Println("2. Call stream.StreamWrite() to send the data")

	// Example code for writing data:
	// data := []byte("$GPGGA,123519,4807.038,N,01131.000,E,1,08,0.9,545.4,M,46.9,M,,*47\r\n")
	// n := stream.StreamWrite(data, len(data))
	// fmt.Printf("Wrote %d bytes to serial port\n", n)

	// Close the stream when done
	fmt.Println("\nClosing the serial port:")
	fmt.Println("Call stream.StreamClose() to close the connection")

	// stream.StreamClose()
	// fmt.Println("Serial port closed")
}
