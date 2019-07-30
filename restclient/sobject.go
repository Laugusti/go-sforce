package restclient

import (
	"fmt"
	"strings"
)

var (
	errInvalidField = &InvalidFieldErr{}
)

// InvalidFieldErr represents an invalid field error.
type InvalidFieldErr struct {
	FieldName string
}

func (e *InvalidFieldErr) Error() string {
	return "field name is required"
}

// FieldNotFoundErr represents an error for field not in the object.
type FieldNotFoundErr struct {
	FieldName string
}

func (e *FieldNotFoundErr) Error() string {
	return fmt.Sprintf("field %q not in salesforce object", e.FieldName)
}

// InvalidFieldTypeErr represents an invalid field type when traversing the object.
type InvalidFieldTypeErr struct {
	FullFieldName   string
	FailedFieldName string
	FailedField     interface{}
}

func (e *InvalidFieldTypeErr) Error() string {
	return fmt.Sprintf("failed to get field %q at %q: want map got %T", e.FullFieldName, e.FailedFieldName, e.FailedField)
}

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

// NewSObjectBuilder returns a builder for a Salesforce Object.
func NewSObjectBuilder() *SObjectBuilder {
	return &SObjectBuilder{}
}

// SObjectBuilderFromSObject returns a builder for the SObject.
func SObjectBuilderFromSObject(s SObject) *SObjectBuilder {
	fields := []*SObjectField{}

	// new struct for sobjects remember the field path
	type nestedSObject struct {
		fieldPath string
		sobj      SObject
	}
	//
	items := []*nestedSObject{&nestedSObject{"", s}}
	for i := 0; i < len(items); i++ {
		item := items[i]
		for k, v := range item.sobj {
			fieldPath := strings.TrimPrefix(fmt.Sprintf("%s.%s", item.fieldPath, k), ".")
			root := item.fieldPath == ""

			// if not map, add to fields slice
			if _, ok := v.(map[string]interface{}); !ok {
				fields = append(fields, &SObjectField{
					Name:   fieldPath,
					Value:  v,
					Dotted: !root,
				})
				continue
			}
			// object is map, add to items slice to be processed
			items = append(items, &nestedSObject{fieldPath, v.(map[string]interface{})})
		}
	}
	return &SObjectBuilder{fields}
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

// AddField adds the field to the Salesforce Object overwritting any existing value.
// This methods returns an error if the field is invalid
// Example:
// sobj := SObject{}
// _ = soject.AddField("a.b.c", 100)
// The result of this is the following map: map[a.b.c:100]
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

// AddDottedField adds the field to the Salesforce Object overwritting any existing value.
// The difference from AddField is this method will traverse down the object
// (creating paths if necessary) to add the field.
// This method returns an error if the field path is invalid or a non-leaf node in the field
// path is set to a non map[string]interface{} value.
// Example:
// sobj := SObject{}
// _ = soject.AddDottedField("a.b.c", 100)
// The result of this is the following map: map[a:map[b:map[c:100]]]
// Note: the following would return an error if attempted on the SObject above:
// sobject.AddDottedField("a.b.c.d", 100)
// This is because the `c` entry in the object is set to a non-map value a map .
// The following does succeed however:
// _ = sobject.AddDottedField("a.b.c", SObject{})
// _ = sobject.AddDottedField("a.b.c.d", 100)
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
// This returns an error if the field is invalid or the field is not in the object.
// Example:
// sobj := NewSObjectBuilder().AddDottedField("a.b.c", 100).AddField("a.b.c", 200).Build()
// v, _ = s.GetField("a.b.c")
// The value of v will be 200
// Refer to AddField and AddDottedField.
func (s SObject) GetField(field string) (interface{}, error) {
	if field == "" {
		return nil, errInvalidField
	}
	v, ok := s[field]
	if !ok {
		return nil, &FieldNotFoundErr{field}
	}
	return v, nil
}

// GetDottedField returns the field value on the Salesforce Object.
// The difference from GetField is this will traverse down the object to retrieve the field.
// This methoed returns error if the field is invalid, a non-leaf node is a non-map value, or the
// field is not in the object.
// Example:
// sobj := NewSObjectBuilder().AddDottedField("a.b.c", 100).AddField("a.b.c", 200).Build()
// v, _ = s.GetField("a.b.c")
// The value of v will be 100
// Refer to AddField and AddDottedField.
func (s SObject) GetDottedField(dottedField string) (interface{}, error) {
	sobj, err := getLastSObjectInPath(s, dottedField, false, false)
	if err != nil {
		return nil, err
	}
	return sobj.GetField(getLastFieldInPath(dottedField))
}

// GetMandatoryField calls GetField and panic if it receives an error.
// Refer to the GetField documentation.
func (s SObject) GetMandatoryField(field string) interface{} {
	v, err := s.GetField(field)
	if err != nil {
		panic(err)
	}
	return v
}

// GetMandatoryDottedField calls GetDottedField and panic if it receives an error.
// Refer to the GetDottedField documentation.
func (s SObject) GetMandatoryDottedField(dottedField string) interface{} {
	v, err := s.GetDottedField(dottedField)
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
				return nil, &InvalidFieldTypeErr{dottedField, field, v}
			} else if !ok && replaceIfWrongType {
				sobj.mandatoryAddField(field, map[string]interface{}{})
			}
		}
		sobj = sobj.GetMandatoryField(field).(map[string]interface{})
	}
	return sobj, nil
}
