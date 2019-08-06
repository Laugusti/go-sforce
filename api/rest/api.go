package restapi

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/Laugusti/go-sforce/sforce/request"
)

const (
	sObjectPath = "/services/data/%s/sobjects/"
	queryPath   = "/services/data/%s/query/"
)

// CreateSObjectInput stores the input for creating a SObject.
type CreateSObjectInput struct {
	SObjectName string
	SObject     SObject
}

// CreateSObjectOutput stores the output after create a SObject.
type CreateSObjectOutput struct {
	Result *UpsertResult
}

// CreateSObject creates the SObject using the Salesforce API.
func (c *Client) CreateSObject(input *CreateSObjectInput) (*CreateSObjectOutput, error) {
	// validate parameters
	if isInvalidFieldName(input.SObjectName) {
		return nil, errors.New("invalid sobject name")
	}
	if len(input.SObject) == 0 {
		return nil, errors.New("sobject value is required")
	}

	buf := &bytes.Buffer{}
	if err := json.NewEncoder(buf).Encode(input.SObject); err != nil {
		return nil, fmt.Errorf("couldn't marshal sobject: %v", err)
	}
	var result UpsertResult
	req := c.newRequest(&request.Operation{
		Method: http.MethodPost,
		APIPath: path.Join(fmt.Sprintf(sObjectPath, c.sess.APIVersion),
			input.SObjectName),
		Body: buf,
	}, request.JSONResult, &result, http.StatusCreated)
	return &CreateSObjectOutput{&result}, req.Send()
}

// GetSObjectInput stores the input for retrieving a SObject by ID.
type GetSObjectInput struct {
	SObjectName string
	SObjectID   string
	Fields      []string
}

// GetSObjectOutput stores the output after retrieving a SObject.
type GetSObjectOutput struct {
	SObject SObject
}

// GetSObject retrieves a SObject from Salesforce.
func (c *Client) GetSObject(input *GetSObjectInput) (*GetSObjectOutput, error) {
	// validate parameters
	if isInvalidFieldName(input.SObjectName) {
		return nil, errors.New("invalid sobject name")
	}
	if input.SObjectID == "" {
		return nil, errors.New("sobject id is required")
	}
	for _, f := range input.Fields {
		if isInvalidFieldName(f) {
			return nil, errors.New("invalid field list")
		}
	}

	query := ""
	if len(input.Fields) > 0 {
		query = "fields=" + strings.Join(input.Fields, ",")
	}
	var sobj SObject
	req := c.newRequest(&request.Operation{
		Method:   http.MethodGet,
		RawQuery: query,
		APIPath: path.Join(fmt.Sprintf(sObjectPath, c.sess.APIVersion),
			input.SObjectName, input.SObjectID),
	}, request.JSONResult, &sobj, http.StatusOK)

	return &GetSObjectOutput{sobj}, req.Send()
}

// GetSObjectByExternalIDInput stores the input for retrieving a SObject by external ID.
type GetSObjectByExternalIDInput struct {
	SObjectName     string
	ExternalIDField string
	ExternalID      string
	Fields          []string
}

// GetSObjectByExternalIDOutput stores the output after retrieving a SObject by external
// ID.
type GetSObjectByExternalIDOutput struct {
	SObject SObject
}

// GetSObjectByExternalID retrieves the SObject from the Salesforce API using the external Id.
func (c *Client) GetSObjectByExternalID(input *GetSObjectByExternalIDInput) (*GetSObjectByExternalIDOutput, error) {
	// validate parameters
	if isInvalidFieldName(input.SObjectName) {
		return nil, errors.New("invalid sobject name")
	}
	if isInvalidFieldName(input.ExternalIDField) {
		return nil, errors.New("invalid external id field")
	}
	if input.ExternalID == "" {
		return nil, errors.New("external id is required")
	}
	for _, f := range input.Fields {
		if isInvalidFieldName(f) {
			return nil, errors.New("invalid field list")
		}
	}

	query := ""
	if len(input.Fields) > 0 {
		query = "fields=" + strings.Join(input.Fields, ",")
	}
	var sobj SObject
	req := c.newRequest(&request.Operation{
		Method:   http.MethodGet,
		RawQuery: query,
		APIPath: path.Join(fmt.Sprintf(sObjectPath, c.sess.APIVersion),
			input.SObjectName, input.ExternalIDField, input.ExternalID),
	}, request.JSONResult, &sobj, http.StatusOK)
	return &GetSObjectByExternalIDOutput{sobj}, req.Send()
}

// UpsertSObjectInput stores the input for upserting a SObject by ID.
type UpsertSObjectInput struct {
	SObjectName string
	SObjectID   string
	SObject     SObject
}

// UpsertSObjectOutput stores the output after upserting a SObject By ID.
type UpsertSObjectOutput struct {
	Result *UpsertResult
}

// UpsertSObject creates/updates the SObject using the Salesforce API.
func (c *Client) UpsertSObject(input *UpsertSObjectInput) (*UpsertSObjectOutput, error) {
	// validate parameters
	if isInvalidFieldName(input.SObjectName) {
		return nil, errors.New("invalid sobject name")
	}
	if input.SObjectID == "" {
		return nil, errors.New("sobject id is required")
	}
	if len(input.SObject) == 0 {
		return nil, errors.New("sobject value is required")
	}

	buf := &bytes.Buffer{}
	if err := json.NewEncoder(buf).Encode(input.SObject); err != nil {
		return nil, fmt.Errorf("couldn't marshal sobject: %v", err)
	}
	var result UpsertResult
	req := c.newRequest(&request.Operation{
		Method: http.MethodPatch,
		APIPath: path.Join(fmt.Sprintf(sObjectPath, c.sess.APIVersion),
			input.SObjectName, input.SObjectID),
		Body: buf,
	}, request.JSONResult, &result, http.StatusOK, http.StatusCreated)
	return &UpsertSObjectOutput{&result}, req.Send()
}

// UpsertSObjectByExternalIDInput stores the input for upserting a SObject by external ID.
type UpsertSObjectByExternalIDInput struct {
	SObjectName     string
	ExternalIDField string
	ExternalID      string
	SObject         SObject
}

// UpsertSObjectByExternalIDOutput stores the output after upserting a SObject by external
// ID.
type UpsertSObjectByExternalIDOutput struct {
	Result *UpsertResult
}

// UpsertSObjectByExternalID creates/updates the SObject using the Salesforce API.
func (c *Client) UpsertSObjectByExternalID(input *UpsertSObjectByExternalIDInput) (*UpsertSObjectByExternalIDOutput, error) {
	// validate parameters
	if isInvalidFieldName(input.SObjectName) {
		return nil, errors.New("invalid sobject name")
	}
	if isInvalidFieldName(input.ExternalIDField) {
		return nil, errors.New("invalid external id field")
	}
	if input.ExternalID == "" {
		return nil, errors.New("external id is required")
	}
	if len(input.SObject) == 0 {
		return nil, errors.New("sobject value is required")
	}

	buf := &bytes.Buffer{}
	if err := json.NewEncoder(buf).Encode(input.SObject); err != nil {
		return nil, fmt.Errorf("couldn't marshal sobject: %v", err)
	}
	var result UpsertResult
	req := c.newRequest(&request.Operation{
		Method: http.MethodPatch,
		APIPath: path.Join(fmt.Sprintf(sObjectPath, c.sess.APIVersion),
			input.SObjectName, input.ExternalIDField, input.ExternalID),
		Body: buf,
	}, request.JSONResult, &result, http.StatusOK, http.StatusCreated)
	return &UpsertSObjectByExternalIDOutput{&result}, req.Send()
}

// DeleteSObjectInput stores the input for deleting a SObject.
type DeleteSObjectInput struct {
	SObjectName string
	SObjectID   string
}

// DeleteSObjectOutput stores the output after deleting a SObject.
type DeleteSObjectOutput struct{}

// DeleteSObject deletes the Sobject using the Salesforce API.
func (c *Client) DeleteSObject(input *DeleteSObjectInput) (*DeleteSObjectOutput, error) {
	// validate parameters
	if isInvalidFieldName(input.SObjectName) {
		return nil, errors.New("invalid sobject name")
	}
	if input.SObjectID == "" {
		return nil, errors.New("sobject id is required")
	}

	req := c.newRequest(&request.Operation{
		Method: http.MethodDelete,
		APIPath: path.Join(fmt.Sprintf(sObjectPath, c.sess.APIVersion),
			input.SObjectName, input.SObjectID),
	}, request.JSONResult, nil, http.StatusNoContent)

	// do delete
	return &DeleteSObjectOutput{}, req.Send()
}

// QueryInput stores the input for querying SObjects.
type QueryInput struct {
	Query string
}

// QueryOutput stores the output after querying SObjects.
type QueryOutput struct {
	Result *QueryResult
}

// Query executes a SOQL query using the Salesforce API.
func (c *Client) Query(input *QueryInput) (*QueryOutput, error) {
	// validate parameters
	if input.Query == "" {
		return nil, errors.New("query string is required")
	}

	var queryResult QueryResult
	req := c.newRequest(&request.Operation{
		Method:   http.MethodGet,
		APIPath:  fmt.Sprintf(queryPath, c.sess.APIVersion),
		RawQuery: fmt.Sprintf("q=%s", url.QueryEscape(input.Query)),
	}, request.JSONResult, &queryResult, http.StatusOK)
	return &QueryOutput{&queryResult}, req.Send()
}

// QueryMoreInput stores the input for querying the next batch of records.
type QueryMoreInput struct {
	NextRecordsURL string
}

// QueryMoreOutput stores the output after querying the next batch of records.
type QueryMoreOutput struct {
	Result *QueryResult
}

// QueryMore retrieves the next batch of query records from the Salesforce API.
func (c *Client) QueryMore(input *QueryMoreInput) (*QueryMoreOutput, error) {
	// validate parameters
	if input.NextRecordsURL == "" {
		return nil, errors.New("missing next records url")
	}

	var queryResult QueryResult
	req := c.newRequest(&request.Operation{
		Method:  http.MethodGet,
		APIPath: input.NextRecordsURL,
	}, request.JSONResult, &queryResult, http.StatusOK)
	return &QueryMoreOutput{&queryResult}, req.Send()
}

func (c *Client) newRequest(op *request.Operation, resultType request.ResultType,
	result interface{}, statusCodes ...int) *request.Request {
	return request.New(c.sess, op,
		request.NewResultExpectation(resultType, statusCodes...),
		result, c.setAuthAndContentTypeFunc())
}

func (c *Client) setAuthAndContentTypeFunc() func(*http.Request) {
	return func(r *http.Request) {
		r.Header.Set("Content-Type", "application/json")
		r.Header.Set("Authorization", "Bearer "+c.sess.AccessToken())
	}
}
