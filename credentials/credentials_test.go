package credentials

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
		{},
		{authInput{"user", "pass", "id", "secret"}, OAuth{"user", "pass", "id", "secret"}},
	}

	for _, test := range tests {
		assertMsg := fmt.Sprintf("input: %v", test)
		auth := New(test.input.username, test.input.password, test.input.clientID, test.input.clientSecret)
		assert.Equal(t, &OAuth{test.input.username, test.input.password, test.input.clientID, test.input.clientSecret}, auth, assertMsg)
	}
}
