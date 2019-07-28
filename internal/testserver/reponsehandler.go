package testserver

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

type ResponseHandler interface {
	Handle(*testing.T, http.ResponseWriter, string)
}

type JSONBodyResponseHandler struct {
	StatusCode int
	Body       interface{}
}

func (h *JSONBodyResponseHandler) Handle(t *testing.T, w http.ResponseWriter, assertMsg string) {
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
