package server

import (
	"context"
	"sync"
	"time"

	"github.com/bramburn/gnssgo/pkg/gnssgo"
)

// RTCMDataSource is a data source that provides RTCM data from a gnssgo.Stream
type RTCMDataSource struct {
	stream     gnssgo.Stream
	dataChan   chan []byte
	ctx        context.Context
	cancel     context.CancelFunc
	running    bool
	mutex      sync.Mutex
	bufferSize int
	interval   time.Duration
}

// NewRTCMDataSource creates a new RTCM data source
func NewRTCMDataSource(streamType int, streamPath string, bufferSize int, interval time.Duration) *RTCMDataSource {
	return &RTCMDataSource{
		dataChan:   make(chan []byte, 10),
		bufferSize: bufferSize,
		interval:   interval,
	}
}

// Start starts the data source
func (ds *RTCMDataSource) Start() error {
	ds.mutex.Lock()
	defer ds.mutex.Unlock()

	if ds.running {
		return nil
	}

	// Initialize the stream
	ds.stream.InitStream()

	// Create a cancellable context
	ds.ctx, ds.cancel = context.WithCancel(context.Background())

	// Start the data source in a goroutine
	go ds.run()

	ds.running = true
	return nil
}

// Stop stops the data source
func (ds *RTCMDataSource) Stop() error {
	ds.mutex.Lock()
	defer ds.mutex.Unlock()

	if !ds.running {
		return nil
	}

	// Cancel the context
	if ds.cancel != nil {
		ds.cancel()
	}

	// Close the stream
	ds.stream.StreamClose()

	// Close the data channel
	close(ds.dataChan)

	ds.running = false
	return nil
}

// Data returns the data channel
func (ds *RTCMDataSource) Data() <-chan []byte {
	return ds.dataChan
}

// run runs the data source
func (ds *RTCMDataSource) run() {
	buffer := make([]byte, ds.bufferSize)

	for {
		// Check if the context is done
		select {
		case <-ds.ctx.Done():
			return
		default:
		}

		// Read data from the stream
		n := ds.stream.StreamRead(buffer, ds.bufferSize)
		if n <= 0 {
			// Wait before retrying
			select {
			case <-ds.ctx.Done():
				return
			case <-time.After(ds.interval):
			}
			continue
		}

		// Copy the data to avoid race conditions
		data := make([]byte, n)
		copy(data, buffer[:n])

		// Send the data to the channel
		select {
		case ds.dataChan <- data:
		default:
			// Skip if the channel is full
		}

		// Wait before reading again
		select {
		case <-ds.ctx.Done():
			return
		case <-time.After(ds.interval):
		}
	}
}

// FileDataSource is a data source that provides RTCM data from a file
type FileDataSource struct {
	filePath   string
	dataChan   chan []byte
	ctx        context.Context
	cancel     context.CancelFunc
	running    bool
	mutex      sync.Mutex
	bufferSize int
	interval   time.Duration
}

// NewFileDataSource creates a new file data source
func NewFileDataSource(filePath string, bufferSize int, interval time.Duration) *FileDataSource {
	return &FileDataSource{
		filePath:   filePath,
		dataChan:   make(chan []byte, 10),
		bufferSize: bufferSize,
		interval:   interval,
	}
}

// Start starts the data source
func (ds *FileDataSource) Start() error {
	ds.mutex.Lock()
	defer ds.mutex.Unlock()

	if ds.running {
		return nil
	}

	// Create a cancellable context
	ds.ctx, ds.cancel = context.WithCancel(context.Background())

	// Start the data source in a goroutine
	go ds.run()

	ds.running = true
	return nil
}

// Stop stops the data source
func (ds *FileDataSource) Stop() error {
	ds.mutex.Lock()
	defer ds.mutex.Unlock()

	if !ds.running {
		return nil
	}

	// Cancel the context
	if ds.cancel != nil {
		ds.cancel()
	}

	// Close the data channel
	close(ds.dataChan)

	ds.running = false
	return nil
}

// Data returns the data channel
func (ds *FileDataSource) Data() <-chan []byte {
	return ds.dataChan
}

// run runs the data source
func (ds *FileDataSource) run() {
	var stream gnssgo.Stream
	stream.InitStream()

	// Open the file
	if stream.OpenStream(gnssgo.STR_FILE, gnssgo.STR_MODE_R, ds.filePath) <= 0 {
		return
	}
	defer stream.StreamClose()

	buffer := make([]byte, ds.bufferSize)

	for {
		// Check if the context is done
		select {
		case <-ds.ctx.Done():
			return
		default:
		}

		// Read data from the file
		n := stream.StreamRead(buffer, ds.bufferSize)
		if n <= 0 {
			// Reopen the file if we reached the end
			stream.StreamClose()
			if stream.OpenStream(gnssgo.STR_FILE, gnssgo.STR_MODE_R, ds.filePath) <= 0 {
				return
			}
			continue
		}

		// Copy the data to avoid race conditions
		data := make([]byte, n)
		copy(data, buffer[:n])

		// Send the data to the channel
		select {
		case ds.dataChan <- data:
		default:
			// Skip if the channel is full
		}

		// Wait before reading again
		select {
		case <-ds.ctx.Done():
			return
		case <-time.After(ds.interval):
		}
	}
}
