package restclient

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/Laugusti/go-sforce/credentials"
	"github.com/Laugusti/go-sforce/internal/testserver"
	"github.com/Laugusti/go-sforce/session"
	"github.com/stretchr/testify/assert"
)

const (
	accessToken = "MOCK_TOKEN"
	apiVersion  = "mock"
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
	unauthorizedHandler = testserver.StaticJSONHandler(&testing.T{}, APIError{
		Message:   "Session expired or invalid",
		ErrorCode: "INVALID_SESSION_ID",
	}, http.StatusUnauthorized)

	// api error
	genericApiError = APIError{Message: "Generic API error", ErrorCode: "GENERIC_ERROR"}

	// request validators
	jsonContentTypeValidator = &testserver.HeaderValidator{"Content-Type", "application/json"}
	authTokenValidator       = &testserver.HeaderValidator{"Authorization", "Bearer " + accessToken}
	emptyQueryValidator      = &testserver.QueryValidator{url.Values{}}
	emptyBodyValidator       = &testserver.JSONBodyValidator{nil}
)

func createClientAndServer(t *testing.T) (*Client, *testserver.Server) {
	// start server
	s := testserver.New(t)

	// create session and login
	s.HandlerFunc = loginSuccessHandler
	sess := session.Must(session.New(s.URL(), apiVersion, credentials.New("user", "pass", "cid", "csecret")))
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
		{"", nil, 0, 0, "sobject name is required"},
		{"Object", nil, 0, 0, "sobject value is required"},
		{"Object", map[string]interface{}{}, 0, 0, "sobject value is required"},
		{"", map[string]interface{}{"Field1": "one", "Field2": 2}, 0, 0, "sobject name is required"},
		{"Object", map[string]interface{}{"Field1": "one", "Field2": 2}, 201, 1, ""},
		{"Object", map[string]interface{}{"Field1": "one", "Field2": 2}, 400, 1, "GENERIC_ERROR"},
	}

	for _, test := range tests {
		assertMessage := fmt.Sprintf("input: %v", test)
		path := fmt.Sprintf("/services/data/%s/sobjects/%s", apiVersion, test.objectType)
		validators := []testserver.RequestValidator{authTokenValidator, jsonContentTypeValidator,
			emptyQueryValidator, &testserver.JSONBodyValidator{test.object},
			&testserver.PathValidator{path}}

		// set server response
		setHandlerFunc(server, t, assertMessage,
			UpsertResult{ID: "id", Success: true, Errors: []interface{}{}},
			test.statusCode, test.errSnippet != "", validators...)

		// do request
		server.RequestCount = 0 // reset counter
		res, err := client.CreateSObject(test.objectType, test.object)

		// assertions
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

func TestGetSObject(t *testing.T) {
	client, server := createClientAndServer(t)
	defer server.Stop()

	tests := []struct {
		objectType   string
		objectID     string
		statusCode   int
		requestCount int
		errSnippet   string
		wantedObject SObject
	}{
		{"", "", 0, 0, "sobject name is required", nil},
		{"", "A", 0, 0, "sobject name is required", nil},
		{"Object", "", 0, 0, "sobject id is required", nil},
		{"Object", "A", 200, 1, "", map[string]interface{}{"A": "one", "B": 2.0, "C": true}},
		{"Object", "A", 400, 1, "GENERIC_ERROR", nil},
	}

	for _, test := range tests {
		assertMessage := fmt.Sprintf("input: %v", test)
		path := fmt.Sprintf("/services/data/%s/sobjects/%s/%s", apiVersion, test.objectType,
			test.objectID)
		validators := []testserver.RequestValidator{authTokenValidator, jsonContentTypeValidator,
			emptyQueryValidator, emptyBodyValidator, &testserver.PathValidator{path}}

		// set server response
		setHandlerFunc(server, t, assertMessage, test.wantedObject, test.statusCode,
			test.errSnippet != "", validators...)

		// do request
		server.RequestCount = 0 // reset counter
		sObj, err := client.GetSObject(test.objectType, test.objectID)

		// assertions
		assert.Equal(t, test.requestCount, server.RequestCount)
		if test.errSnippet != "" {
			assert.NotNilf(t, err, "input %v: expected error", test)
			assert.Containsf(t, err.Error(), test.errSnippet,
				"input %v: wrong error message: %v", test, err)
			assert.Nilf(t, sObj, "input: %v", test)
		} else {
			assert.Nilf(t, err, "input %v: unexpected error", test)
			assert.Equalf(t, test.wantedObject, sObj, "input: %v", test)
		}

	}
}

func TestGetSObjectByExternalID(t *testing.T) {
	client, server := createClientAndServer(t)
	defer server.Stop()

	tests := []struct {
		objectType      string
		externalIDField string
		externalID      string
		wantedObject    SObject
		statusCode      int
		requestCount    int
		errSnippet      string
	}{
		{"", "", "", nil, 0, 0, "sobject name is required"},
		{"", "A", "", nil, 0, 0, "sobject name is required"},
		{"", "A", "a", nil, 0, 0, "sobject name is required"},
		{"Object", "", "", nil, 0, 0, "external id field is required"},
		{"Object", "", "a", nil, 0, 0, "external id field is required"},
		{"Object", "A", "", nil, 0, 0, "external id is required"},
		{"Object", "A", "a", map[string]interface{}{"A": "one", "B": 2.0, "C": true}, 200, 1, ""},
		{"Object", "A", "a", nil, 400, 1, "GENERIC_ERROR"},
	}

	for _, test := range tests {
		assertMessage := fmt.Sprintf("input: %v", test)
		path := fmt.Sprintf("/services/data/%s/sobjects/%s/%s/%s", apiVersion, test.objectType,
			test.externalIDField, test.externalID)
		validators := []testserver.RequestValidator{authTokenValidator, jsonContentTypeValidator,
			emptyQueryValidator, emptyBodyValidator, &testserver.PathValidator{path}}

		// set server response
		setHandlerFunc(server, t, assertMessage, test.wantedObject, test.statusCode,
			test.errSnippet != "", validators...)

		// do request
		server.RequestCount = 0 // reset counter
		sObj, err := client.GetSObjectByExternalID(test.objectType,
			test.externalIDField, test.externalID)

		// assertions
		assert.Equal(t, test.requestCount, server.RequestCount)
		if test.errSnippet != "" {
			assert.NotNilf(t, err, "input %v: expected error", test)
			assert.Containsf(t, err.Error(), test.errSnippet,
				"input %v: wrong error message: %v", test, err)
			assert.Nilf(t, sObj, "input: %v", test)
		} else {
			assert.Nilf(t, err, "input %v: unexpected error", test)
			assert.Equalf(t, test.wantedObject, sObj, "input: %v", test)
		}

	}
}

func TestUpsertSObject(t *testing.T) {
	client, server := createClientAndServer(t)
	defer server.Stop()

	tests := []struct {
		objectType   string
		objectID     string
		object       SObject
		statusCode   int
		requestCount int
		errSnippet   string
	}{
		{"", "", nil, 0, 0, "sobject name is required"},
		{"", "A", nil, 0, 0, "sobject name is required"},
		{"", "A", map[string]interface{}{"A": "one", "B": 2}, 0, 0, "sobject name is required"},
		{"Object", "", nil, 0, 0, "sobject id is required"},
		{"Object", "", map[string]interface{}{"A": "one", "B": 2}, 0, 0, "sobject id is required"},
		{"Object", "A", nil, 0, 0, "sobject value is required"},
		{"Object", "A", map[string]interface{}{}, 0, 0, "sobject value is required"},
		{"Object", "A", map[string]interface{}{"A": "one", "B": 2}, 200, 1, ""},
		{"Object", "A", map[string]interface{}{"A": "one", "B": 2}, 400, 1, "GENERIC_ERROR"},
	}

	for _, test := range tests {
		assertMessage := fmt.Sprintf("input: %v", test)
		path := fmt.Sprintf("/services/data/%s/sobjects/%s/%s", apiVersion, test.objectType,
			test.objectID)
		validators := []testserver.RequestValidator{authTokenValidator, jsonContentTypeValidator,
			emptyQueryValidator, &testserver.JSONBodyValidator{test.object},
			&testserver.PathValidator{path}}

		// set server response
		setHandlerFunc(server, t, assertMessage,
			UpsertResult{ID: test.objectID, Success: true, Errors: []interface{}{}},
			test.statusCode, test.errSnippet != "", validators...)

		// do request
		server.RequestCount = 0 // reset counter
		res, err := client.UpsertSObject(test.objectType, test.objectID, test.object)

		// assertions
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
			assert.Equal(t, test.objectID, res.ID)
		}
	}
}

func TestUnauthorizedClient(t *testing.T) {
	client, server := createClientAndServer(t)
	defer server.Stop()
	// server handler return 401

	server.HandlerFunc = unauthorizedHandler
	_, err := client.CreateSObject("Object", map[string]interface{}{"A": "B"})
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "INVALID_SESSION_ID")
	// 2 requests (create POST and login POST)
	assert.Equal(t, 2, server.RequestCount)

	server.RequestCount = 0 // reset counter
	// 1st request fails, 2nd returns login, other return upsert result
	server.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
		switch server.RequestCount {
		case 0:
			t.Error("request count can't be 0")
		case 1:
			unauthorizedHandler(w, r)
		case 2:
			loginSuccessHandler(w, r)
		default:
			testserver.StaticJSONHandler(t, UpsertResult{"id", true, nil}, http.StatusCreated)(w, r)
		}
	}
	_, err = client.CreateSObject("Object", map[string]interface{}{"A": "B"})
	assert.Nil(t, err, "client request should've succeeded")
	// 3 requests (create POST and login POST and retry create POST)
	assert.Equal(t, 3, server.RequestCount)
}

func setHandlerFunc(server *testserver.Server, t *testing.T, assertMessage string,
	responseBody interface{}, responseStatus int, apiError bool, validators ...testserver.RequestValidator) {
	if apiError {
		responseBody = genericApiError
	}

	// set server response
	server.HandlerFunc = testserver.ValidateAndSetResponseHandler(t, assertMessage,
		responseBody, responseStatus, validators...)
}
