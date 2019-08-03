package sforceerr

import "fmt"

// APIError is an unsuccessful response from the Salesforce API.
type APIError struct {
	Fields              []string `json:"fields" xml:"fields"`
	Message             string   `json:"message" xml:"message"`
	ErrorCode           string   `json:"errorCode" xml:"errorCode"`
	ExpectedStatusCodes []int    `json:"-" xml:"-"`
	ActualStatusCode    int      `json:"-" xml:"-"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("%+v", *e)
}
