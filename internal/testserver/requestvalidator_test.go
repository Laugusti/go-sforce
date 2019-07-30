package testserver

import (
	"fmt"
	"net/http"
	"net/textproto"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHeaderValidator(t *testing.T) {
	tests := []struct {
		key     string
		value   string
		headers map[string]string
	}{
		{"", "", nil},
		{"", "", map[string]string{}},
		{"k", "v", map[string]string{"a": "1", "k": "v"}},
	}

	for _, test := range tests {
		v := &HeaderValidator{test.key, test.value}
		var req http.Request
		req.Header = make(map[string][]string)
		for k, v := range test.headers {
			req.Header[textproto.CanonicalMIMEHeaderKey(k)] = []string{v}
		}
		assert.Nil(t, v.Validate(&req), fmt.Sprintf("input: %v", test))
	}
}

func TestJSONBodyValidator(t *testing.T) {
	tests := []struct {
		body interface{}
	}{
		{nil},
		{map[string]string(nil)},
		{map[string]string{}},
		{map[string]string{"a": "1", "b": "2"}},
	}

	for _, test := range tests {
		assertMsg := fmt.Sprintf("input: %v", test)
		v := &JSONBodyValidator{test.body}
		var req http.Request
		if test.body != nil {
			body, err := jsonObjectToReadCloser(test.body)
			assert.Nil(t, err, assertMsg)
			req.Body = body
		}
		assert.Nil(t, v.Validate(&req), assertMsg)
	}
}

func TestPathValidator(t *testing.T) {
	tests := []struct {
		path string
		url  string
	}{
		{"", "http://localhost"},
		{"/", "http://localhost/"},
		{"", "http://localhost/"},
		{"a", "http://localhost/a"},
		{"a/", "http://localhost/a/"},
		{"a/b/c", "http://localhost/a/b/c"},
		{"a", "http://localhost/a?b=c"},
	}

	for _, test := range tests {
		assertMsg := fmt.Sprintf("input: %v", test)
		u, err := url.Parse(test.url)
		assert.Nil(t, err, assertMsg)
		var req http.Request
		req.URL = u
		v := &PathValidator{test.path}
		assert.Nil(t, v.Validate(&req), fmt.Sprintf("input: %v", test))
	}
}

func TestQueryValidator(t *testing.T) {
	tests := []struct {
		query string
		url   string
	}{
		{"", "http://localhost"},
		{"", "http://localhost/"},
		{"a=b", "http://localhost?a=b"},
		{"a=b", "http://localhost/a/b/c/?a=b"},
		{"a=b,c&d=e", "http://localhost/a/b/c/?a=b,c&d=e"},
	}

	for _, test := range tests {
		assertMsg := fmt.Sprintf("input: %v", test)
		u, err := url.Parse(test.url)
		assert.Nil(t, err, assertMsg)
		var req http.Request
		req.URL = u
		q, err := url.ParseQuery(test.query)
		assert.Nil(t, err, assertMsg)
		v := &QueryValidator{q}
		assert.Nil(t, v.Validate(&req), fmt.Sprintf("input: %v", test))
	}
}

func TestMethodValidator(t *testing.T) {
	tests := []struct {
		method string
	}{
		{""},
		{http.MethodGet},
		{http.MethodPost},
		{http.MethodDelete},
	}

	for _, test := range tests {
		var req http.Request
		req.Method = test.method
		v := &MethodValidator{test.method}
		assert.Nil(t, v.Validate(&req), fmt.Sprintf("input: %v", test))
	}
}

func TestFormValidator(t *testing.T) {
	tests := []struct {
		rawForm string
	}{
		{""},
		{"&"},
		{"="},
		{"a"},
		{"a=b"},
		{"a=b,c"},
		{"a=b&c"},
		{"a=b&c=d"},
		{"a=b&c=d,e"},
	}

	for _, test := range tests {
		assertMsg := fmt.Sprintf("input: %v", test)
		var req http.Request
		form, err := url.ParseQuery(test.rawForm)
		assert.Nil(t, err, assertMsg)
		req.Form = form
		v := &FormValidator{form}
		assert.Nil(t, v.Validate(&req), assertMsg)
	}
}
