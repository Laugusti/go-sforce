package restclient

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/Laugusti/go-sforce/credentials"
	"github.com/Laugusti/go-sforce/internal/testserver"
	"github.com/Laugusti/go-sforce/session"
	"github.com/stretchr/testify/assert"
)

const (
	accessToken = "MOCK_TOKEN"
)

var (
	loginSuccessHandler = func(w http.ResponseWriter, r *http.Request) {
		serverURL := fmt.Sprintf("http://%s", r.Host)
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(session.RequestToken{
			AccessToken: accessToken,
			InstanceURL: serverURL,
		})
	}
	unauthorizedHandler = testserver.StaticJSONHandler(APIError{
		Message:   "Session expired or invalid",
		ErrorCode: "INVALID_SESSION_ID",
	}, http.StatusUnauthorized)
)

func createClientAndServer(t *testing.T) (*Client, *testserver.Server) {
	// start server
	s := testserver.New()
	s.Start()

	// create session and login
	s.HandlerFunc = loginSuccessHandler
	sess := session.Must(session.New(s.URL(), "version", credentials.New("user", "pass", "cid", "csecret")))
	if err := sess.Login(); err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 1, s.RequestCount)
	s.RequestCount = 0 // reset counter

	// create client
	client := &Client{sess, s.Client()}

	return client, s
}

func TestCreateSObject(t *testing.T) {
	client, server := createClientAndServer(t)
	defer server.Stop()

	tests := []struct {
		objectType   string
		object       SObject
		statusCode   int
		requestCount int
		errSnippet   string
	}{
		{"", nil, 400, 0, "sobject name is required"},
		{"Object", nil, 400, 0, "sobject value is required"},
		{"Object", map[string]interface{}{}, 400, 0, "sobject value is required"},
		{"", map[string]interface{}{"Field1": "one", "Field2": 2}, 400, 0, "sobject name is required"},
		{"Object", map[string]interface{}{"Field1": "one", "Field2": 2}, 201, 1, ""},
		{"Object", map[string]interface{}{"Field1": "one", "Field2": 2}, 400, 1, "GENERIC_ERROR"},
	}

	for _, test := range tests {
		// set server response
		server.HandlerFunc = testserver.ValidateJSONBodyHandler(t, test.object,
			UpsertResult{ID: "id", Success: true, Errors: []interface{}{}},
			test.statusCode, fmt.Sprintf("input: %v", test))
		if test.errSnippet != "" {
			server.HandlerFunc = testserver.ValidateJSONBodyHandler(t, test.object,
				APIError{Message: "Generic API error", ErrorCode: "GENERIC_ERROR"},
				test.statusCode, fmt.Sprintf("input: %v", test))
		}

		// do request
		server.RequestCount = 0 // reset counter
		res, err := client.CreateSObject(test.objectType, test.object)
		assert.Equal(t, test.requestCount, server.RequestCount)
		if test.errSnippet != "" {
			assert.NotNilf(t, err, "input %v: expected error", test)
			assert.Containsf(t, err.Error(), test.errSnippet,
				"input %v: wrong error message: %v", test, err)
			assert.Nilf(t, res, "input: %v", test)
		} else {
			assert.Nilf(t, err, "input %v: unexpected error", test)
			assert.NotNilf(t, res, "input: %v", test)
			assert.True(t, res.Success, "input: %v", test)
		}
	}
}
