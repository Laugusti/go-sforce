package restclient

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path"
)

const (
	sObjectPath = "/services/data/%s/sobjects/"
	queryPath   = "/services/data/%s/query/"
)

// CreateSObject creates the SObject using the Salesforce API.
func (c *Client) CreateSObject(sObjectName string, sObject SObject) (*UpsertResult, error) {
	// validate parameters
	if sObjectName == "" {
		return nil, errors.New("sobject name is required")
	}
	if sObject == nil || len(sObject) == 0 {
		return nil, errors.New("sobject value is required")
	}
	// build api path
	apiPath := path.Join(fmt.Sprintf(sObjectPath, c.sess.APIVersion), sObjectName)
	// do post and unmarshal response to result object
	var result UpsertResult
	if err := c.doPost(apiPath, sObject, &result); err != nil {
		return nil, fmt.Errorf("failed to create sobject: %v", err)
	}
	return &result, nil
}

// GetSObject retrieves the SObject from Salesforce API using the Id.
func (c *Client) GetSObject(sObjectName, id string) (SObject, error) {
	// validate parameters
	if sObjectName == "" {
		return nil, errors.New("sobject name is required")
	}
	if id == "" {
		return nil, errors.New("sobject id is required")
	}
	// build api path
	apiPath := path.Join(fmt.Sprintf(sObjectPath, c.sess.APIVersion), sObjectName, id)
	// do post and unmarshal response to result object
	var sObject SObject
	if err := c.doGet(apiPath, "", sObject); err != nil {
		return nil, fmt.Errorf("failed to get SObject: %v", err)
	}
	return sObject, nil
}

// GetSObjectByExternalID retrieves the SObject from the Salesforce API using the external Id.
func (c *Client) GetSObjectByExternalID(sObjectName, externalIDField, externalID string) (SObject, error) {
	// validate parameters
	if sObjectName == "" {
		return nil, errors.New("sobject name is required")
	}
	if externalIDField == "" {
		return nil, errors.New("external id field is required")
	}
	if externalID == "" {
		return nil, errors.New("external id is required")
	}
	// build api path
	apiPath := path.Join(fmt.Sprintf(sObjectPath, c.sess.APIVersion), sObjectName, externalIDField, externalID)
	// do post and unmarshal response to result object
	var sObject SObject
	if err := c.doGet(apiPath, "", sObject); err != nil {
		return nil, fmt.Errorf("failed to get SObject: %v", err)
	}
	return sObject, nil
}

// UpsertSObject creates/updates the SObject using the Salesforce API.
func (c *Client) UpsertSObject(sObjectName, id string, sObject SObject) (*UpsertResult, error) {
	// validate parameters
	if sObjectName == "" {
		return nil, errors.New("sobject name is required")
	}
	if id == "" {
		return nil, errors.New("sobject id is required")
	}
	if sObject == nil || len(sObject) == 0 {
		return nil, errors.New("sobject value is required")
	}
	// build api path
	apiPath := path.Join(fmt.Sprintf(sObjectPath, c.sess.APIVersion), sObjectName, id)
	// do post and unmarshal response to result object
	var result UpsertResult
	if err := c.doPatch(apiPath, sObject, &result); err != nil {
		return nil, fmt.Errorf("failed to upsert sobject: %v", err)
	}
	return &result, nil
}

// UpsertSObjectByExternalID creates/updates the SObject using the Salesforce API.
func (c *Client) UpsertSObjectByExternalID(sObjectName, externalIDField, externalID string, sObject SObject) (*UpsertResult, error) {
	// validate parameters
	if sObjectName == "" {
		return nil, errors.New("sobject name is required")
	}
	if externalIDField == "" {
		return nil, errors.New("external id field is required")
	}
	if externalID == "" {
		return nil, errors.New("external id is required")
	}
	if sObject == nil || len(sObject) == 0 {
		return nil, errors.New("sobject value is required")
	}
	// build api path
	apiPath := path.Join(fmt.Sprintf(sObjectPath, c.sess.APIVersion), sObjectName, externalIDField, externalID)
	// do post and unmarshal response to result object
	var result UpsertResult
	if err := c.doPatch(apiPath, sObject, &result); err != nil {
		return nil, fmt.Errorf("failed to upsert sobject: %v", err)
	}
	return &result, nil
}

// DeleteSObject deletes the Sobject using the Salesforce API.
func (c *Client) DeleteSObject(sObjectName, id string) error {
	// validate parameters
	if sObjectName == "" {
		return errors.New("sobject name is required")
	}
	if id == "" {
		return errors.New("sobject id is required")
	}
	// build api path
	apiPath := path.Join(fmt.Sprintf(sObjectPath, c.sess.APIVersion), sObjectName, id)
	// build http request
	req, err := c.buildRequest(apiPath, "", http.MethodDelete, nil)
	if err != nil {
		return fmt.Errorf("failed to deleted sobject: %v", err)
	}
	// do delete
	if err := c.doRequest(req, nil, http.StatusNoContent); err != nil {
		return fmt.Errorf("failed to deleted sobject: %v", err)
	}
	return nil
}

// Query executes a SOQL query using the Salesforce API.
func (c *Client) Query(query string) (*QueryResult, error) {
	// validate parameters
	if query == "" {
		return nil, errors.New("query string is required")
	}
	// build api path
	apiPath := fmt.Sprintf(queryPath, c.sess.APIVersion)
	rawQuery := fmt.Sprintf("q=%s", url.QueryEscape(query))

	// do get and unmarshal response to result object
	var queryResult QueryResult
	if err := c.doGet(apiPath, rawQuery, &queryResult); err != nil {
		return nil, fmt.Errorf("failed to query salesforce: %v", err)
	}
	return &queryResult, nil
}

// QueryMore retrieves the next batch of query records from the Salesforce API.
func (c *Client) QueryMore(query *QueryResult) (*QueryResult, error) {
	// validate parameters
	if query == nil || query.NextRecordsURL == "" {
		return nil, errors.New("missing next records url")
	}
	// do get and unmarshal response to result object
	var queryResult QueryResult
	if err := c.doGet(query.NextRecordsURL, "", &queryResult); err != nil {
		return nil, fmt.Errorf("failed to query salesforce: %v", err)
	}
	return &queryResult, nil
}

// FullQuery executes the SOQL query using the Salesforce API and executes QueryMore to get the
// complete query result if necessary.
func (c *Client) FullQuery(query string) (*QueryResult, error) {
	partial, err := c.Query(query)
	if err != nil {
		return nil, err
	}
	for !partial.Done {
		next, err := c.QueryMore(partial)
		if err != nil {
			return nil, err
		}
		partial.Records = append(partial.Records, next.Records...)
		partial.Done = next.Done
		partial.NextRecordsURL = next.NextRecordsURL
	}
	return partial, nil
}

func (c *Client) doGet(apiPath, rawQuery string, result interface{}) error {
	req, err := c.buildRequest(apiPath, rawQuery, http.MethodGet, nil)
	if err != nil {
		return err
	}

	err = c.doRequest(req, result, http.StatusOK)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) doPatch(apiPath string, sObject SObject, result *UpsertResult) error {
	body := &bytes.Buffer{}
	if err := json.NewEncoder(body).Encode(sObject); err != nil {
		return fmt.Errorf("failed to marshal sobject: %v", err)
	}
	req, err := c.buildRequest(apiPath, "", http.MethodPatch, body)
	if err != nil {
		return err
	}
	err = c.doRequest(req, result, http.StatusOK, http.StatusCreated)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) doPost(apiPath string, sObject SObject, result *UpsertResult) error {
	body := &bytes.Buffer{}
	if err := json.NewEncoder(body).Encode(sObject); err != nil {
		return fmt.Errorf("failed to marshal sobject: %v", err)
	}
	req, err := c.buildRequest(apiPath, "", http.MethodPost, body)
	if err != nil {
		return err
	}
	err = c.doRequest(req, result, http.StatusCreated)
	if err != nil {
		return err
	}
	return nil
}
