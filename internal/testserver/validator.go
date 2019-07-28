package testserver

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

// RequestValidator is an interface with a single method to validate a http request
type RequestValidator interface {
	Validate(*testing.T, *http.Request)
}

// HeaderValidator validates headers on the request
type HeaderValidator struct {
	Key   string
	Value string
}

// Validate implements the RequestValidator interface
func (v *HeaderValidator) Validate(t *testing.T, r *http.Request) {
	got := r.Header.Get(v.Key)
	assert.Equal(t, v.Value, got)
}

// JSONBodyValidator validates the request body as JSON
type JSONBodyValidator struct {
	Body interface{}
}

// Validate implements the RequestValidator interface
func (v *JSONBodyValidator) Validate(t *testing.T, r *http.Request) {
	if v.Body == nil {
		assert.Empty(t, r.Body)
		return
	}
	//get expected body as map
	want, err := jsonObjectToMap(v.Body)
	assert.Nil(t, err)
	// get request body as map
	got := make(map[string]interface{})
	assert.Nil(t, json.NewDecoder(r.Body).Decode(&got))
	// assert body matches
	assert.Equal(t, want, got)
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
