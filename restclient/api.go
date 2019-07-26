package restclient

import (
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

// GetSObject retrieves the SObject from Salesforce API using the Id.
func (c *Client) GetSObject(sObjectName, id string) (SObject, error) {
	apiPath := path.Join(fmt.Sprintf(sObjectPath, c.sess.APIVersion), sObjectName, id)
	var sObject SObject
	if err := c.doGet(apiPath, "", sObject); err != nil {
		return nil, fmt.Errorf("failed to get SObject: %v", err)
	}
	return sObject, nil
}

// GetSObjectByExternalID retrieves the SObject from the Salesforce API using the external Id.
func (c *Client) GetSObjectByExternalID(sObjectName, externalIDField, externalID string) (SObject, error) {
	apiPath := path.Join(fmt.Sprintf(sObjectPath, c.sess.APIVersion), sObjectName, externalIDField, externalID)
	var sObject SObject
	if err := c.doGet(apiPath, "", sObject); err != nil {
		return nil, fmt.Errorf("failed to get SObject: %v", err)
	}
	return sObject, nil
}

// DeleteSObject deletes the Sobject using the Salesforce API.
func (c *Client) DeleteSObject(sObjectName, id string) error {
	apiPath := path.Join(fmt.Sprintf(sObjectPath, c.sess.APIVersion), sObjectName, id)
	req, err := c.buildRequest(apiPath, "", http.MethodDelete, nil)
	if err != nil {
		return fmt.Errorf("failed to deleted sobject: %v", err)
	}
	if err := c.doRequest(req, nil, http.StatusNoContent); err != nil {
		return fmt.Errorf("failed to deleted sobject: %v", err)
	}
	return nil
}

// Query executes a SOQL query using the Salesforce API.
func (c *Client) Query(query string) (*QueryResult, error) {
	apiPath := fmt.Sprintf(queryPath, c.sess.APIVersion)
	rawQuery := fmt.Sprintf("q=%s", url.QueryEscape(query))

	var queryResult QueryResult
	if err := c.doGet(apiPath, rawQuery, &queryResult); err != nil {
		return nil, fmt.Errorf("failed to query salesforce: %v", err)
	}
	return &queryResult, nil
}

// QueryMore retrieves the next batch of query records from the Salesforce API.
func (c *Client) QueryMore(query *QueryResult) (*QueryResult, error) {
	var queryResult QueryResult
	if query == nil || query.NextRecordsURL == "" {
		return nil, errors.New("missing next records url")
	}
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
