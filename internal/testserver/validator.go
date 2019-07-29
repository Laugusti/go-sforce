package testserver

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// RequestValidator is an interface with a single method to validate a http request.
type RequestValidator interface {
	Validate(*testing.T, *http.Request, string)
}

// HeaderValidator validates headers on the request.
type HeaderValidator struct {
	Key   string
	Value string
}

// Validate implements the RequestValidator interface.
func (v *HeaderValidator) Validate(t *testing.T, r *http.Request, assertMsg string) {
	assert.Equal(t, v.Value, r.Header.Get(v.Key), assertMsg)
}

// JSONBodyValidator validates the request body as JSON.
type JSONBodyValidator struct {
	Body interface{}
}

// Validate implements the RequestValidator interface.
func (v *JSONBodyValidator) Validate(t *testing.T, r *http.Request, assertMsg string) {
	if v.Body == nil {
		assert.Empty(t, r.Body, assertMsg)
		return
	}
	//get expected body as map
	want, err := jsonObjectToMap(v.Body)
	assert.Nil(t, err, assertMsg)
	// get request body as map
	got := make(map[string]interface{})
	assert.Nil(t, json.NewDecoder(r.Body).Decode(&got), assertMsg)
	// assert body matches
	assert.Equal(t, want, got, assertMsg)
}

// PathValidator validates the request path.
type PathValidator struct {
	Path string
}

// Validate implements the RequestValidator interface.
func (v *PathValidator) Validate(t *testing.T, r *http.Request, assertMsg string) {
	path := r.URL.Path
	if !strings.HasPrefix(v.Path, "/") {
		path = strings.TrimPrefix(r.URL.Path, "/")
	}
	assert.Equal(t, v.Path, path, assertMsg)
}

// QueryValidator validates the request query.
type QueryValidator struct {
	Query url.Values
}

// Validate implements the RequestValidator interface.
func (v *QueryValidator) Validate(t *testing.T, r *http.Request, assertMsg string) {
	assert.Equal(t, v.Query, r.URL.Query(), assertMsg)
}

// MethodValidator validates the request method.
type MethodValidator struct {
	Method string
}

// Validate implements the RequestValidator interface.
func (v *MethodValidator) Validate(t *testing.T, r *http.Request, assertMsg string) {
	assert.Equal(t, v.Method, r.Method, assertMsg)
}

// FormValidator validates the request form.
type FormValidator struct {
	Form url.Values
}

// Validate implements the RequestValidator interface.
func (v *FormValidator) Validate(t *testing.T, r *http.Request, assertMsg string) {
	err := r.ParseForm()
	assert.Nil(t, err, assertMsg)
	assert.Equal(t, v.Form, r.Form, assertMsg)
}

func jsonObjectToMap(v interface{}) (map[string]interface{}, error) {
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(v)
	if err != nil {
		return nil, err
	}
	m := make(map[string]interface{})
	err = json.NewDecoder(&buf).Decode(&m)
	return m, err
}
