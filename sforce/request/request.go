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
	JSONResult ResultType = iota
	XMLResult
)

// Operation represents an http operation.
type Operation struct {
	Method   string
	APIPath  string
	RawQuery string
	Body     io.Reader
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
	sess            *session.Session
	op              *Operation
	expect          *ResultExpectation
	result          interface{}
	preSendHandlers []func(*http.Request)
}

// New creates a new Request.
func New(sess *session.Session, op *Operation, expect *ResultExpectation,
	result interface{}, preSendHandlers ...func(*http.Request)) *Request {
	return &Request{sess, op, expect, result, preSendHandlers}
}

// buildRequest creates a http.Request struct for the api path.
func (r *Request) buildRequest(op *Operation, preSendHandlers []func(*http.Request)) (*http.Request, error) {
	// ensure session is authorized
	if !r.sess.HasToken() {
		if err := r.sess.Login(); err != nil {
			return nil, err
		}
	}

	// build api url using instance url
	apiURL, err := joinURL(r.sess.InstanceURL(), op.APIPath)
	if err != nil {
		return nil, fmt.Errorf("failed to build request: %v", err)
	}
	// add query to api url
	u, _ := url.Parse(apiURL)
	u.RawQuery = op.RawQuery
	apiURL = u.String()

	// creates http reqeust
	req, err := http.NewRequest(op.Method, apiURL, op.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to build request: %v", err)
	}

	// run pre send handlers
	for _, h := range preSendHandlers {
		h(req)
	}

	return req, nil
}

// Send executes the http operation and unmarshals the result into the interface.
func (r *Request) Send() error {
	req, err := r.buildRequest(r.op, r.preSendHandlers)
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
	switch r.expect.Type {
	case JSONResult:
		return unmarshalResponse(json.Unmarshal, resp,
			r.expect.StatusCodes, r.result)
	case XMLResult:
		return unmarshalResponse(xml.Unmarshal, resp,
			r.expect.StatusCodes, r.result)
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
	if !isInSlice(resp.StatusCode, validCodes) {
		var apiErr sforceerr.APIError
		if err := unmarshalFunc(data, &[]*sforceerr.APIError{&apiErr}); err != nil || apiErr.ErrorCode == "" {
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
