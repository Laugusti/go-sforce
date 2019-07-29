package testserver

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// Server is a wrapper for a test server.
type Server struct {
	s            *httptest.Server
	t            *testing.T
	RequestCount int
	HandlerFunc  http.HandlerFunc
}

// New returns a new started Server
func New(t *testing.T) *Server {
	s := &Server{t: t}
	s.Start()
	return s
}

// Start starts the server and sets the response to login success response
func (s *Server) Start() {
	// already started
	if s.s != nil {
		return
	}
	// reset counter and handler
	s.RequestCount = 0
	s.HandlerFunc = StaticJSONHandlerFunc(s.t, map[string]string{"message": "hello world"}, http.StatusOK)
	s.s = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.RequestCount++
		s.HandlerFunc(w, r)
	}))
}

// Stop stops the test server.
func (s *Server) Stop() {
	// already stopped
	if s.s == nil {
		return
	}
	s.s.CloseClientConnections()
	s.s.Close()
	s.s = nil
}

// Client returns a HTTP client for the test server.
func (s *Server) Client() *http.Client {
	// server not started
	if s.s == nil {
		return nil
	}
	return s.s.Client()
}

// URL returns the base url of server
func (s *Server) URL() string {
	// server not started
	if s.s == nil {
		return ""
	}
	return s.s.URL
}
