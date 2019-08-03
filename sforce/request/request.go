package request

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/Laugusti/go-sforce/sforce/session"
	"github.com/Laugusti/go-sforce/sforce/sforceerr"
)

// ResultType is body type (json, xml, etc.) of the Request result.
type ResultType int

// result types
const (
	DiscardResult ResultType = iota
	JSONResult
	XMLResult
)

// Operation represents an http operation.
type Operation struct {
	Method  string
	APIPath string
	Query   string
	Body    io.Reader
}

// ResultExpectation stores the expected result type and status codes.
type ResultExpectation struct {
	Type        ResultType
	StatusCodes []int
}

// NewResultExpectation creates a result expectation.
func NewResultExpectation(respType ResultType, statusCodes ...int) *ResultExpectation {
	return &ResultExpectation{respType, statusCodes}
}

// Request is used to execute an Operation.
type Request struct {
	sess *session.Session
}

// New creates a new Request.
func New(sess *session.Session) *Request {
	return &Request{sess}
}

// buildRequest creates a http.Request struct for the api path.
func (r *Request) buildRequest(apiPath, rawQuery, method string, body io.Reader) (*http.Request, error) {
	// ensure session is authorized
	if !r.sess.HasToken() {
		if err := r.sess.Login(); err != nil {
			return nil, err
		}
	}

	// build api url using instance url
	apiURL, err := joinURL(r.sess.InstanceURL(), apiPath)
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
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", r.sess.AccessToken()))
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

// DoWithResult executes the http operation and unmarshals the result into the interface.
func (r *Request) DoWithResult(op *Operation, expectedResponse *ResultExpectation, result interface{}) error {
	req, err := r.buildRequest(op.APIPath, op.Query, op.Method, op.Body)
	if err != nil {
		return err
	}
	// do http request
	resp, err := r.sess.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// login and retry if unauthorized
	if resp.StatusCode == http.StatusUnauthorized {
		err := r.sess.Login()
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
		retryResp, err := r.sess.HTTPClient.Do(req)
		if err != nil {
			return fmt.Errorf("request failed: %v", err)
		}
		// close original body and update to new response
		_ = resp.Body.Close()
		resp = retryResp
	}

	// unmarshal response based on wanted type
	switch expectedResponse.Type {
	case DiscardResult:
		return nil
	case JSONResult:
		return unmarshalResponse(json.Unmarshal, resp,
			expectedResponse.StatusCodes, result)
	case XMLResult:
		return unmarshalResponse(xml.Unmarshal, resp,
			expectedResponse.StatusCodes, result)
	default:
		return errors.New("unknown result type")
	}
}

func unmarshalResponse(unmarshalFunc func([]byte, interface{}) error, resp *http.Response,
	validCodes []int, result interface{}) error {
	// get response body
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %v", err)
	}
	// return api error if status code is unexpected
	if len(validCodes) > 0 && !isInSlice(resp.StatusCode, validCodes) {
		var apiErr sforceerr.APIError
		if err := unmarshalFunc(data, apiErr); err != nil || apiErr.ErrorCode == "" {
			// failed to get api error
			return fmt.Errorf("unexpected status code (want %v, got %d): %s",
				validCodes, resp.StatusCode, data)
		}
		apiErr.ActualStatusCode = resp.StatusCode
		apiErr.ExpectedStatusCodes = validCodes
		return &apiErr
	}

	// unmarshal to result
	if result != nil {
		if err := unmarshalFunc(data, result); err != nil {
			return fmt.Errorf("failed to unmarshal response: %v", err)
		}
	}
	return nil
}
