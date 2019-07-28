package testserver

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/textproto"
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
		v.Validate(t, &req)
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
		v.Validate(t, &req)
	}
}
