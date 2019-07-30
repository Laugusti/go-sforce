package testserver

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJsonObjectToReadCloser(t *testing.T) {
	tests := []struct {
		object interface{}
		want   string
	}{
		{},
		{"", `""`},
		{1.0, "1"},
		{true, "true"},
		{struct{}{}, "{}"},
		{map[string]interface{}{}, "{}"},
		{struct{ A string }{"1"}, `{"A":"1"}`},
		{struct{ B int }{2}, `{"B":2}`},
		{struct{ C bool }{true}, `{"C":true}`},
		{map[string]interface{}{"A": "1", "B": 2.0, "C": true}, `{"A":"1","B":2,"C":true}`},
	}

	for _, test := range tests {
		assertMsg := fmt.Sprintf("input: %v", test)
		r, err := jsonObjectToReadCloser(test.object)
		assert.Nil(t, err, assertMsg)
		buf := &bytes.Buffer{}
		if r != nil && assert.NotNil(t, r, assertMsg) {
			_, err = io.Copy(buf, r)
			assert.Nil(t, err, assertMsg)
		} else {
			assert.Nil(t, r, assertMsg)
		}
		got := strings.TrimSuffix(buf.String(), "\n")
		assert.Equal(t, test.want, got, assertMsg)
	}
}
func TestJsonObjectToMap(t *testing.T) {
	tests := []struct {
		shouldErr bool
		object    interface{}
		want      map[string]interface{}
	}{
		{},
		{false, struct{ A string }{"value"}, map[string]interface{}{"A": "value"}},
		{false, struct{ B int }{3}, map[string]interface{}{"B": 3.0}},
		{false, struct{ C bool }{true}, map[string]interface{}{"C": true}},
		{true, "", nil},
		{true, "{}", nil},
		{true, "1", nil},
		{true, 1.0, nil},
		{true, false, nil},
	}

	for _, test := range tests {
		assertMsg := fmt.Sprintf("input: %v", test)
		got, err := jsonObjectToMap(test.object)
		if test.shouldErr {
			assert.NotNil(t, err, assertMsg)
			assert.Nil(t, got, assertMsg)
		} else {
			assert.Nil(t, err, assertMsg)
			assert.Equal(t, test.want, got, assertMsg)
		}
	}
}

func TestJsonReaderToMap(t *testing.T) {
	tests := []struct {
		shouldErr bool
		input     string
		want      map[string]interface{}
	}{
		{},
		{false, "{}", map[string]interface{}{}},
		{false, `{"A":"1","B":2,"C":true}`, map[string]interface{}{"A": "1", "B": 2.0, "C": true}},
		{true, "fail", nil},
	}

	for _, test := range tests {
		assertMsg := fmt.Sprintf("input: %v", test)
		got, err := jsonReaderToMap(strings.NewReader(test.input))
		if test.shouldErr {
			assert.NotNil(t, err, assertMsg)
			assert.Nil(t, got, assertMsg)
		} else {
			assert.Nil(t, err, assertMsg)
			assert.Equal(t, test.want, got, assertMsg)
		}
	}
}
