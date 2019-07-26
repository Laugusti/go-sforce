package session

import (
	"testing"

	"github.com/Laugusti/salesforce-client/credentials"
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
