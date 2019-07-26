package credentials

import "testing"

func TestNew(t *testing.T) {
	type authInput struct {
		username     string
		password     string
		clientID     string
		clientSecret string
	}
	tests := []struct {
		input authInput
		auth  OAuth
	}{
		{authInput{"user", "pass", "id", "secret"}, OAuth{"user", "pass", "id", "secret"}},
	}

	for _, test := range tests {
		auth := New(test.input.username, test.input.password, test.input.clientID, test.input.clientSecret)
		if auth.Username != test.input.username || auth.Password != test.input.password || auth.ClientID != test.input.clientID || auth.ClientSecret != test.input.clientSecret {
			t.Errorf("TestNew => input : %v; expected: %v; got: %v", test.input, test.auth, auth)
		}
	}
}
