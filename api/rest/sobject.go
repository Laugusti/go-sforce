package restapi

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

var (
	errNilMap          = errors.New("cannot add field to nil map")
	invalidFieldNameRE = regexp.MustCompile(`^([^a-zA-Z].*|.*__.*|.*[\W]+.*|.*[^a-zA-Z0-9]|)$`)
)

// InvalidFieldError represents an invalid field error.
type InvalidFieldError struct {
	FieldName string
}

func (e *InvalidFieldError) Error() string {
	return fmt.Sprintf("field %q is not valid", e.FieldName)
}

// UnknownFieldError represents an error for field not in the object.
type UnknownFieldError struct {
	FieldName string
}

func (e *UnknownFieldError) Error() string {
	return fmt.Sprintf("field %q not in salesforce object", e.FieldName)
}

// NonMapFieldError represents an invalid field type when traversing the object.
type NonMapFieldError struct {
	FullFieldName   string
	FailedFieldName string
	FailedField     interface{}
}

func (e *NonMapFieldError) Error() string {
	return fmt.Sprintf("failed to get field %q at %q: want map got %T", e.FullFieldName, e.FailedFieldName, e.FailedField)
}

// SObjectBuilderField is represents a field on a Salesforce Object. This is used by the
// Salesforce Object Builder to add fields to the Salesforce Object.
type SObjectBuilderField struct {
	Name   string
	Value  interface{}
	Dotted bool
}

// SObjectBuilder is used to build a Salesforce Object using fluent chainging.
type SObjectBuilder struct {
	Fields []*SObjectBuilderField
}

// NewSObjectBuilder returns a builder for a Salesforce Object.
func NewSObjectBuilder() *SObjectBuilder {
	return &SObjectBuilder{}
}

// SObjectBuilderFromSObject returns a builder for the SObject.
func SObjectBuilderFromSObject(s SObject) *SObjectBuilder {
	fields := []*SObjectBuilderField{}

	// new struct for sobjects remember the field path
	type nestedSObject struct {
		fieldPath string
		sobj      SObject
	}

	// start with root sobject
	items := []*nestedSObject{&nestedSObject{"", s}}
	for i := 0; i < len(items); i++ {
		item := items[i]
		for k, v := range item.sobj {
			fieldPath := strings.TrimPrefix(fmt.Sprintf("%s.%s", item.fieldPath, k), ".")
			root := item.fieldPath == ""

			// if not map, add to fields slice
			if _, ok := v.(map[string]interface{}); !ok {
				fields = append(fields, &SObjectBuilderField{
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
	sb.Fields = append(sb.Fields, &SObjectBuilderField{field, value, false})
	return sb
}

// NewDottedField adds the field using the AddDottedField method and returns the SObjectBuilder
// for fluent chaining.
func (sb *SObjectBuilder) NewDottedField(field string, value interface{}) *SObjectBuilder {
	sb.Fields = append(sb.Fields, &SObjectBuilderField{field, value, true})
	return sb
}

// Build builds the Salesforce Object.
func (sb *SObjectBuilder) Build() (SObject, error) {
	s := SObject{}
	for _, f := range sb.Fields {
		if f.Dotted {
			// add field to last sobject
			sobj, err := getLastSObjectInPath(s, f.Name, true, true)
			// invalid field should be the only error here
			if _, ok := err.(*InvalidFieldError); ok {
				return nil, &InvalidFieldError{f.Name}
			} else if err != nil {
				panic(err)
			}
			field := getLastFieldInPath(f.Name)
			err = sobj.AddField(field, f.Value)
			// invalid field should be the only error here
			if _, ok := err.(*InvalidFieldError); ok {
				return nil, &InvalidFieldError{f.Name}
			} else if err != nil {
				panic(err)
			}
		} else {
			// add field to root sobject, invalid field should be only error here
			err := s.AddField(f.Name, f.Value)
			if _, ok := err.(*InvalidFieldError); ok {
				return nil, err
			} else if err != nil {
				panic(err)
			}
		}
	}
	return s, nil
}

// MustBuild builds the Salesforce Object and panic if an error is received.
func (sb *SObjectBuilder) MustBuild() SObject {
	sobj, err := sb.Build()
	if err != nil {
		panic(err)
	}
	return sobj
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
	if isInvalidFieldName(field) {
		return &InvalidFieldError{field}
	}
	if s == nil {
		return errNilMap
	}
	s[field] = v
	return nil
}

func (s SObject) mustAddField(field string, v interface{}) {
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
	return sobj.AddField(getLastFieldInPath(dottedField), v)
}

// GetField returns the field value on the Salesforce Object.
// This returns an error if the field is invalid or the field is not in the object.
// Example:
// sobj := NewSObjectBuilder().AddDottedField("a.b.c", 100).AddField("a.b.c", 200).MustBuild()
// v, _ = s.GetField("a.b.c")
// The value of v will be 200
// Refer to AddField and AddDottedField.
func (s SObject) GetField(field string) (interface{}, error) {
	if isInvalidFieldName(field) {
		return nil, &InvalidFieldError{field}
	}
	v, ok := s[field]
	if !ok {
		return nil, &UnknownFieldError{field}
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

// split the dotted field by '.' and return the last string.
func getLastFieldInPath(dottedField string) string {
	// split field param using "."
	fields := strings.Split(dottedField, ".")
	return fields[len(fields)-1]
}

// returns the last sobject in field path for the dotted field, creating/replacing
// fields if selected and necessary.
func getLastSObjectInPath(s SObject, dottedField string, createIfNotExists, replaceIfWrongType bool) (SObject, error) {
	// split field param using "."
	fields := strings.Split(dottedField, ".")

	// drill down map object to get value
	var sobj = s
	for i := 0; i < len(fields)-1; i++ {
		field := fields[i]
		v, err := sobj.GetField(field)
		switch err := err.(type) {
		case nil:
			// found field, value must be map
			switch v := v.(type) {
			case map[string]interface{}:
				if v == nil && createIfNotExists {
					// field is nil, create it
					sobj.mustAddField(field, map[string]interface{}{})
				}
			default:
				// field is wrong type, replace it
				if !replaceIfWrongType {
					return nil, &NonMapFieldError{dottedField, field, v}
				}
				sobj.mustAddField(field, map[string]interface{}{})
			}
		case *UnknownFieldError:
			// field not found, add it
			if !createIfNotExists {
				return nil, err
			}
			sobj.mustAddField(field, map[string]interface{}{})
		default:
			return nil, err
		}
		sobj = sobj.GetMandatoryField(field).(map[string]interface{})
	}
	return sobj, nil
}

func isInvalidFieldName(fieldName string) bool {
	// allow field to end with __c/__r for custom fields/relationships
	fieldName = strings.TrimSuffix(fieldName, "__c")
	fieldName = strings.TrimSuffix(fieldName, "__r")
	return invalidFieldNameRE.MatchString(fieldName)
}
