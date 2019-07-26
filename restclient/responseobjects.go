package restclient

import (
	"errors"
	"fmt"
	"strings"
)

// SObject stores a Salesforce Object returned by the Salesforce API.
type SObject map[string]interface{}

// GetField returns the field value on the Salesforce Object.
func (s SObject) GetField(field string) (interface{}, error) {
	if field == "" {
		return nil, errors.New("field argument required")
	}

	// split field param using "." and drill down map object to get value
	var value interface{} = s
	for _, f := range strings.Split(field, ".") {
		m, ok := value.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("failed to get field: want map got %T", value)
		}
		v, ok := m[f]
		if !ok {
			return nil, fmt.Errorf("field %q not in salesforce object", field)
		}
		value = v
	}
	return value, nil

}

// GetMandatoryField ensures the field is valid for the Salesforce Object and there was no error
// when calling the GetField function.
func (s SObject) GetMandatoryField(field string) interface{} {
	value, err := s.GetField(field)
	if err != nil {
		panic(err)
	}
	return value
}

// UpsertResult is a successful response from the Salesforce API after an upsert.
type UpsertResult struct {
	ID      string        `json:"id"`
	Success bool          `json:"success"`
	Errors  []interface{} `json:"errors"`
}

// QueryResult is successful response from the Salesforce API after a query.
type QueryResult struct {
	TotalSize      int       `json:"totalSize"`
	Done           bool      `json:"done"`
	NextRecordsURL string    `json:"nextRecordsURL,omitempty"`
	Records        []SObject `json:"records"`
}

// APIError is an unsuccessful response from the Salesforce API.
type APIError struct {
	Fields    []string `json:"fields"`
	Message   string   `json:"message"`
	ErrorCode string   `json:"errorCode"`
}
