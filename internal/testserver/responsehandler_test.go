package testserver

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

type responseWriter struct {
	statusCode int
	header     http.Header
	buf        *bytes.Buffer
}

func (w *responseWriter) Header() http.Header { return w.header }
func (w *responseWriter) WriteHeader(sc int)  { w.statusCode = sc }
func (w *responseWriter) Write(b []byte) (int, error) {
	n, err := io.Copy(w.buf, bytes.NewReader(b))
	return int(n), err
}
func newResponseWriter() *responseWriter {
	return &responseWriter{0, http.Header{}, &bytes.Buffer{}}
}

func TestJSONResponseHandler(t *testing.T) {
	tests := []struct {
		statusCode int
		body       interface{}
	}{
		{},
		{100, struct{ A string }{"1"}},
		{200, struct{ B int }{2.0}},
		{300, struct{ C bool }{true}},
		{-1, map[string]interface{}{"A": true, "B": "1", "C": 2.0}},
	}

	for _, test := range tests {
		assertMsg := fmt.Sprintf("input: %v", test)

		// test json handler
		h := &JSONResponseHandler{test.statusCode, test.body}
		assertResponseMatch(t, h, test.statusCode, test.body, assertMsg)
	}
}

func TestConsecutiveResponseHandler(t *testing.T) {
	tests := []struct {
		handlers []*JSONResponseHandler
	}{
		{[]*JSONResponseHandler{&JSONResponseHandler{100, struct{ A int }{1}}}},
		{[]*JSONResponseHandler{&JSONResponseHandler{100, struct{ A int }{1}}, &JSONResponseHandler{200, struct{ B int }{2}}}},
		{[]*JSONResponseHandler{&JSONResponseHandler{100, struct{ A int }{1}}, &JSONResponseHandler{200, struct{ B int }{2}}, &JSONResponseHandler{300, struct{ C int }{3}}}},
	}
	for _, test := range tests {
		assertMsg := fmt.Sprintf("input: %v", test)

		if !assert.Greater(t, len(test.handlers), 0, assertMsg) {
			continue
		}
		// handler
		handlers := []ResponseHandler{}
		for _, h := range test.handlers {
			handlers = append(handlers, h)
		}
		h := &ConsecutiveResponseHandler{Handlers: handlers}

		// iterate through handlers and check response
		for _, current := range test.handlers {
			// test concurrent handler
			assertResponseMatch(t, h, current.StatusCode, current.Body, assertMsg)
		}
		// more requests should use final handler
		last := test.handlers[len(test.handlers)-1]
		for i := 0; i < 5; i++ {
			// test concurrent handler
			assertResponseMatch(t, h, last.StatusCode, last.Body, assertMsg)
		}
	}
}

func assertResponseMatch(t *testing.T, h ResponseHandler, statusCode int, body interface{},
	assertMsg string) {
	// handle with test response writer
	w := newResponseWriter()
	h.Handle(t, w, assertMsg)

	// status code match
	assert.Equal(t, statusCode, w.statusCode)
	// bodies match
	want, err := jsonObjectToMap(body)
	assert.Nil(t, err, assertMsg)
	got, err := jsonReaderToMap(w.buf)
	assert.Nil(t, err, assertMsg)
	assert.Equal(t, want, got, assertMsg)
}
