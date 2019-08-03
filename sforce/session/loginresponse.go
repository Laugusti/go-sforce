package session

import "fmt"

// RequestToken stores the oauth token result from the Salesforce API.
type RequestToken struct {
	AccessToken string `json:"access_token"`
	InstanceURL string `json:"instance_url"`
	ID          string `json:"id"`
	TokenType   string `json:"token_type"`
	IssuedAt    string `json:"issued_at"`
	Signature   string `json:"signature"`
}

// LoginError is an unsuccessful login response
type LoginError struct {
	ErrorCode string `json:"error" xml:"Body>Fault>faultcode"`
	Message   string `json:"error_description" xml:"Body>Fault>faultstring"`
}

func (e *LoginError) Error() string {
	return fmt.Sprintf("login failed with %s: %s", e.ErrorCode, e.Message)
}

func (e LoginError) hasError() bool {
	return e.ErrorCode != "" && e.Message != ""
}
