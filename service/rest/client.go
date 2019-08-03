package restclient

import (
	"net/http"

	"github.com/Laugusti/go-sforce/session"
)

// Client handles request/response with the Salesforce API.
type Client struct {
	sess       *session.Session
	httpClient *http.Client
}

// New returns a new rest client for the Salesforce session.
func New(sess *session.Session) *Client {
	return &Client{sess, &http.Client{}}
}
