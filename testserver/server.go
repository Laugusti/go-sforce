package testserver

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/Laugusti/go-sforce/restclient"
	"github.com/Laugusti/go-sforce/session"
)

const (
	// TestAccessToken is a access token used in the login success response
	TestAccessToken = "MOCK_TOKEN"
)

// Server is a wrapper for a test server.
type Server struct {
	s              *httptest.Server
	ServerResponse func(w http.ResponseWriter, r *http.Request)
}

// New returns a new unstarted Server
func New() *Server {
	s := &Server{}
	return s
}

// Start starts the server and sets the response to login success response
func (s *Server) Start() {
	// already started
	if s.s != nil {
		return
	}
	s.ServerResponse = LoginSuccessResponse
	s.s = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.ServerResponse(w, r)
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
	return s.s.URL
}

var (
	// LoginSuccessResponse is the response for a successful login
	LoginSuccessResponse = func(w http.ResponseWriter, r *http.Request) {
		serverURL := fmt.Sprintf("http://%s", r.Host)
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(session.RequestToken{
			AccessToken: TestAccessToken,
			InstanceURL: serverURL,
		})
	}

	// UnauthorizedResponse is the response for a unauthorized request
	UnauthorizedResponse = func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(restclient.APIError{
			Message:   "Session expired or invalid",
			ErrorCode: "INVALID_SESSION_ID",
		})
	}

	// CreateSuccessResponse is the response for a successful create.
	CreateSuccessResponse = func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(restclient.UpsertResult{
			ID:      "id",
			Success: true,
			Errors:  []interface{}{},
		})
	}

	// UpdateSuccessResponse is the response for a successful create.
	UpdateSuccessResponse = func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(restclient.UpsertResult{
			ID:      "id",
			Success: true,
			Errors:  []interface{}{},
		})
	}
)
