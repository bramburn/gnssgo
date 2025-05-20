package caster

import (
	"context"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// Constants for NTRIP protocol
const (
	NTRIPVersionHeaderKey      = "Ntrip-Version"
	NTRIPVersionHeaderValueV1  = "Ntrip/1.0"
	NTRIPVersionHeaderValueV2  = "Ntrip/2.0"
	RequestIDContextKey        = "request_id"
	ErrorNotAuthorized         = Error("not authorized")
	ErrorNotFound              = Error("not found")
	ErrorInternalServerError   = Error("internal server error")
)

// Error is a simple error type
type Error string

func (e Error) Error() string {
	return string(e)
}

// SourceService represents a provider of stream data
type SourceService interface {
	GetSourcetable() Sourcetable
	// Publisher creates a new publisher for the given mountpoint
	Publisher(ctx context.Context, mount, username, password string) (io.WriteCloser, error)
	// Subscriber creates a new subscriber for the given mountpoint
	Subscriber(ctx context.Context, mount, username, password string) (chan []byte, error)
}

// Caster wraps http.Server, providing an NTRIP caster implementation
type Caster struct {
	http.Server
}

// NewCaster constructs a Caster, setting up the Handler and timeouts
func NewCaster(addr string, svc SourceService, logger logrus.FieldLogger) *Caster {
	return &Caster{
		http.Server{
			Addr:        addr,
			Handler:     getHandler(svc, logger),
			IdleTimeout: 10 * time.Second,
			// Read timeout kills publishing connections because they don't necessarily read from
			// the response body
			//ReadTimeout: 10 * time.Second,
			// Write timeout kills subscriber connections because they don't write to the request
			// body
			//WriteTimeout: 10 * time.Second,
		},
	}
}

// getHandler creates a new HTTP handler for the caster
func getHandler(svc SourceService, logger logrus.FieldLogger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestVersion := 1
		if strings.ToUpper(r.Header.Get(NTRIPVersionHeaderKey)) == strings.ToUpper(NTRIPVersionHeaderValueV2) {
			requestVersion = 2
		}

		requestID := uuid.New().String()
		ctx := context.WithValue(r.Context(), RequestIDContextKey, requestID)

		username, _, _ := r.BasicAuth()

		l := logger.WithFields(logrus.Fields{
			"request_id":      requestID,
			"request_version": requestVersion,
			"path":            r.URL.Path,
			"method":          r.Method,
			"source_ip":       r.RemoteAddr,
			"username":        username,
			"user_agent":      r.UserAgent(),
		})

		h := &handler{svc, l}
		h.handleRequest(w, r.WithContext(ctx))
	})
}
