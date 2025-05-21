package util

import (
	"time"
)

// TickGet returns the current tick count in milliseconds
func TickGet() uint32 {
	return uint32(time.Now().UnixNano() / int64(time.Millisecond))
}

// Sleepms sleeps for the specified number of milliseconds
func Sleepms(ms int) {
	time.Sleep(time.Duration(ms) * time.Millisecond)
}

// Tracet prints a trace message (placeholder for now)
func Tracet(level int, format string, args ...interface{}) {
	// Placeholder for trace functionality
}
