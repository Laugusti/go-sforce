package restclient

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetField(t *testing.T) {
	tests := []struct {
		object    SObject
		field     string
		want      interface{}
		shouldErr bool
	}{
		{nil, "", nil, true},
		{nil, "a", nil, true},
		{map[string]interface{}{}, "", nil, true},
		{map[string]interface{}{}, "a", nil, true},
		{map[string]interface{}{"field1": 1, "field2": "two", "field3": false}, "field2", "two", false},
		{map[string]interface{}{"field1": 1, "field2": "two", "field3": false}, "field1", 1, false},
		{map[string]interface{}{"field1": 1, "field2": "two", "field3": false}, "field3", false, false},
		{map[string]interface{}{"field1": 1, "field2": "two", "field3": false}, "field4", nil, true},
		{map[string]interface{}{"field1": map[string]interface{}{"field2": "value"}}, "field1.field2", "value", false},
		{map[string]interface{}{"field1": map[string]interface{}{"field2": "value"}}, "field1.field3", "value", true},
	}

	for _, test := range tests {
		got, err := test.object.GetField(test.field)
		if test.shouldErr {
			assert.NotNilf(t, err, "input :%v", test)
		} else {
			assert.Nil(t, err, "input: %v", test)
			assert.Equal(t, test.want, got, "input: %v", test)
		}
	}
}
