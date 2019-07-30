package restclient

import (
	"errors"
	"fmt"
	"strings"
)

var (
	errInvalidField = errors.New("field name is required")
)

// SObjectField is represents a field on a Salesforce Object. This is used by the
// Salesforce Object Builder to add fields to the Salesforce Object.
type SObjectField struct {
	Name   string
	Value  interface{}
	Dotted bool
}

// SObjectBuilder is used to build a Salesforce Object using fluent chainging.
type SObjectBuilder struct {
	Fields []*SObjectField
}

// NewSObjectBuilder returns an builder for the Salesforce Object
func NewSObjectBuilder() *SObjectBuilder {
	return &SObjectBuilder{}
}

// NewField adds the field using the AddField method and returns the SObjectBuilder
// for fluent chaining.
func (sb *SObjectBuilder) NewField(field string, value interface{}) *SObjectBuilder {
	sb.Fields = append(sb.Fields, &SObjectField{field, value, false})
	return sb
}

// NewDottedField adds the field using the AddDottedField method and returns the SObjectBuilder
// for fluent chaining.
func (sb *SObjectBuilder) NewDottedField(field string, value interface{}) *SObjectBuilder {
	sb.Fields = append(sb.Fields, &SObjectField{field, value, true})
	return sb
}

// Build builds the Salesforce Object.
func (sb *SObjectBuilder) Build() SObject {
	s := SObject{}
	for _, f := range sb.Fields {
		if f.Dotted {
			if _, err := getLastSObjectInPath(s, f.Name, true, true); err != nil {
				panic(err)
			}
			if err := s.AddDottedField(f.Name, f.Value); err != nil {
				panic(err)
			}
		} else {
			if err := s.AddField(f.Name, f.Value); err != nil {
				panic(err)
			}
		}
	}
	return s
}

// SObject stores a Salesforce Object used by the Salesforce API.
type SObject map[string]interface{}

// AddField adds the field to the sobject.
func (s SObject) AddField(field string, v interface{}) error {
	if field == "" {
		return errInvalidField
	}
	s[field] = v
	return nil
}

func (s SObject) mandatoryAddField(field string, v interface{}) {
	if err := s.AddField(field, v); err != nil {
		panic(err)
	}
}

// AddDottedField adds the field to the Salesforce Object.
// The difference from AddField is this will traverse down the
// object (creating paths if necessary) to add the field.
func (s SObject) AddDottedField(dottedField string, v interface{}) error {
	sobj, err := getLastSObjectInPath(s, dottedField, true, false)
	if err != nil {
		return err
	}

	// add last field in fields list to working sobject
	sobj.mandatoryAddField(getLastFieldInPath(dottedField), v)
	return nil
}

// GetField returns the field value on the Salesforce Object.
func (s SObject) GetField(field string) (interface{}, error) {
	if field == "" {
		return nil, errInvalidField
	}
	v, ok := s[field]
	if !ok {
		return nil, fmt.Errorf("field %q not in salesforce object", field)
	}
	return v, nil
}

// GetDottedField returns the field value on the Salesforce Object.
// The difference from GetField is this will traverse down the
// object to retrieve the field.
// Example s.GetDottedField("Contact.Name"
func (s SObject) GetDottedField(dottedField string) (interface{}, error) {
	sobj, err := getLastSObjectInPath(s, dottedField, false, false)
	if err != nil {
		return nil, err
	}
	return sobj.GetField(getLastFieldInPath(dottedField))
}

// GetMandatoryField ensures the field is valid for the Salesforce Object and there was no error
// when calling the GetField function.
func (s SObject) GetMandatoryField(field string) interface{} {
	v, err := s.GetField(field)
	if err != nil {
		panic(err)
	}
	return v
}

func getLastFieldInPath(dottedField string) string {
	// split field param using "."
	fields := strings.Split(dottedField, ".")
	return fields[len(fields)-1]
}

func getLastSObjectInPath(s SObject, dottedField string, createIfNotExists, replaceIfWrongType bool) (SObject, error) {
	// split field param using "."
	fields := strings.Split(dottedField, ".")
	// validate all fields in path
	for _, f := range fields {
		if f == "" {
			return nil, errInvalidField
		}
	}
	// drill down map object to get value
	var sobj = s
	for i := 0; i < len(fields)-1; i++ {
		field := fields[i]
		v, err := sobj.GetField(field)
		if err != nil {
			// field not found, add it
			if createIfNotExists {
				sobj.mandatoryAddField(field, map[string]interface{}{})
			} else {
				return nil, err
			}
		} else {
			// found field, value must be map
			if _, ok := v.(map[string]interface{}); !ok && !replaceIfWrongType {
				return nil, fmt.Errorf("failed to get field %q: want map got %T",
					dottedField, v)
			} else if !ok && replaceIfWrongType {
				sobj.mandatoryAddField(field, map[string]interface{}{})
			}
		}
		sobj = sobj.GetMandatoryField(field).(map[string]interface{})
	}
	return sobj, nil
}
