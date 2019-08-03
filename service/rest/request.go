package restclient

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"

	"github.com/Laugusti/go-sforce/sforceerr"
)

// buildRequest creates a http.Request struct for the api path.
func (c *Client) buildRequest(apiPath, rawQuery, method string, body io.Reader) (*http.Request, error) {
	// ensure session is authorized
	if !c.sess.HasToken() {
		if err := c.sess.Login(); err != nil {
			return nil, err
		}
	}

	// build api url using instance url
	apiURL, err := joinURL(c.sess.InstanceURL(), apiPath)
	if err != nil {
		return nil, fmt.Errorf("failed to build request: %v", err)
	}
	// add query to api url
	u, _ := url.Parse(apiURL)
	u.RawQuery = rawQuery
	apiURL = u.String()

	// creates http reqeust
	req, err := http.NewRequest(method, apiURL, body)
	if err != nil {
		return nil, fmt.Errorf("failed to build request: %v", err)
	}

	// set auth and content type
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.sess.AccessToken()))
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

// doRequest sends the HTTP request and unmarshals the result into the interface.
func (c *Client) doRequest(req *http.Request, result interface{}, validStatuses ...int) error {
	// do http request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// login and retry if unauthorized
	if resp.StatusCode == http.StatusUnauthorized {
		err := c.sess.Login()
		if err != nil {
			return err
		}
		// request body was consumed, resetting
		body, err := req.GetBody()
		if err != nil {
			return fmt.Errorf("failed to get request body for retry: %v", err)
		}
		req.Body = body
		// retry request
		retryResp, err := c.httpClient.Do(req)
		if err != nil {
			return fmt.Errorf("request failed: %v", err)
		}
		// close original body and update to new response
		_ = resp.Body.Close()
		resp = retryResp
	}

	// check status code
	if len(validStatuses) > 0 && !isInSlice(resp.StatusCode, validStatuses) {
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("unexpected status code (want %v, got %d): %s",
				validStatuses, resp.StatusCode, "failed to read request body")
		}
		var apiErr sforceerr.APIError
		if err := json.Unmarshal(b, &apiErr); err != nil || apiErr.ErrorCode == "" {
			return fmt.Errorf("unexpected status code (want %v, got %d): %s",
				validStatuses, resp.StatusCode, b)
		}
		apiErr.ActualStatusCode = resp.StatusCode
		apiErr.ExpectedStatusCodes = validStatuses
		return &apiErr
	}

	// unmarshal response
	if result != nil {
		if resp.StatusCode == http.StatusNoContent {
			return errors.New("expected response body, got <nil>")
		}
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("failed to unmarshal response: %v", err)
		}
	}
	return nil
}

func joinURL(baseURL string, paths ...string) (string, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}
	u.Path = path.Join(u.Path, path.Join(paths...))
	return u.String(), nil
}

// isInSlice returns true if the value is in the slice.
func isInSlice(value int, slice []int) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}
