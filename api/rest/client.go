package restapi

import (
	"github.com/Laugusti/go-sforce/sforce/session"
)

// Client handles request/response with the Salesforce API.
type Client struct {
	sess *session.Session
}

// NewClient returns a new rest client for the Salesforce session.
func NewClient(sess *session.Session) *Client {
	return &Client{sess}
}
