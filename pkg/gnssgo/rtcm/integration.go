package rtcm

import (
	"fmt"
	"sync"
	"time"
)

// RTCMProcessor processes RTCM messages from a byte stream
type RTCMProcessor struct {
	parser        *RTCMParser
	buffer        []byte
	messages      []RTCMMessage
	messagesMutex sync.Mutex
	callbacks     map[int][]RTCMMessageCallback
	callbackMutex sync.Mutex
}

// RTCMMessageCallback is a callback function for RTCM messages
type RTCMMessageCallback func(msg *RTCMMessage)

// NewRTCMProcessor creates a new RTCM processor
func NewRTCMProcessor() *RTCMProcessor {
	return &RTCMProcessor{
		parser:    NewRTCMParser(),
		buffer:    make([]byte, 0, 1024),
		messages:  make([]RTCMMessage, 0),
		callbacks: make(map[int][]RTCMMessageCallback),
	}
}

// ProcessData processes RTCM data from a byte stream
func (p *RTCMProcessor) ProcessData(data []byte) error {
	// Parse RTCM messages
	messages, remaining, err := p.parser.ParseRTCMMessage(data)
	if err != nil {
		return fmt.Errorf("error parsing RTCM messages: %w", err)
	}

	// Store remaining data in buffer
	p.buffer = remaining

	// Store messages
	p.messagesMutex.Lock()
	p.messages = append(p.messages, messages...)
	p.messagesMutex.Unlock()

	// Process messages
	for i := range messages {
		p.processMessage(&messages[i])
	}

	return nil
}

// processMessage processes a single RTCM message
func (p *RTCMProcessor) processMessage(msg *RTCMMessage) {
	// Call callbacks for this message type
	p.callbackMutex.Lock()
	callbacks := p.callbacks[msg.Type]
	p.callbackMutex.Unlock()

	for _, callback := range callbacks {
		callback(msg)
	}

	// Call callbacks for all message types
	p.callbackMutex.Lock()
	callbacks = p.callbacks[0]
	p.callbackMutex.Unlock()

	for _, callback := range callbacks {
		callback(msg)
	}
}

// RegisterCallback registers a callback function for a specific message type
// If messageType is 0, the callback will be called for all message types
func (p *RTCMProcessor) RegisterCallback(messageType int, callback RTCMMessageCallback) {
	p.callbackMutex.Lock()
	defer p.callbackMutex.Unlock()

	p.callbacks[messageType] = append(p.callbacks[messageType], callback)
}

// UnregisterCallback unregisters a callback function for a specific message type
func (p *RTCMProcessor) UnregisterCallback(messageType int, callback RTCMMessageCallback) {
	p.callbackMutex.Lock()
	defer p.callbackMutex.Unlock()

	callbacks := p.callbacks[messageType]
	for i, cb := range callbacks {
		if fmt.Sprintf("%p", cb) == fmt.Sprintf("%p", callback) {
			p.callbacks[messageType] = append(callbacks[:i], callbacks[i+1:]...)
			break
		}
	}
}

// GetMessages returns all stored messages
func (p *RTCMProcessor) GetMessages() []RTCMMessage {
	p.messagesMutex.Lock()
	defer p.messagesMutex.Unlock()

	// Create a copy of the messages
	messages := make([]RTCMMessage, len(p.messages))
	copy(messages, p.messages)

	return messages
}

// GetMessagesByType returns all stored messages of a specific type
func (p *RTCMProcessor) GetMessagesByType(messageType int) []RTCMMessage {
	p.messagesMutex.Lock()
	defer p.messagesMutex.Unlock()

	var messages []RTCMMessage
	for _, msg := range p.messages {
		if msg.Type == messageType {
			messages = append(messages, msg)
		}
	}

	return messages
}

// GetLatestMessageByType returns the latest message of a specific type
func (p *RTCMProcessor) GetLatestMessageByType(messageType int) *RTCMMessage {
	p.messagesMutex.Lock()
	defer p.messagesMutex.Unlock()

	var latest *RTCMMessage
	var latestTime time.Time

	for i, msg := range p.messages {
		if msg.Type == messageType && (latest == nil || msg.Timestamp.After(latestTime)) {
			latest = &p.messages[i]
			latestTime = msg.Timestamp
		}
	}

	return latest
}

// ClearMessages clears all stored messages
func (p *RTCMProcessor) ClearMessages() {
	p.messagesMutex.Lock()
	defer p.messagesMutex.Unlock()

	p.messages = p.messages[:0]
}

// GetStats returns the statistics for all message types
func (p *RTCMProcessor) GetStats() map[int]*RTCMMessageStats {
	return p.parser.GetStats()
}

// FilterRTCMMessages filters RTCM messages based on a filter function
func FilterRTCMMessages(messages []RTCMMessage, filter func(msg *RTCMMessage) bool) []RTCMMessage {
	var filtered []RTCMMessage
	for i, msg := range messages {
		if filter(&messages[i]) {
			filtered = append(filtered, msg)
		}
	}
	return filtered
}

// DefaultRTCMFilter provides a default filter that excludes unnecessary message types
func DefaultRTCMFilter(msg *RTCMMessage) bool {
	// Include only essential message types
	switch msg.Type {
	case RTCM_STATION_COORDINATES, RTCM_STATION_COORDINATES_ALT:
		return true
	case RTCM_GPS_EPHEMERIS, RTCM_GLONASS_EPHEMERIS:
		return true
	case MSM_GPS_RANGE_START + 3, MSM_GPS_RANGE_START + 4, MSM_GPS_RANGE_START + 5, MSM_GPS_RANGE_START + 6, MSM_GPS_RANGE_START + 7:
		return true
	case MSM_GLONASS_RANGE_START + 3, MSM_GLONASS_RANGE_START + 4, MSM_GLONASS_RANGE_START + 5, MSM_GLONASS_RANGE_START + 6, MSM_GLONASS_RANGE_START + 7:
		return true
	case MSM_GALILEO_RANGE_START + 3, MSM_GALILEO_RANGE_START + 4, MSM_GALILEO_RANGE_START + 5, MSM_GALILEO_RANGE_START + 6, MSM_GALILEO_RANGE_START + 7:
		return true
	case MSM_BEIDOU_RANGE_START + 3, MSM_BEIDOU_RANGE_START + 4, MSM_BEIDOU_RANGE_START + 5, MSM_BEIDOU_RANGE_START + 6, MSM_BEIDOU_RANGE_START + 7:
		return true
	default:
		return false
	}
}

// CriticalRTCMFilter provides a filter that only allows critical message types for RTK
func CriticalRTCMFilter(msg *RTCMMessage) bool {
	// Include only critical message types for RTK
	switch msg.Type {
	case RTCM_STATION_COORDINATES, RTCM_STATION_COORDINATES_ALT:
		return true
	case MSM_GPS_RANGE_START + 4, MSM_GPS_RANGE_START + 7:
		return true
	case MSM_GLONASS_RANGE_START + 4, MSM_GLONASS_RANGE_START + 7:
		return true
	case MSM_GALILEO_RANGE_START + 4, MSM_GALILEO_RANGE_START + 7:
		return true
	case MSM_BEIDOU_RANGE_START + 4, MSM_BEIDOU_RANGE_START + 7:
		return true
	default:
		return false
	}
}

// RTCMMessageFilter is a function type for filtering RTCM messages
type RTCMMessageFilter func(msg *RTCMMessage) bool

// RTCMMessageFilterChain combines multiple filters with AND logic
func RTCMMessageFilterChain(filters ...RTCMMessageFilter) RTCMMessageFilter {
	return func(msg *RTCMMessage) bool {
		for _, filter := range filters {
			if !filter(msg) {
				return false
			}
		}
		return true
	}
}

// RTCMMessageTypeFilter creates a filter for specific message types
func RTCMMessageTypeFilter(types ...int) RTCMMessageFilter {
	return func(msg *RTCMMessage) bool {
		for _, t := range types {
			if msg.Type == t {
				return true
			}
		}
		return false
	}
}

// RTCMMessageStationFilter creates a filter for specific station IDs
func RTCMMessageStationFilter(stationIDs ...uint16) RTCMMessageFilter {
	return func(msg *RTCMMessage) bool {
		for _, id := range stationIDs {
			if msg.StationID == id {
				return true
			}
		}
		return false
	}
}

// RTCMMessageTimeFilter creates a filter for messages within a time range
func RTCMMessageTimeFilter(start, end time.Time) RTCMMessageFilter {
	return func(msg *RTCMMessage) bool {
		return (start.IsZero() || !msg.Timestamp.Before(start)) &&
			(end.IsZero() || !msg.Timestamp.After(end))
	}
}
