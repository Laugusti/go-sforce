package restclient

import "reflect"
import "testing"

func TestGetField(t *testing.T) {
	tests := []struct {
		object    SObject
		field     string
		want      interface{}
		shouldErr bool
	}{
		{map[string]interface{}{"field1": 1, "field2": "two", "field3": false}, "field2", "two", false},
		{map[string]interface{}{"field1": 1, "field2": "two", "field3": false}, "field1", 1, false},
		{map[string]interface{}{"field1": 1, "field2": "two", "field3": false}, "field3", false, false},
		{map[string]interface{}{"field1": 1, "field2": "two", "field3": false}, "field4", nil, true},
		{map[string]interface{}{"field1": map[string]interface{}{"field2": "value"}}, "field1.field2", "value", false},
	}

	for _, test := range tests {
		got, err := test.object.GetField(test.field)
		if test.shouldErr && err == nil {
			t.Errorf("TestGetField => input %v: expected error", test)
		}
		if !test.shouldErr && err != nil {
			t.Errorf("TestGetField => input %v: unexpected error: %v", test, err)
		}
		if !reflect.DeepEqual(test.want, got) {
			t.Errorf("TestGetField => input %v: want %v, got %v", test, test.want, got)
		}
	}
}
