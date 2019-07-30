package testserver

import (
	"errors"
	"io"
	"net/http"
)

// ResponseHandler specfies an interface with a single method to handle http responses.
type ResponseHandler interface {
	Handle(http.ResponseWriter) error
}

// JSONResponseHandler writes a status code and body to the response writer.
type JSONResponseHandler struct {
	StatusCode int
	Body       interface{}
}

// Handle implements the ResponseHandler interface.
func (h *JSONResponseHandler) Handle(w http.ResponseWriter) error {
	// write response
	if h.Body == nil {
		w.WriteHeader(h.StatusCode)
		return nil
	}
	body, err := jsonObjectToReadCloser(h.Body)
	if err != nil {
		return err
	}
	w.WriteHeader(h.StatusCode)
	w.Header().Set("Content-Type", "application/json")
	if _, err := io.Copy(w, body); err != nil {
		return err
	}
	return nil
}

// ConsecutiveResponseHandler allows for consective handler calls. The last handler determines the behavior of further calls
type ConsecutiveResponseHandler struct {
	Handlers     []ResponseHandler
	HandledCount int
}

// Handle implements the ResponseHandler interface.
func (h *ConsecutiveResponseHandler) Handle(w http.ResponseWriter) error {
	if len(h.Handlers) < 1 {
		return errors.New("need at least 1 handler")
	}
	if h.HandledCount < 0 {
		return errors.New("cannot have negative handled count")
	}
	// default to last handler
	handler := h.Handlers[len(h.Handlers)-1]
	// use handler at count if less than size of handlers
	if h.HandledCount < len(h.Handlers) {
		handler = h.Handlers[h.HandledCount]
	}
	h.HandledCount++
	return handler.Handle(w)
}
