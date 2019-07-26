package session

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/Laugusti/go-sforce/credentials"
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
		_, err := New(test.input.loginURL, test.input.apiVersion, credentials.New(test.input.username, test.input.password, test.input.clientID, test.input.clientSecret))
		if err != nil && err.Error() != test.errMsg || err == nil && test.errMsg != "" {
			t.Errorf("TestNew => input: %v; expected: %v; got: %v", test.input, test.errMsg, err)
		}
	}
}

func TestLogin(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		checkLoginFormValues(t, r.Form, "grant_type", "password")
		checkLoginFormValues(t, r.Form, "client_id", "clientid")
		checkLoginFormValues(t, r.Form, "client_secret", "clientsecret")
		checkLoginFormValues(t, r.Form, "username", "user")
		checkLoginFormValues(t, r.Form, "password", "pass")
		io.Copy(w, strings.NewReader(`{"instance_url": "testUrl", "access_token": "testToken"}`))
	}))
	defer s.Close()
	sess := &Session{
		LoginURL:   s.URL,
		creds:      &credentials.OAuth{"user", "pass", "clientid", "clientsecret"},
		httpClient: s.Client(),
	}

	if sess.HasToken() {
		t.Error("TestLogin => should not have token")
	}
	err := sess.Login()
	if err != nil {
		t.Errorf("TestLogin => login failed: %v", err)
	}

	if !sess.HasToken() {
		t.Error("TestLogin => should have token")
	}
	if sess.AccessToken() != "testToken" {
		t.Errorf("TestLogin => wrong access token: expected %s, got %s", "testToken", sess.AccessToken())
	}
	if sess.InstanceURL() != "testUrl" {
		t.Errorf("TestLogin => wrong instance url: expected: %s, got %s", "testUrl", sess.InstanceURL())
	}
}

func checkLoginFormValues(t *testing.T, form url.Values, key, want string) {
	got := form.Get(key)
	if got != want {
		t.Errorf("TestLogin => wrong %s: expected %s, got %s", key, want, got)
	}
}
