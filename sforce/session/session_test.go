package session

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/Laugusti/go-sforce/internal/testserver"
	"github.com/Laugusti/go-sforce/sforce/credentials"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	type sessInput struct {
		loginURL     string
		apiVersion   string
		username     string
		password     string
		clientID     string
		clientSecret string
	}
	tests := []struct {
		input  sessInput
		errMsg string
	}{
		{sessInput{"", "", "", "", "", ""}, "Login URL is required;API Version is required;Username is required;Password is required;Client ID is required;Client Secret is required"},
		{sessInput{"a", "", "", "", "", ""}, "API Version is required;Username is required;Password is required;Client ID is required;Client Secret is required"},
		{sessInput{"a", "b", "", "", "", ""}, "Username is required;Password is required;Client ID is required;Client Secret is required"},
		{sessInput{"a", "b", "c", "", "", ""}, "Password is required;Client ID is required;Client Secret is required"},
		{sessInput{"a", "b", "c", "d", "", ""}, "Client ID is required;Client Secret is required"},
		{sessInput{"a", "b", "c", "d", "e", ""}, "Client Secret is required"},
		{sessInput{"a", "b", "c", "d", "e", "f"}, ""},
	}

	for _, test := range tests {
		assertMsg := fmt.Sprintf("input: %v", test)
		_, err := New(
			test.input.loginURL,
			test.input.apiVersion,
			credentials.New(
				test.input.username,
				test.input.password,
				test.input.clientID,
				test.input.clientSecret,
			),
		)
		if test.errMsg == "" {
			assert.Nil(t, err, assertMsg)
		} else {
			if assert.NotNilf(t, err, assertMsg) {
				assert.Contains(t, err.Error(), test.errMsg, assertMsg)
			}
		}
	}
}

func TestLogin(t *testing.T) {
	server := testserver.New(t)
	defer server.Stop()

	tests := []struct {
		// inputs
		shouldErr    bool
		apiVersion   string
		username     string
		password     string
		clientID     string
		clientSecret string
		// wants
		accessToken string
		instanceURL string
		id          string
		tokenType   string
		issuedAt    string
		signature   string
	}{
		{false, "1.0", "user", "pass", "id", "secret", "", "", "", "", "", ""},
		{false, "2.0", "u", "p", "i", "s", "token", "url", "id", "type", "at", "sig"},
		{true, "1.0", "user", "pass", "id", "secret", "", "", "", "", "", ""},
	}

	for _, test := range tests {
		assertMsg := fmt.Sprintf("input: %v", test)

		// wanted request token
		wantToken := RequestToken{
			AccessToken: test.accessToken,
			InstanceURL: test.instanceURL,
			ID:          test.id,
			TokenType:   test.tokenType,
			IssuedAt:    test.issuedAt,
			Signature:   test.signature,
		}

		// oauth creds
		creds := credentials.New(
			test.username,
			test.password,
			test.clientID,
			test.clientSecret,
		)

		// set server handler func
		setHandlerFunc(t, assertMsg, server, *creds, wantToken, test.shouldErr)

		// create new session
		sess := Must(New(server.URL(), test.apiVersion, creds))
		sess.httpClient = server.Client()

		// assert session does not have token
		assert.Nil(t, sess.requestToken, assertMsg)

		// set dummy request token to ensure it is overwritten
		sess.requestToken = &RequestToken{"DUMMY", "DUMMY", "DUMMY", "DUMMY", "DUMMY", "DUMMY"}

		// try session login
		err := sess.Login()

		// session assertions
		if test.shouldErr {
			assert.NotNil(t, err, assertMsg)
			_, ok := err.(*LoginError)
			assert.True(t, ok, assertMsg)
			assert.Nil(t, sess.requestToken, assertMsg)
		} else {
			assert.Nil(t, err, assertMsg)
			assert.NotNil(t, sess.requestToken, assertMsg)
			assert.True(t, sess.HasToken(), assertMsg)
			assert.Equal(t, &wantToken, sess.requestToken, assertMsg)
			assert.Equal(t, test.instanceURL, sess.InstanceURL(), assertMsg)
			assert.Equal(t, test.accessToken, sess.AccessToken(), assertMsg)
		}
	}
}

func TestConcurrentSession(t *testing.T) {
	server := testserver.New(t)
	defer server.Stop()

	// delay next login request
	delayDuration := 3 * time.Second
	delayNextLogin(server, delayDuration)

	// create new session
	sess := Must(New(server.URL(), "1", credentials.New("u", "p", "ci", "cs")))
	sess.httpClient = server.Client()

	// channel with duration
	durationChan := make(chan time.Duration)

	// session functions
	loginFunc := func() { _ = sess.Login() } // first invocation is delayed
	hasTokenFunc := func() { _ = sess.HasToken() }
	instanceURLFunc := func() { _ = sess.InstanceURL() }
	accessTokenFunc := func() { _ = sess.AccessToken() }

	// invoke functions concurrently and report duration
	go reportFuncDuration(durationChan, loginFunc) // this should lock session method calls
	go reportFuncDuration(durationChan, hasTokenFunc)
	go reportFuncDuration(durationChan, instanceURLFunc)
	go reportFuncDuration(durationChan, accessTokenFunc)
	go reportFuncDuration(durationChan, loginFunc)

	// assert durations
	assert.Greater(t, int64(<-durationChan), int64(delayDuration))
	assert.Greater(t, int64(<-durationChan), int64(delayDuration))
	assert.Greater(t, int64(<-durationChan), int64(delayDuration))
	assert.Greater(t, int64(<-durationChan), int64(delayDuration))
	assert.Greater(t, int64(<-durationChan), int64(delayDuration))

	// request are no longer delayed
	go reportFuncDuration(durationChan, loginFunc)
	go reportFuncDuration(durationChan, accessTokenFunc)
	assert.Less(t, int64(<-durationChan), int64(delayDuration))
	assert.Less(t, int64(<-durationChan), int64(delayDuration))

	// close channel
	close(durationChan)
}

func delayNextLogin(server *testserver.Server, delayDuration time.Duration) {
	// set handler func
	handledCount := 0
	server.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
		h := &testserver.JSONResponseHandler{
			StatusCode: 200,
			Body: RequestToken{
				AccessToken: "token",
				InstanceURL: server.URL(),
			},
		}
		// delay first response
		if handledCount == 0 {
			time.Sleep(delayDuration)
		}
		handledCount++
		_ = h.Handle(w)
	}
}

func reportFuncDuration(durationChan chan<- time.Duration, f func()) {
	start := time.Now()
	f()
	durationChan <- time.Since(start)
}

func setHandlerFunc(t *testing.T, assertMsg string, s *testserver.Server, creds credentials.OAuth,
	token RequestToken, shouldErr bool) {
	// default handler returns token and 200 status
	handler := &testserver.JSONResponseHandler{
		StatusCode: http.StatusOK,
		Body:       token,
	}
	// error handler returns error response and 400 status
	if shouldErr {
		handler = &testserver.JSONResponseHandler{
			StatusCode: http.StatusBadRequest,
			Body: map[string]interface{}{
				"error":             "login_failed",
				"error_description": "authentication failure",
			},
		}
	}
	// validating query={empty}, method=POST,
	// path=oathtokenpath, form=creds, body={empty}
	validators := []testserver.RequestValidator{
		&testserver.QueryValidator{Query: url.Values{}},
		&testserver.MethodValidator{Method: http.MethodPost},
		&testserver.PathValidator{Path: "/services/oauth2/token"},
		&testserver.FormValidator{Form: credsAsForm(creds)},
		&testserver.JSONBodyValidator{}, // validate body after form
	}

	// set server handler func
	s.HandlerFunc = testserver.ValidateRequestHandlerFunc(t, assertMsg, handler, validators...)
}

func credsAsForm(creds credentials.OAuth) url.Values {
	form := url.Values{}
	form.Set("grant_type", "password")
	form.Set("username", creds.Username)
	form.Set("password", creds.Password)
	form.Set("client_id", creds.ClientID)
	form.Set("client_secret", creds.ClientSecret)
	return form
}
