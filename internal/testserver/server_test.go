package testserver

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestServer(t *testing.T) {
	// create new server
	s := New(t)
	assert.NotNil(t, s.s, "server should be started")
	assert.NotNil(t, s.Client(), "client should not be nil")
	assert.NotEmpty(t, s.URL(), "server url should not be empty")

	// create request
	statusCode, jsonResp, err := doRequest(s)
	assert.Nilf(t, err, "unexpected error: %v", err)
	assert.Equal(t, http.StatusOK, statusCode)
	assert.Equal(t, "hello world", jsonResp["message"], "wrong mesage")
	assert.Equal(t, 1, s.RequestCount)

	// new response
	s.HandlerFunc = StaticJSONHandlerFunc(t, 400, map[string]string{"message": "error"})
	statusCode, jsonResp, err = doRequest(s)
	assert.Nilf(t, err, "unexpected error: %v", err)
	assert.Equal(t, 400, statusCode)
	assert.Equal(t, "error", jsonResp["message"], "wrong message")
	assert.Equal(t, 2, s.RequestCount)

	// restart server and send new request
	s.Stop()
	s.Start()
	statusCode, jsonResp, err = doRequest(s)
	assert.Nilf(t, err, "unexpected error: %v", err)
	assert.Equal(t, http.StatusOK, statusCode)
	assert.Equal(t, "hello world", jsonResp["message"], "wrong mesage")
	assert.Equal(t, 1, s.RequestCount)

	// stop server
	s.Stop()
	assert.Nil(t, s.s)
}

func doRequest(s *Server) (int, map[string]interface{}, error) {
	resp, err := s.Client().Get(s.URL())
	if err != nil {
		return 0, nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	m, err := jsonReaderToMap(resp.Body)
	return resp.StatusCode, m, err
}
