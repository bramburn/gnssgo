package rtcm_test

import (
	"fmt"
	"log"
	"time"

	"github.com/bramburn/gnssgo/pkg/gnssgo/rtcm"
	"github.com/bramburn/gnssgo/pkg/gnssgo/stream"
)

// ExampleRTCMParser demonstrates how to use the RTCM parser
func Example_rTCMParser() {
	// Create a new RTCM parser
	parser := rtcm.NewRTCMParser()

	// Sample RTCM data (this is just an example, not real RTCM data)
	data := []byte{
		0xD3, 0x00, 0x13, // Header (preamble + length)
		0x3E, 0xD7, 0xD3, 0x02, 0x02, 0x98, 0x0E, 0xDE, 0xEF, 0x34, 0xB4, 0xBD, 0x62, 0xAC, 0x09, 0x41, 0x98, 0x6F, 0x33, // Data
		0x36, 0x0B, 0x98, // CRC
	}

	// Parse RTCM messages
	messages, remaining, err := parser.ParseRTCMMessage(data)
	if err != nil {
		log.Fatalf("Failed to parse RTCM message: %v", err)
	}

	// Print the results
	fmt.Printf("Parsed %d messages\n", len(messages))
	fmt.Printf("Remaining bytes: %d\n", len(remaining))

	// Print message details
	for i, msg := range messages {
		fmt.Printf("Message %d: Type=%d, Length=%d, StationID=%d\n",
			i+1, msg.Type, msg.Length, msg.StationID)
	}

	// Get message statistics
	stats := parser.GetStats()
	fmt.Printf("Message statistics:\n")
	for msgType, stat := range stats {
		fmt.Printf("Type %d: Count=%d, TotalBytes=%d\n",
			msgType, stat.Count, stat.TotalBytes)
	}

	// Output:
	// Parsed 1 messages
	// Remaining bytes: 0
	// Message 1: Type=1005, Length=19, StationID=1000
	// Message statistics:
	// Type 1005: Count=1, TotalBytes=19
}

// ExampleRTCMProcessor demonstrates how to use the RTCM processor
func Example_rTCMProcessor() {
	// Create a new RTCM processor
	processor := rtcm.NewRTCMProcessor()

	// Register a callback for all message types
	processor.RegisterCallback(0, func(msg *rtcm.RTCMMessage) {
		fmt.Printf("Received message: Type=%d, Length=%d, StationID=%d\n",
			msg.Type, msg.Length, msg.StationID)
	})

	// Register a callback for a specific message type
	processor.RegisterCallback(rtcm.RTCM_STATION_COORDINATES, func(msg *rtcm.RTCMMessage) {
		fmt.Printf("Received station coordinates message: StationID=%d\n",
			msg.StationID)
	})

	// Sample RTCM data (this is just an example, not real RTCM data)
	data := []byte{
		0xD3, 0x00, 0x13, // Header (preamble + length)
		0x3E, 0xD7, 0xD3, 0x02, 0x02, 0x98, 0x0E, 0xDE, 0xEF, 0x34, 0xB4, 0xBD, 0x62, 0xAC, 0x09, 0x41, 0x98, 0x6F, 0x33, // Data
		0x36, 0x0B, 0x98, // CRC
	}

	// Process RTCM data
	err := processor.ProcessData(data)
	if err != nil {
		log.Fatalf("Failed to process RTCM data: %v", err)
	}

	// Get all messages
	messages := processor.GetMessages()
	fmt.Printf("Stored %d messages\n", len(messages))

	// Get messages of a specific type
	stationMessages := processor.GetMessagesByType(rtcm.RTCM_STATION_COORDINATES)
	fmt.Printf("Stored %d station coordinates messages\n", len(stationMessages))

	// Get the latest message of a specific type
	latestStation := processor.GetLatestMessageByType(rtcm.RTCM_STATION_COORDINATES)
	if latestStation != nil {
		fmt.Printf("Latest station coordinates message: StationID=%d\n",
			latestStation.StationID)
	}

	// Clear messages
	processor.ClearMessages()
	fmt.Printf("After clearing, stored %d messages\n", len(processor.GetMessages()))

	// Output:
	// Received message: Type=1005, Length=19, StationID=1000
	// Received station coordinates message: StationID=1000
	// Stored 1 messages
	// Stored 1 station coordinates messages
	// Latest station coordinates message: StationID=1000
	// After clearing, stored 0 messages
}

// ExampleIntegrationWithNTRIP demonstrates how to integrate the RTCM parser with an NTRIP client
func Example_integrationWithNTRIP() {
	// This is a conceptual example and won't actually run in tests
	// Create an NTRIP client configuration
	config := stream.DefaultNTripConfig()
	config.Server = "rtk2go.com"
	config.Port = 2101
	config.Mountpoint = "EXAMPLE"
	config.Username = "user"
	config.Password = "pass"
	config.Debug = true

	// Create an NTRIP client
	client := stream.NewEnhancedNTrip(config, stream.STR_NTRIPCLI)
	if client == nil {
		log.Fatalf("Failed to create NTRIP client")
	}

	// Connect to the NTRIP server
	err := client.Connect()
	if err != nil {
		log.Fatalf("Failed to connect to NTRIP server: %v", err)
	}

	// Create an RTCM processor
	processor := rtcm.NewRTCMProcessor()

	// Register callbacks for specific message types
	processor.RegisterCallback(rtcm.RTCM_STATION_COORDINATES, func(msg *rtcm.RTCMMessage) {
		fmt.Printf("Received station coordinates: StationID=%d\n", msg.StationID)
		// Decode the message
		stationCoords, err := rtcm.DecodeRTCMMessage(msg)
		if err != nil {
			log.Printf("Failed to decode station coordinates: %v", err)
			return
		}
		// Process the station coordinates
		fmt.Printf("Station coordinates: %+v\n", stationCoords)
	})

	processor.RegisterCallback(rtcm.MSM_GPS_RANGE_START+4, func(msg *rtcm.RTCMMessage) {
		fmt.Printf("Received GPS MSM4 message: StationID=%d\n", msg.StationID)
		// Decode the message
		msmData, err := rtcm.DecodeRTCMMessage(msg)
		if err != nil {
			log.Printf("Failed to decode GPS MSM4 message: %v", err)
			return
		}
		// Process the MSM data
		fmt.Printf("GPS MSM4 data: %+v\n", msmData)
	})

	// Start a goroutine to read data from the NTRIP client and process it
	go func() {
		buffer := make([]byte, 1024)
		for {
			// Read data from the NTRIP client
			n, err := client.Read(buffer)
			if err != nil {
				log.Printf("Failed to read from NTRIP client: %v", err)
				break
			}

			// Process the data
			err = processor.ProcessData(buffer[:n])
			if err != nil {
				log.Printf("Failed to process RTCM data: %v", err)
			}

			// Sleep to avoid busy-waiting
			time.Sleep(10 * time.Millisecond)
		}
	}()

	// Wait for some time to receive messages
	time.Sleep(10 * time.Second)

	// Close the NTRIP client
	client.Close()

	// Print statistics
	stats := processor.GetStats()
	fmt.Printf("Message statistics:\n")
	for msgType, stat := range stats {
		fmt.Printf("Type %d (%s): Count=%d, TotalBytes=%d\n",
			msgType, rtcm.GetMessageTypeDescription(msgType), stat.Count, stat.TotalBytes)
	}

	// Output:
	// (Example output would depend on actual NTRIP data received)
}

// ExampleFilterRTCMMessages demonstrates how to filter RTCM messages
func Example_filterRTCMMessages() {
	// Create some sample messages
	messages := []rtcm.RTCMMessage{
		{Type: rtcm.RTCM_STATION_COORDINATES, StationID: 1000},
		{Type: rtcm.RTCM_GPS_EPHEMERIS, StationID: 1000},
		{Type: rtcm.MSM_GPS_RANGE_START + 4, StationID: 1000},
		{Type: rtcm.MSM_GLONASS_RANGE_START + 4, StationID: 1000},
		{Type: 1230, StationID: 1000}, // Some other message type
	}

	// Filter messages using the default filter
	filtered := rtcm.FilterRTCMMessages(messages, rtcm.DefaultRTCMFilter)
	fmt.Printf("After default filter: %d messages\n", len(filtered))

	// Filter messages using the critical filter
	filtered = rtcm.FilterRTCMMessages(messages, rtcm.CriticalRTCMFilter)
	fmt.Printf("After critical filter: %d messages\n", len(filtered))

	// Create a custom filter for specific message types
	customFilter := rtcm.RTCMMessageTypeFilter(
		rtcm.RTCM_STATION_COORDINATES,
		rtcm.MSM_GPS_RANGE_START+4,
	)
	filtered = rtcm.FilterRTCMMessages(messages, customFilter)
	fmt.Printf("After custom filter: %d messages\n", len(filtered))

	// Create a filter chain
	filterChain := rtcm.RTCMMessageFilterChain(
		rtcm.RTCMMessageTypeFilter(rtcm.MSM_GPS_RANGE_START+4, rtcm.MSM_GLONASS_RANGE_START+4),
		rtcm.RTCMMessageStationFilter(1000),
	)
	filtered = rtcm.FilterRTCMMessages(messages, filterChain)
	fmt.Printf("After filter chain: %d messages\n", len(filtered))

	// Output:
	// After default filter: 4 messages
	// After critical filter: 3 messages
	// After custom filter: 2 messages
	// After filter chain: 2 messages
}
