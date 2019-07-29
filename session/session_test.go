package session

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/Laugusti/go-sforce/credentials"
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
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		assert.Nil(t, err)
		checkLoginFormValues(t, r.Form, "grant_type", "password")
		checkLoginFormValues(t, r.Form, "client_id", "clientid")
		checkLoginFormValues(t, r.Form, "client_secret", "clientsecret")
		checkLoginFormValues(t, r.Form, "username", "user")
		checkLoginFormValues(t, r.Form, "password", "pass")
		_, err = io.Copy(w, strings.NewReader(`{"instance_url": "testUrl", "access_token": "testToken"}`))
		assert.Nil(t, err)
	}))
	defer s.Close()
	sess := &Session{
		LoginURL: s.URL,
		creds: &credentials.OAuth{
			Username:     "user",
			Password:     "pass",
			ClientID:     "clientid",
			ClientSecret: "clientsecret",
		},
		httpClient: s.Client(),
	}

	if sess.HasToken() {
		t.Error("TestLogin => should not have token")
	}
	err := sess.Login()
	assert.Nil(t, err)

	assert.True(t, sess.HasToken(), "should have token")
	assert.Equal(t, "testToken", sess.AccessToken(), "wrong access token")
	assert.Equal(t, "testUrl", sess.InstanceURL(), "wrong instance url")
}

func checkLoginFormValues(t *testing.T, form url.Values, key, want string) {
	got := form.Get(key)
	if got != want {
		t.Errorf("TestLogin => wrong %s: expected %s, got %s", key, want, got)
	}
}
