package testserver

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestServer(t *testing.T) {
	// create new server
	s := New()
	assert.Nil(t, s.s, "TestServer => server should not be started")
	//assert.NotNil(t, s.HandlerFunc, "TestServer => handler should not be nil")

	// start server
	s.Start()
	assert.NotNil(t, s.s, "TestServer => server should be started")
	assert.NotNil(t, s.Client(), "TestServer => client should not be nil")
	assert.NotEmpty(t, s.URL(), "TestServer => server url should not be empty")

	// create request
	jsonResp := make(map[string]interface{})
	statusCode, err := doRequest(s, &jsonResp)
	assert.Nilf(t, err, "TestServer => unexpected error: %v", err)
	assert.Equal(t, http.StatusOK, statusCode)
	assert.Equal(t, "hello world", jsonResp["message"], "TestServer => wrong mesage")

	// new response
	s.HandlerFunc = JSONResponse(map[string]string{"message": "error"}, 400)
	statusCode, err = doRequest(s, &jsonResp)
	assert.Nilf(t, err, "TestServer => unexpected error: %v", err)
	assert.Equal(t, 400, statusCode)
	assert.Equal(t, "error", jsonResp["message"], "TestServer => wrong message")

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
