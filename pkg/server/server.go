package server

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// Constants for NTRIP protocol
const (
	NTRIPVersionHeaderKey     = "Ntrip-Version"
	NTRIPVersionHeaderValueV2 = "Ntrip/2.0"
	UserAgentHeaderKey        = "User-Agent"
	UserAgentValue            = "GNSSGO NTRIP Server/1.0"
)

// DataSource is an interface for providing RTCM data to the server
type DataSource interface {
	// Start starts the data source
	Start() error
	// Stop stops the data source
	Stop() error
	// Data returns a channel that provides RTCM data
	Data() <-chan []byte
}

// Server represents an NTRIP server
type Server struct {
	host        string
	port        string
	username    string
	password    string
	mountpoint  string
	dataSource  DataSource
	client      *http.Client
	running     bool
	ctx         context.Context
	cancel      context.CancelFunc
	mutex       sync.Mutex
	logger      logrus.FieldLogger
}

// NewServer creates a new NTRIP server
func NewServer(host, port, username, password, mountpoint string, logger logrus.FieldLogger) *Server {
	return &Server{
		host:       host,
		port:       port,
		username:   username,
		password:   password,
		mountpoint: mountpoint,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: logger,
	}
}

// SetDataSource sets the data source for the server
func (s *Server) SetDataSource(dataSource DataSource) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.dataSource = dataSource
}

// Start starts the server
func (s *Server) Start() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.running {
		return fmt.Errorf("server already running")
	}

	if s.dataSource == nil {
		return fmt.Errorf("no data source set")
	}

	// Create a cancellable context
	s.ctx, s.cancel = context.WithCancel(context.Background())

	// Start the data source
	if err := s.dataSource.Start(); err != nil {
		return fmt.Errorf("failed to start data source: %w", err)
	}

	// Start the server in a goroutine
	go s.run()

	s.running = true
	return nil
}

// Stop stops the server
func (s *Server) Stop() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if !s.running {
		return nil
	}

	// Cancel the context
	if s.cancel != nil {
		s.cancel()
	}

	// Stop the data source
	if s.dataSource != nil {
		if err := s.dataSource.Stop(); err != nil {
			return fmt.Errorf("failed to stop data source: %w", err)
		}
	}

	s.running = false
	return nil
}

// run runs the server
func (s *Server) run() {
	s.logger.Infof("Starting NTRIP server for mountpoint %s", s.mountpoint)

	for {
		// Check if the context is done
		select {
		case <-s.ctx.Done():
			s.logger.Info("Server stopped")
			return
		default:
		}

		// Connect to the caster
		err := s.connect()
		if err != nil {
			s.logger.Errorf("Failed to connect to caster: %v", err)
			// Wait before retrying
			select {
			case <-s.ctx.Done():
				return
			case <-time.After(5 * time.Second):
			}
			continue
		}

		// Wait before reconnecting
		select {
		case <-s.ctx.Done():
			return
		case <-time.After(5 * time.Second):
		}
	}
}

// connect connects to the caster and streams data
func (s *Server) connect() error {
	// Create the URL
	url := fmt.Sprintf("http://%s:%s/%s", s.host, s.port, s.mountpoint)

	// Create the request
	req, err := http.NewRequestWithContext(s.ctx, http.MethodPost, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set(NTRIPVersionHeaderKey, NTRIPVersionHeaderValueV2)
	req.Header.Set(UserAgentHeaderKey, UserAgentValue)
	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("Transfer-Encoding", "chunked")

	// Set basic auth
	req.SetBasicAuth(s.username, s.password)

	// Create a pipe for streaming data
	pr, pw := io.Pipe()
	req.Body = pr

	// Start a goroutine to write data to the pipe
	go func() {
		defer pw.Close()

		for {
			select {
			case <-s.ctx.Done():
				return
			case data, ok := <-s.dataSource.Data():
				if !ok {
					return
				}
				_, err := pw.Write(data)
				if err != nil {
					s.logger.Errorf("Failed to write data to pipe: %v", err)
					return
				}
			}
		}
	}()

	// Send the request
	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Check the response
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	s.logger.Infof("Connected to caster at %s", url)

	// Read the response body to keep the connection alive
	_, err = io.Copy(io.Discard, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	return nil
}
