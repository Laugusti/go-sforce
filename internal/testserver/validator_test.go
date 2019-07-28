package testserver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
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
		v.Validate(t, &req, fmt.Sprintf("input: %v", test))
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
		v := &JSONBodyValidator{test.body}
		var req http.Request
		if test.body != nil {
			var buf bytes.Buffer
			err := json.NewEncoder(&buf).Encode(test.body)
			assert.Nil(t, err)
			req.Body = ioutil.NopCloser(&buf)
		}
		v.Validate(t, &req, fmt.Sprintf("input: %v", test))
	}
}

func TestPathValidator(t *testing.T) {
	tests := []struct {
		path string
		url  string
	}{
		{"", "http://localhost"},
		{"/", "http://localhost/"},
		{"/a", "http://localhost/a"},
		{"/a/b/c", "http://localhost/a/b/c"},
		{"/a/", "http://localhost/a/"},
		{"/a", "http://localhost/a?b=c"},
	}

	for _, test := range tests {
		u, err := url.Parse(test.url)
		assert.Nil(t, err)
		var req http.Request
		req.URL = u
		v := &PathValidator{test.path}
		v.Validate(t, &req, fmt.Sprintf("input: %v", test))
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
		u, err := url.Parse(test.url)
		assert.Nil(t, err)
		var req http.Request
		req.URL = u
		q, err := url.ParseQuery(test.query)
		assert.Nil(t, err)
		v := &QueryValidator{q}
		v.Validate(t, &req, fmt.Sprintf("input: %v", test))
	}
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
