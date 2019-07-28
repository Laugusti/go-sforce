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

	tests := []struct {
		body       interface{}
		statusCode int
	}{
		{nil, 204},
		{map[string]interface{}{"field1": "one", "field2": 2.0}, 200},
		{struct{ A string }{"reponse"}, 200},
	}
	for i, test := range tests {
		assertMsg := fmt.Sprintf("input: %v", test)
		// set server response to static json
		s.HandlerFunc = StaticJSONHandler(t, test.body, test.statusCode)

		// get response using http client
		resp, err := s.Client().Get(s.URL())
		if !assert.NoError(t, err, assertMsg) {
			t.Fatal("missing response")
		}
		defer func() { assert.Nil(t, resp.Body.Close(), assertMsg) }()

		// assert response body matches expected
		assert.Equal(t, i+1, s.RequestCount, assertMsg)
		assert.Equal(t, test.statusCode, resp.StatusCode, assertMsg)
		assertExepctedBody(t, test.body, resp, assertMsg)
	}
}

func TestValidateAndSetResponseHandler(t *testing.T) {
	// start server
	s := New(t)
	defer s.Stop()

	tests := []struct {
		reqBody    interface{}
		respBody   interface{}
		path       string
		rawHeaders string
		rawQuery   string
		statusCode int
	}{
		{nil, nil, "", "", "", 200},
		{map[string]interface{}{"A": "a", "B": 2.0}, nil, "", "", "", 200},
		{nil, nil, "api/path", "q=query", "A=a&Content-Type=application/json", 400},
		{nil, nil, "api/path", "q=query", "A=a&Content-Type=application/json", 400},
		{struct{ A string }{"request"}, struct{ A string }{"response"}, "path", "a=1,2&b=3,4,5", "a=1,2&b=3,4,5", 202},
	}

	for i, test := range tests {
		assertMsg := fmt.Sprintf("input: %v", test)

		query, err := url.ParseQuery(test.rawQuery)
		assert.Nil(t, err, assertMsg)
		headers, err := url.ParseQuery(test.rawHeaders)
		assert.Nil(t, err, assertMsg)

		var buf bytes.Buffer
		if test.reqBody != nil {
			err = json.NewEncoder(&buf).Encode(test.reqBody)
			assert.Nil(t, err, assertMsg)
		}
		req, err := http.NewRequest(http.MethodGet,
			fmt.Sprintf("%s/%s?%s", s.URL(), test.path, test.rawQuery),
			&buf)
		assert.Nil(t, err, assertMsg)

		validators := []RequestValidator{&JSONBodyValidator{test.reqBody},
			&PathValidator{test.path}, &QueryValidator{query}}
		for k := range headers {
			v := headers.Get(k)
			req.Header.Set(k, v)
			validators = append(validators, &HeaderValidator{k, v})
		}

		s.HandlerFunc = ValidateAndSetResponseHandler(t, assertMsg,
			test.respBody, test.statusCode, validators...)
		resp, err := s.Client().Do(req)
		if !assert.NoError(t, err, assertMsg) {
			t.Fatal("missing response")
		}
		defer func() { assert.Nil(t, resp.Body.Close(), assertMsg) }()

		assert.Equal(t, i+1, s.RequestCount, assertMsg)
		assert.Equal(t, test.statusCode, resp.StatusCode, assertMsg)
		assertExepctedBody(t, test.respBody, resp, assertMsg)

	}
}

func assertExepctedBody(t *testing.T, want interface{}, resp *http.Response, assertMsg string) {
	if want == nil {
		assert.Empty(t, resp.Body, assertMsg)
	} else {
		want, err := jsonObjectToMap(want)
		assert.Nil(t, err, assertMsg)
		got := make(map[string]interface{})
		err = json.NewDecoder(resp.Body).Decode(&got)
		assert.Nil(t, err, assertMsg)
		assert.Equal(t, want, got, assertMsg)
	}
}
