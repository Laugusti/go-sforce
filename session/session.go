package session

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/Laugusti/go-sforce/credentials"
)

const (
	oauthTokenPath = "/services/oauth2/token"
)

// Session stores the credentials and is used to create clients.
type Session struct {
	LoginURL     string
	APIVersion   string
	creds        *credentials.OAuth
	httpClient   *http.Client
	instanceURL  string
	requestToken *RequestToken
}

// New returns a new Session.
func New(loginURL, apiVersion string, creds *credentials.OAuth) (*Session, error) {
	var errMsg []string
	if loginURL == "" {
		errMsg = append(errMsg, "Login URL is required")
	}
	if apiVersion == "" {
		errMsg = append(errMsg, "API Version is required")
	}
	if creds.Username == "" {
		errMsg = append(errMsg, "Username is required")
	}
	if creds.Password == "" {
		errMsg = append(errMsg, "Password is required")
	}
	if creds.ClientID == "" {
		errMsg = append(errMsg, "Client ID is required")
	}
	if creds.ClientSecret == "" {
		errMsg = append(errMsg, "Client Secret is required")
	}
	if len(errMsg) != 0 {
		return nil, errors.New(strings.Join(errMsg, ";"))
	}

	return &Session{
		LoginURL:   loginURL,
		APIVersion: apiVersion,
		creds: &credentials.OAuth{Username: creds.Username,
			Password:     creds.Password,
			ClientID:     creds.ClientID,
			ClientSecret: creds.ClientSecret,
		},
		httpClient: &http.Client{},
	}, nil
}

// Must is a helper function to ensure the Session is valid and there was no error when calling
// the New function
func Must(sess *Session, err error) *Session {
	if err != nil {
		panic(err)
	}
	return sess
}

// Login requests an access token from the Salesforce API
func (s *Session) Login() error {
	u, err := url.Parse(s.LoginURL)
	if err != nil {
		return err
	}
	u.Path = path.Join(u.Path, oauthTokenPath)
	form := url.Values{}
	form.Set("grant_type", "password")
	form.Set("client_id", s.creds.ClientID)
	form.Set("client_secret", s.creds.ClientSecret)
	form.Set("username", s.creds.Username)
	form.Set("password", s.creds.Password)

	resp, err := s.httpClient.PostForm(u.String(), form)
	if err != nil {
		return fmt.Errorf("failed to get access token: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("unexpected status (want %d, got %d): failed to read response body",
				200, resp.StatusCode)
		}
		return fmt.Errorf("unexpected status (want %d, got %d): %s", 200, resp.StatusCode, b)
	}
	var result RequestToken
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return fmt.Errorf("failed to unmarshal request token: %v", err)
	}
	s.requestToken = &result
	return nil
}

// HasToken returns true if the session has a request token, otherwise false.
func (s *Session) HasToken() bool {
	return s.requestToken != nil
}

// AccessToken returns the access token from the Login response.
func (s *Session) AccessToken() string {
	if s.requestToken == nil {
		return ""
	}
	return s.requestToken.AccessToken
}

// InstanceURL returns the instance url from the Login response.
func (s *Session) InstanceURL() string {
	if s.requestToken == nil {
		return ""
	}
	return s.requestToken.InstanceURL
}
