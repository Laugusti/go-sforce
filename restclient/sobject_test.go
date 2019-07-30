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
		{false, map[string]interface{}{}, "a", nil, map[string]interface{}{"a": nil}},
		{false, map[string]interface{}{"a": 1}, "a", nil, map[string]interface{}{"a": nil}},
		{false, map[string]interface{}{"a": 1}, "b", 2, map[string]interface{}{"a": 1, "b": 2}},
		{false, map[string]interface{}{"a": 1}, "a.b", nil, map[string]interface{}{"a": 1, "a.b": nil}},
		{false, map[string]interface{}{"a": map[string]interface{}{"b": "1"}}, "a.b", "c", map[string]interface{}{"a": map[string]interface{}{"b": "1"}, "a.b": "c"}},
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
		{true, map[string]interface{}{}, "", nil},
		{true, map[string]interface{}{}, "a", nil},
		{false, map[string]interface{}{"a": 14}, "a", 14},
		{false, map[string]interface{}{"a.b": 14}, "a.b", 14},
		{false, map[string]interface{}{"a.b": 14}, "a.b", 14},
		{true, map[string]interface{}{"a": map[string]interface{}{"b": "value"}}, "a.b", nil},
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

func TestSObjectBuilder(t *testing.T) {
	sb := NewSObjectBuilder()
	assert.NotNil(t, sb, "want sobject builder")
	got := sb.NewField("a", "1").NewDottedField("a.b.c", true).NewDottedField("a.b.d", 19.3).
		NewField("a.b", "value").Build()
	var want SObject = map[string]interface{}{
		"a.b": "value",
		"a": map[string]interface{}{
			"b": map[string]interface{}{
				"c": true,
				"d": 19.3,
			},
		},
	}

	assert.Equal(t, want, got)
}

func TestFromSObject(t *testing.T) {
	rng := rand.New(rand.NewSource(time.Now().UTC().UnixNano()))
	for i := 0; i < 100; i++ {
		sobj := randomSObject(rng)
		assert.Equal(t, sobj, SObjectBuilderFromSObject(sobj).Build())
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
	return sb.Build()
}
