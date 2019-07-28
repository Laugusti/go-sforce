package testserver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStaticJSONHandler(t *testing.T) {
	// start server
	s := New(t)
	defer s.Stop()

	// set server response to static json
	want := map[string]interface{}{
		"field1": "one",
		"field2": 2.0,
	}
	s.HandlerFunc = StaticJSONHandler(t, want, http.StatusCreated)

	// get response using http client
	resp, err := s.Client().Get(s.URL())
	assert.Nil(t, err)
	defer func() { _ = resp.Body.Close() }()

	// assert response body matches expected
	assert.Equal(t, resp.StatusCode, http.StatusCreated)
	got := make(map[string]interface{})
	err = json.NewDecoder(resp.Body).Decode(&got)
	assert.Nil(t, err)
	assert.Equal(t, want, got)
}

func TestValidateAndSetResponseHandler(t *testing.T) {
	// start server
	s := New(t)
	defer s.Stop()

	// set server handler
	reqBody := map[string]interface{}{
		"field1": "one",
		"field2": 2.0,
	}
	respBody := map[string]interface{}{
		"field2": "one",
		"field1": 2.0,
	}

	headerValidator1 := &HeaderValidator{"Key", "Value"}
	headerValidator2 := &HeaderValidator{"Content-Type", "application/json"}
	bodyValidator := &JSONBodyValidator{reqBody}
	pathValidator := &PathValidator{"/path/to/resource"}
	q, err := url.ParseQuery("q=query&s=test")
	assert.Nil(t, err)
	queryValidator := &QueryValidator{q}

	s.HandlerFunc = ValidateAndSetResponseHandler(t, fmt.Sprintf("req: %v, resp: %v", reqBody, respBody), respBody, http.StatusAccepted,
		headerValidator1, headerValidator2, bodyValidator, pathValidator, queryValidator)

	// get response using http client
	var buf bytes.Buffer
	err = json.NewEncoder(&buf).Encode(reqBody)
	assert.Nil(t, err)
	req, err := http.NewRequest(http.MethodPost, s.URL()+"/path/to/resource?q=query&s=test", &buf)
	assert.Nil(t, err)
	req.Header.Set("Key", "Value")
	req.Header.Set("Content-Type", "application/json")
	resp, err := s.Client().Do(req)
	if !assert.NoError(t, err) {
		t.Fatal("need response")
	}
	defer func() { _ = resp.Body.Close() }()

	// assert response body matches expected
	assert.Equal(t, resp.StatusCode, http.StatusAccepted)
	got := make(map[string]interface{})
	err = json.NewDecoder(resp.Body).Decode(&got)
	assert.Nil(t, err)
	assert.Equal(t, respBody, got)
}

func TestValidateJSONBodyHandler(t *testing.T) {
	// start server
	s := New(t)
	defer s.Stop()

	// set server handler
	reqBody := map[string]interface{}{
		"field1": "one",
		"field2": 2.0,
	}
	respBody := map[string]interface{}{
		"field2": "one",
		"field1": 2.0,
	}
	s.HandlerFunc = ValidateJSONBodyHandler(t, reqBody, respBody, http.StatusAccepted, "wrong body")

	// get response using http client
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(reqBody)
	assert.Nil(t, err)
	resp, err := s.Client().Post(s.URL(), "application/json", &buf)
	assert.Nil(t, err)
	defer func() { _ = resp.Body.Close() }()

	// assert response body matches expected
	assert.Equal(t, resp.StatusCode, http.StatusAccepted)
	got := make(map[string]interface{})
	err = json.NewDecoder(resp.Body).Decode(&got)
	assert.Nil(t, err)
	assert.Equal(t, respBody, got)
}

func TestJSONObjectToMap(t *testing.T) {
	tests := []struct {
		object interface{}
		want   map[string]interface{}
	}{
		{struct{ A string }{"value"}, map[string]interface{}{"A": "value"}},
		{struct{ B int }{3}, map[string]interface{}{"B": 3.0}},
		{struct{ C bool }{true}, map[string]interface{}{"C": true}},
	}

	for _, test := range tests {
		got, err := jsonObjectToMap(test.object)
		assert.Nil(t, err)
		assert.Equal(t, test.want, got)
	}
}
