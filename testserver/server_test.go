package testserver

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/Laugusti/go-sforce/restclient"
	"github.com/Laugusti/go-sforce/session"
	"github.com/stretchr/testify/assert"
)

func TestServer(t *testing.T) {
	// create new server
	s := New()
	assert.Nil(t, s.s, "TestServer => server should not be started")
	assert.Nil(t, s.ServerResponse, "TestServer => response should be nil")

	// start server
	s.Start()
	assert.NotNil(t, s.s, "TestServer => server should be started")
	assert.NotNil(t, s.ServerResponse, "TestServer => response should not be nil")
	assert.NotEmpty(t, s.URL(), "TestServer => server url should not be empty")
	assert.NotNil(t, s.Client(), "TestServer => client should not be nil")

	// create request
	var token session.RequestToken
	statusCode, err := doRequest(s, &token)
	assert.Nilf(t, err, "TestServer => unexpected error: %v", err)
	assert.Equal(t, http.StatusOK, statusCode)
	assert.Equal(t, TestAccessToken, token.AccessToken, "TestServer => wrong access token")
	assert.Equal(t, s.URL(), token.InstanceURL, "TestServer => wrong instance url")

	// new response
	s.ServerResponse = UnauthorizedResponse
	var apiErr restclient.APIError
	statusCode, err = doRequest(s, &apiErr)
	assert.Nilf(t, err, "TestServer => unexpected error: %v", err)
	assert.Equal(t, http.StatusUnauthorized, statusCode)
	assert.Equal(t, "INVALID_SESSION_ID", apiErr.ErrorCode)

	// stop server
	s.Stop()
	assert.Nil(t, s.s)
}

func doRequest(s *Server, result interface{}) (int, error) {
	resp, err := s.Client().Get(s.URL())
	if err != nil {
		return 0, err
	}
	defer func() { _ = resp.Body.Close() }()
	return resp.StatusCode, json.NewDecoder(resp.Body).Decode(result)
}
