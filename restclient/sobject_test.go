package restclient

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAddField(t *testing.T) {
	tests := []struct {
		shouldErr bool
		object    SObject
		field     string
		value     interface{}
		want      SObject
	}{
		{true, nil, "", nil, nil},
		{true, map[string]interface{}{}, "a.,b", nil, nil},
		{true, map[string]interface{}{}, "a%b", nil, nil},
		{true, map[string]interface{}(nil), "a", nil, nil},
		{false, map[string]interface{}{}, "a", nil, map[string]interface{}{"a": nil}},
		{false, map[string]interface{}{"a": 1}, "a", nil, map[string]interface{}{"a": nil}},
		{false, map[string]interface{}{"a": 1}, "b", 2, map[string]interface{}{"a": 1, "b": 2}},
		{false, map[string]interface{}{"a": map[string]interface{}{"b": "1"}}, "a", "b", map[string]interface{}{"a": "b"}},
	}

	for _, test := range tests {
		assertMsg := fmt.Sprintf("input :%v", test)
		err := test.object.AddField(test.field, test.value)
		if test.shouldErr {
			assert.NotNil(t, err, assertMsg)
		} else {
			assert.Nil(t, err, assertMsg)
			assert.Equal(t, test.want, test.object, assertMsg)
		}
	}
}

func TestAddDottedField(t *testing.T) {
	tests := []struct {
		shouldErr bool
		object    SObject
		field     string
		value     interface{}
		want      SObject
	}{
		{true, nil, "", nil, nil},
		{false, map[string]interface{}{}, "a", nil, map[string]interface{}{"a": nil}},
		{false, map[string]interface{}{"a": 1}, "a", nil, map[string]interface{}{"a": nil}},
		{false, map[string]interface{}{"a": 1}, "b", 2, map[string]interface{}{"a": 1, "b": 2}},
		{true, map[string]interface{}{"a": 1}, "a.b", nil, nil},
		{false, map[string]interface{}{"a": map[string]interface{}{"b": "1"}}, "a.b", "c", map[string]interface{}{"a": map[string]interface{}{"b": "c"}}},
		{false, map[string]interface{}{"a": map[string]interface{}{"b": "1"}}, "a", "b", map[string]interface{}{"a": "b"}},
		{false, map[string]interface{}{"a.b": 1}, "a.b", 2, map[string]interface{}{"a": map[string]interface{}{"b": 2}, "a.b": 1}},
		{false, map[string]interface{}{}, "a.b.c", true, map[string]interface{}{"a": map[string]interface{}{"b": map[string]interface{}{"c": true}}}},
		{false, map[string]interface{}{"a": map[string]interface{}(nil)}, "a.b", 1, map[string]interface{}{"a": map[string]interface{}{"b": 1}}},
	}

	for _, test := range tests {
		assertMsg := fmt.Sprintf("input :%v", test)
		err := test.object.AddDottedField(test.field, test.value)
		if test.shouldErr {
			assert.NotNil(t, err, assertMsg)
		} else {
			assert.NoError(t, err, assertMsg)
			assert.Equal(t, test.want, test.object, assertMsg)
		}
	}
}

func TestGetField(t *testing.T) {
	tests := []struct {
		shouldErr bool
		object    SObject
		field     string
		want      interface{}
	}{
		{true, nil, "", nil},
		{true, nil, "a", nil},
		{true, map[string]interface{}{"a.b": 14}, "a.b", nil},
		{true, map[string]interface{}{"a": map[string]interface{}{"b": "value"}}, "a.b", nil},
		{true, map[string]interface{}{"a.b": 14}, "a__b", nil},
		{true, map[string]interface{}{}, "", nil},
		{true, map[string]interface{}{}, "a", nil},
		{false, map[string]interface{}{"a": 14}, "a", 14},
	}

	for _, test := range tests {
		assertMsg := fmt.Sprintf("input :%v", test)
		got, err := test.object.GetField(test.field)
		if test.shouldErr {
			assert.NotNil(t, err, assertMsg)
		} else {
			assert.Nil(t, err, assertMsg)
			assert.Equal(t, test.want, got, assertMsg)
		}
	}
}

func TestGetDottedField(t *testing.T) {
	tests := []struct {
		shouldErr bool
		object    SObject
		field     string
		want      interface{}
	}{
		{true, nil, "", nil},
		{true, nil, "a", nil},
		{true, map[string]interface{}{}, "", nil},
		{true, map[string]interface{}{}, "a", nil},
		{true, map[string]interface{}{"a": 1}, "a.b", 1},
		{true, map[string]interface{}{"a.b": 1}, "a.b", 1},
		{false, map[string]interface{}{"a": 1, "b": "two", "c": false}, "b", "two"},
		{false, map[string]interface{}{"a": 1, "b": "two", "c": false}, "a", 1},
		{false, map[string]interface{}{"a": 1, "b": "two", "c": false}, "c", false},
		{true, map[string]interface{}{"a": 1, "b": "two", "c": false}, "d", nil},
		{false, map[string]interface{}{"a": map[string]interface{}{"b": "value"}}, "a.b", "value"},
		{true, map[string]interface{}{"a": map[string]interface{}{"b": "value"}}, "a.", "value"},
		{true, map[string]interface{}{"a": map[string]interface{}{"b": "value"}}, "a.c", "value"},
	}

	for _, test := range tests {
		assertMsg := fmt.Sprintf("input :%v", test)
		got, err := test.object.GetDottedField(test.field)
		if test.shouldErr {
			assert.NotNil(t, err, assertMsg)
		} else {
			assert.Nil(t, err, assertMsg)
			assert.Equal(t, test.want, got, assertMsg)
		}
	}
}

func TestBuild(t *testing.T) {
	tests := []struct {
		shouldErr bool
		fields    []SObjectBuilderField
		want      SObject
	}{
		{true, []SObjectBuilderField{SObjectBuilderField{Name: ""}}, nil},
		{false, []SObjectBuilderField{SObjectBuilderField{Name: "a"}}, map[string]interface{}{"a": nil}},
		{false, []SObjectBuilderField{SObjectBuilderField{Name: "a", Value: 1}}, map[string]interface{}{"a": 1}},
		{true, []SObjectBuilderField{SObjectBuilderField{Name: "a."}}, nil},
		{true, []SObjectBuilderField{SObjectBuilderField{Name: "a.", Dotted: true}}, nil},
		{true, []SObjectBuilderField{SObjectBuilderField{Name: "a.b.", Dotted: true}}, nil},
		{true, []SObjectBuilderField{SObjectBuilderField{Name: "a."}, SObjectBuilderField{Name: "a.", Dotted: true}}, nil},
		{false, []SObjectBuilderField{SObjectBuilderField{Name: "a.b.c.d", Dotted: true, Value: "e"}}, map[string]interface{}{"a": map[string]interface{}{"b": map[string]interface{}{"c": map[string]interface{}{"d": "e"}}}}},
		{true, []SObjectBuilderField{SObjectBuilderField{Name: "a.b..d", Dotted: true, Value: "e"}}, nil},
		{false, []SObjectBuilderField{SObjectBuilderField{Name: "a", Value: map[string]interface{}(nil)}, SObjectBuilderField{Name: "a.b", Dotted: true}}, map[string]interface{}{"a": map[string]interface{}{"b": nil}}},
	}

	for _, test := range tests {
		assertMsg := fmt.Sprintf("input: %v", test)
		fields := make([]*SObjectBuilderField, 0, len(test.fields))
		for _, f := range test.fields {
			fields = append(fields, &f)
		}
		sb := &SObjectBuilder{fields}
		got, err := sb.Build()
		if test.shouldErr {
			assert.NotNil(t, err, assertMsg)
			assert.Nil(t, got, assertMsg)
		} else {
			assert.Nil(t, err, assertMsg)
			assert.Equal(t, test.want, got, assertMsg)
		}
	}
}

func TestInvalidFields(t *testing.T) {
	tests := []struct {
		shouldErr bool
		name      string
		dotted    bool
	}{
		{true, "", false},
		{true, "", true},
		{false, "a_b", false},
		{false, "a.b", true},
		{true, "a.b", false},
		{true, "a-b", false},
		{true, "a!b", false},
		{true, "a%b", false},
		{true, "a()b", false},
		{true, "0", false},
		{true, "0a", false},
		{false, "a0", false},
		{false, "a_0", false},
		{false, "a.b.c.d.e", true},
		{true, "a.b.c.d.e_", true},
		{false, "a.b__c", true},
		{false, "a.b__r", true},
		{false, "a.b.c.d__r.e", true},
		{true, "a.b__d", true},
	}

	for _, test := range tests {
		assertMsg := fmt.Sprintf("input :%v", test)
		sb := &SObjectBuilder{[]*SObjectBuilderField{&SObjectBuilderField{test.name, nil, test.dotted}}}
		_, err := sb.Build()
		if test.shouldErr {
			err, _ := err.(*InvalidFieldError)
			assert.NotNil(t, err, assertMsg)
		} else {
			assert.Nil(t, err, assertMsg)
		}
	}
}

func TestFromSObject(t *testing.T) {
	rng := rand.New(rand.NewSource(time.Now().UTC().UnixNano()))
	for i := 0; i < 100; i++ {
		sobj := randomSObject(rng)
		assert.Equal(t, sobj, SObjectBuilderFromSObject(sobj).MustBuild())
	}
}

func randomSObject(rng *rand.Rand) SObject {
	sb := NewSObjectBuilder()
	for i := 0; i < 3125; i++ {
		// create random combinations of [a b c d e]
		s := [5]string{}
		for i := 0; i < 5; i++ {
			s[i] = fmt.Sprintf("%c", 97+rng.Intn(5))
		}
		path := strings.Join(s[:], ".")
		sb.NewDottedField(path, path)
	}
	return sb.MustBuild()
}
