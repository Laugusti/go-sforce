package testserver

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

// ResponseHandler specfies an interface with a single method to handle http responses.
type ResponseHandler interface {
	Handle(*testing.T, http.ResponseWriter, string)
}

// JSONResponseHandler writes a status code and body to the response writer.
type JSONResponseHandler struct {
	StatusCode int
	Body       interface{}
}

// Handle implements the ResponseHandler interface.
func (h *JSONResponseHandler) Handle(t *testing.T, w http.ResponseWriter, assertMsg string) {
	// write response
	if h.Body == nil {
		w.WriteHeader(h.StatusCode)
		return
	}
	b, err := json.Marshal(h.Body)
	assert.Nil(t, err, assertMsg)
	w.WriteHeader(h.StatusCode)
	_, err = w.Write(b)
	w.Header().Set("Content-Type", "application/json")
	assert.Nil(t, err, assertMsg)
}
