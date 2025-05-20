package caster

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"
)

// handler is used by Caster to handle HTTP requests
type handler struct {
	svc    SourceService
	logger logrus.FieldLogger
}

// handleRequest handles both NTRIP v1 and v2 requests
func (h *handler) handleRequest(w http.ResponseWriter, r *http.Request) {
	h.logger.Debug("request received")
	defer r.Body.Close()
	switch strings.ToUpper(r.Header.Get(NTRIPVersionHeaderKey)) {
	case strings.ToUpper(NTRIPVersionHeaderValueV2):
		h.handleRequestV2(w, r)
	default:
		h.handleRequestV1(w, r)
	}
}

// handleRequestV1 handles NTRIP v1 requests
func (h *handler) handleRequestV1(w http.ResponseWriter, r *http.Request) {
	// Can only support NTRIP v1 GET requests with http.Server
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusNotImplemented)
		return
	}

	// Extract underlying net.Conn from ResponseWriter
	hj, ok := w.(http.Hijacker)
	if !ok {
		h.logger.Error("server does not implement hijackable response writers, cannot support NTRIP v1")
		// There is no NTRIP v1 response to signal failure, so this is probably the most useful
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	// Get the sourcetable if the path is "/"
	if r.URL.Path == "/" {
		h.handleSourcetable(w, r)
		return
	}

	// Get the mountpoint from the path
	mount := strings.TrimPrefix(r.URL.Path, "/")
	if mount == "" {
		h.handleSourcetable(w, r)
		return
	}

	// Get the username and password from the request
	username, password, _ := r.BasicAuth()

	// Get subscriber
	sub, err := h.svc.Subscriber(r.Context(), mount, username, password)
	if err != nil {
		h.logger.Infof("connection refused with reason: %s", err)
		// NTRIP v1 says to return 401 for unauthorized, but sourcetable for any other error - this goes against that
		if err == ErrorNotAuthorized {
			writeStatusV1(w, r, http.StatusUnauthorized)
		} else if err == ErrorNotFound {
			writeStatusV1(w, r, http.StatusNotFound)
		} else {
			writeStatusV1(w, r, http.StatusInternalServerError)
		}
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
		return
	}

	// Write the NTRIP v1 response header
	_, err = w.Write([]byte("ICY 200 OK\r\n"))
	if err != nil {
		h.logger.WithError(err).Error("failed to write response headers")
		return
	}
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}

	// Get the connection
	conn, bufrw, err := hj.Hijack()
	if err != nil {
		h.logger.WithError(err).Error("failed to hijack connection")
		return
	}
	defer conn.Close()

	// Stream data to the client
	for data := range sub {
		_, err := bufrw.Write(data)
		if err != nil {
			h.logger.WithError(err).Error("failed to write to client")
			return
		}
		err = bufrw.Flush()
		if err != nil {
			h.logger.WithError(err).Error("failed to flush to client")
			return
		}
	}
}

// handleRequestV2 handles NTRIP v2 requests
func (h *handler) handleRequestV2(w http.ResponseWriter, r *http.Request) {
	// Get the sourcetable if the path is "/"
	if r.URL.Path == "/" {
		h.handleSourcetable(w, r)
		return
	}

	// Get the mountpoint from the path
	mount := strings.TrimPrefix(r.URL.Path, "/")
	if mount == "" {
		h.handleSourcetable(w, r)
		return
	}

	// Get the username and password from the request
	username, password, _ := r.BasicAuth()

	// Handle POST requests (publishers)
	if r.Method == http.MethodPost {
		h.handlePublisher(w, r, mount, username, password)
		return
	}

	// Handle GET requests (subscribers)
	if r.Method == http.MethodGet {
		h.handleSubscriber(w, r, mount, username, password)
		return
	}

	// Unsupported method
	w.WriteHeader(http.StatusMethodNotAllowed)
}

// handleSourcetable handles sourcetable requests
func (h *handler) handleSourcetable(w http.ResponseWriter, r *http.Request) {
	sourcetable := h.svc.GetSourcetable()
	sourcetableStr := sourcetable.String()

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte(sourcetableStr))
	if err != nil {
		h.logger.WithError(err).Error("failed to write sourcetable")
	}
}

// handlePublisher handles publisher requests
func (h *handler) handlePublisher(w http.ResponseWriter, r *http.Request, mount, username, password string) {
	// Get publisher
	pub, err := h.svc.Publisher(r.Context(), mount, username, password)
	if err != nil {
		h.logger.Infof("publisher connection refused with reason: %s", err)
		if err == ErrorNotAuthorized {
			w.WriteHeader(http.StatusUnauthorized)
		} else if err == ErrorNotFound {
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	defer pub.Close()

	// Write success response
	w.WriteHeader(http.StatusOK)
	w.(http.Flusher).Flush()

	// Copy data from the request body to the publisher
	_, err = io.Copy(pub, r.Body)
	if err != nil {
		h.logger.WithError(err).Error("failed to copy data from publisher")
	}
}

// handleSubscriber handles subscriber requests
func (h *handler) handleSubscriber(w http.ResponseWriter, r *http.Request, mount, username, password string) {
	// Get subscriber
	sub, err := h.svc.Subscriber(r.Context(), mount, username, password)
	if err != nil {
		h.logger.Infof("subscriber connection refused with reason: %s", err)
		if err == ErrorNotAuthorized {
			w.WriteHeader(http.StatusUnauthorized)
		} else if err == ErrorNotFound {
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	// Set headers for chunked transfer encoding
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Transfer-Encoding", "chunked")
	w.WriteHeader(http.StatusOK)
	w.(http.Flusher).Flush()

	// Stream data to the client
	for data := range sub {
		_, err := w.Write(data)
		if err != nil {
			h.logger.WithError(err).Error("failed to write to client")
			return
		}
		w.(http.Flusher).Flush()
	}
}

// writeStatusV1 writes an NTRIP v1 status response
func writeStatusV1(w http.ResponseWriter, r *http.Request, status int) {
	switch status {
	case http.StatusUnauthorized:
		w.Header().Set("WWW-Authenticate", `Basic realm="NTRIP Caster"`)
		w.WriteHeader(http.StatusUnauthorized)
	case http.StatusNotFound:
		// For NTRIP v1, return the sourcetable for not found
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "SOURCETABLE 200 OK\r\n\r\nENDSOURCETABLE\r\n")
	default:
		w.WriteHeader(status)
	}
}
