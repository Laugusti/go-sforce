package testserver

import (
	"io"
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
	body, err := jsonObjectToReadCloser(h.Body)
	assert.Nil(t, err, assertMsg)
	w.WriteHeader(h.StatusCode)
	w.Header().Set("Content-Type", "application/json")
	_, err = io.Copy(w, body)
	assert.Nil(t, err, assertMsg)
}

// ConsecutiveResponseHandler allows for consective handler calls. The last handler determines the behavior of further calls
type ConsecutiveResponseHandler struct {
	Handlers []ResponseHandler
	Count    int
}

// Handle implements the ResponseHandler interface.
func (h *ConsecutiveResponseHandler) Handle(t *testing.T, w http.ResponseWriter, assertMsg string) {
	// use handler at count
	if h.Count >= 0 && h.Count < len(h.Handlers) {
		h.Handlers[h.Count].Handle(t, w, assertMsg)
		h.Count++
		return
	}
	// use last handler
	if len(h.Handlers) > 0 {
		h.Handlers[len(h.Handlers)-1].Handle(t, w, assertMsg)
		h.Count++
		return
	}
	assert.GreaterOrEqual(t, h.Count, 0, assertMsg, "negative count")
	assert.Greater(t, len(h.Handlers), 0, assertMsg, "no handlers")
}
