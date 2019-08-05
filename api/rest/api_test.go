package restapi

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/Laugusti/go-sforce/internal/testserver"
	"github.com/Laugusti/go-sforce/sforce/credentials"
	"github.com/Laugusti/go-sforce/sforce/session"
	"github.com/Laugusti/go-sforce/sforce/sforceerr"
	"github.com/stretchr/testify/assert"
)

const (
	accessToken = "MOCK_TOKEN"
	apiVersion  = "mock"
)

var (
	unauthorizedHandler = &testserver.JSONResponseHandler{
		StatusCode: http.StatusUnauthorized,
		Body: session.LoginError{
			Message:   "Session expired or invalid",
			ErrorCode: "INVALID_SESSION_ID"},
	}

	// api error
	genericErr = sforceerr.APIError{Message: "Generic API error", ErrorCode: "GENERIC_ERROR"}

	// request validators
	jsonContentTypeValidator = &testserver.HeaderValidator{Key: "Content-Type", Value: "application/json"}
	authTokenValidator       = &testserver.HeaderValidator{Key: "Authorization", Value: "Bearer " + accessToken}
	emptyQueryValidator      = &testserver.QueryValidator{Query: url.Values{}}
	emptyBodyValidator       = &testserver.JSONBodyValidator{Body: nil}
	getMethodValidator       = &testserver.MethodValidator{Method: http.MethodGet}
	postMethodValidator      = &testserver.MethodValidator{Method: http.MethodPost}
	patchMethodValidator     = &testserver.MethodValidator{Method: http.MethodPatch}
	deleteMethodValidator    = &testserver.MethodValidator{Method: http.MethodDelete}
)

func createClientAndServer(t *testing.T) (*Client, *testserver.Server) {
	// start server
	s := testserver.New(t)

	// create session and login
	s.HandlerFunc = testserver.StaticJSONHandlerFunc(t, http.StatusOK,
		session.RequestToken{
			AccessToken: accessToken,
			InstanceURL: s.URL(),
		})
	sess := session.Must(session.New(
		s.URL(),
		apiVersion,
		credentials.New("user", "pass", "cid", "csecret"),
	))
	if err := sess.Login(); err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 1, s.RequestCount, "expected single request (login)")
	s.RequestCount = 0 // reset counter

	// create client
	client := &Client{sess}

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
		{"", nil, 0, 0, "invalid sobject name"},
		{"Object", nil, 0, 0, "sobject value is required"},
		{"Object", map[string]interface{}{}, 0, 0, "sobject value is required"},
		{"", map[string]interface{}{"Field1": "one", "Field2": 2}, 0, 0, "invalid sobject name"},
		{"Object", map[string]interface{}{"Field1": "one", "Field2": 2}, 201, 1, ""},
		{"Object", map[string]interface{}{"Field1": "one", "Field2": 2}, 400, 1, "GENERIC_ERROR"},
	}

	for _, test := range tests {
		assertMsg := fmt.Sprintf("input: %v", test)
		path := fmt.Sprintf("/services/data/%s/sobjects/%s", apiVersion, test.objectType)
		validators := []testserver.RequestValidator{authTokenValidator, jsonContentTypeValidator,
			emptyQueryValidator, &testserver.JSONBodyValidator{Body: test.object},
			&testserver.PathValidator{Path: path}, postMethodValidator}
		requestFunc := func() (interface{}, error) {
			return client.CreateSObject(&CreateSObjectInput{
				SObjectName: test.objectType,
				SObject:     test.object,
			})
		}
		successFunc := func(res interface{}) {
			out, ok := res.(*CreateSObjectOutput)
			if assert.True(t, ok, assertMsg) &&
				assert.NotNil(t, out, assertMsg) {
				assert.True(t, out.Result.Success,
					assertMsg)
			}
		}
		handler := &testserver.JSONResponseHandler{
			StatusCode: test.statusCode,
			Body:       UpsertResult{ID: "id", Success: true, Errors: []interface{}{}},
		}

		assertRequest(t, assertMsg, server, test.errSnippet, requestFunc, successFunc,
			test.requestCount, validators, handler)
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
		{"", "", 0, 0, "invalid sobject name", nil},
		{"", "A", 0, 0, "invalid sobject name", nil},
		{"Object", "", 0, 0, "sobject id is required", nil},
		{"Object", "A", 200, 1, "", map[string]interface{}{"A": "one", "B": 2.0, "C": true}},
		{"Object", "A", 400, 1, "GENERIC_ERROR", nil},
	}

	for _, test := range tests {
		assertMsg := fmt.Sprintf("input: %v", test)
		path := fmt.Sprintf("/services/data/%s/sobjects/%s/%s", apiVersion, test.objectType,
			test.objectID)
		validators := []testserver.RequestValidator{authTokenValidator, jsonContentTypeValidator,
			emptyQueryValidator, emptyBodyValidator, &testserver.PathValidator{Path: path},
			getMethodValidator}

		requestFunc := func() (interface{}, error) {
			return client.GetSObject(&GetSObjectInput{
				SObjectName: test.objectType,
				SObjectID:   test.objectID,
			})
		}
		successFunc := func(res interface{}) {
			out, ok := res.(*GetSObjectOutput)
			if assert.True(t, ok, assertMsg) &&
				assert.NotNil(t, out, assertMsg) {
				assert.Equal(t, test.wantedObject, out.SObject, assertMsg)
			}
		}
		handler := &testserver.JSONResponseHandler{
			StatusCode: test.statusCode,
			Body:       test.wantedObject,
		}

		assertRequest(t, assertMsg, server, test.errSnippet, requestFunc, successFunc,
			test.requestCount, validators, handler)
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
		{"", "", "", nil, 0, 0, "invalid sobject name"},
		{"", "A", "", nil, 0, 0, "invalid sobject name"},
		{"", "A", "a", nil, 0, 0, "invalid sobject name"},
		{"Object", "", "", nil, 0, 0, "invalid external id field"},
		{"Object", "", "a", nil, 0, 0, "invalid external id field"},
		{"Object", "A", "", nil, 0, 0, "external id is required"},
		{"Object", "A", "a", map[string]interface{}{"A": "one", "B": 2.0, "C": true}, 200, 1, ""},
		{"Object", "A", "a", nil, 400, 1, "GENERIC_ERROR"},
	}

	for _, test := range tests {
		assertMsg := fmt.Sprintf("input: %v", test)
		path := fmt.Sprintf("/services/data/%s/sobjects/%s/%s/%s", apiVersion, test.objectType,
			test.externalIDField, test.externalID)
		validators := []testserver.RequestValidator{authTokenValidator, jsonContentTypeValidator,
			emptyQueryValidator, emptyBodyValidator, &testserver.PathValidator{Path: path},
			getMethodValidator}

		requestFunc := func() (interface{}, error) {
			return client.GetSObjectByExternalID(&GetSObjectByExternalIDInput{
				SObjectName:     test.objectType,
				ExternalIDField: test.externalIDField,
				ExternalID:      test.externalID,
			})
		}
		successFunc := func(res interface{}) {
			out, ok := res.(*GetSObjectByExternalIDOutput)
			if assert.True(t, ok, assertMsg) &&
				assert.NotNil(t, out, assertMsg) {
				assert.Equal(t, test.wantedObject, out.SObject, assertMsg)
			}
		}
		handler := &testserver.JSONResponseHandler{
			StatusCode: test.statusCode,
			Body:       test.wantedObject,
		}

		assertRequest(t, assertMsg, server, test.errSnippet, requestFunc, successFunc,
			test.requestCount, validators, handler)
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
		{"", "", nil, 0, 0, "invalid sobject name"},
		{"", "A", nil, 0, 0, "invalid sobject name"},
		{"", "A", map[string]interface{}{"A": "one", "B": 2}, 0, 0, "invalid sobject name"},
		{"Object", "", nil, 0, 0, "sobject id is required"},
		{"Object", "", map[string]interface{}{"A": "one", "B": 2}, 0, 0, "sobject id is required"},
		{"Object", "A", nil, 0, 0, "sobject value is required"},
		{"Object", "A", map[string]interface{}{}, 0, 0, "sobject value is required"},
		{"Object", "A", map[string]interface{}{"A": "one", "B": 2}, 200, 1, ""},
		{"Object", "A", map[string]interface{}{"A": "one", "B": 2}, 400, 1, "GENERIC_ERROR"},
	}

	for _, test := range tests {
		assertMsg := fmt.Sprintf("input: %v", test)
		path := fmt.Sprintf("/services/data/%s/sobjects/%s/%s", apiVersion, test.objectType,
			test.objectID)
		validators := []testserver.RequestValidator{authTokenValidator, jsonContentTypeValidator,
			emptyQueryValidator, &testserver.JSONBodyValidator{Body: test.object},
			&testserver.PathValidator{Path: path}, patchMethodValidator}

		requestFunc := func() (interface{}, error) {
			return client.UpsertSObject(&UpsertSObjectInput{
				SObjectName: test.objectType,
				SObjectID:   test.objectID,
				SObject:     test.object,
			})
		}
		successFunc := func(res interface{}) {
			if assert.NotNil(t, res, assertMsg) {
				out, ok := res.(*UpsertSObjectOutput)
				if assert.True(t, ok, assertMsg) &&
					assert.NotNil(t, out, assertMsg) {
					assert.True(t, out.Result.Success, assertMsg)
					assert.Equal(t, test.objectID, out.Result.ID,
						assertMsg)
				}
			}
		}
		handler := &testserver.JSONResponseHandler{
			StatusCode: test.statusCode,
			Body:       UpsertResult{ID: test.objectID, Success: true, Errors: []interface{}{}},
		}

		assertRequest(t, assertMsg, server, test.errSnippet, requestFunc, successFunc,
			test.requestCount, validators, handler)
	}
}

func TestUpsertSObjectByExternalID(t *testing.T) {
	client, server := createClientAndServer(t)
	defer server.Stop()

	tests := []struct {
		objectType      string
		externalIDField string
		externalID      string
		object          SObject
		statusCode      int
		requestCount    int
		errSnippet      string
	}{
		{"", "", "", nil, 0, 0, "invalid sobject name"},
		{"", "A", "", nil, 0, 0, "invalid sobject name"},
		{"", "A", "a", nil, 0, 0, "invalid sobject name"},
		{"", "A", "a", map[string]interface{}{"A": "one", "B": 2}, 0, 0, "invalid sobject name"},
		{"Object", "", "", nil, 0, 0, "invalid external id field"},
		{"Object", "", "a", nil, 0, 0, "invalid external id field"},
		{"Object", "", "a", map[string]interface{}{"A": "one", "B": 2}, 0, 0, "invalid external id field"},
		{"Object", "A", "", nil, 0, 0, "external id is required"},
		{"Object", "A", "", map[string]interface{}{"A": "one", "B": 2}, 0, 0, "external id is required"},
		{"Object", "A", "a", nil, 0, 0, "sobject value is required"},
		{"Object", "A", "a", map[string]interface{}{}, 0, 0, "sobject value is required"},
		{"Object", "A", "a", map[string]interface{}{"A": "one", "B": 2}, 200, 1, ""},
		{"Object", "A", "a", map[string]interface{}{"A": "one", "B": 2}, 400, 1, "GENERIC_ERROR"},
	}

	for _, test := range tests {
		assertMsg := fmt.Sprintf("input: %v", test)
		path := fmt.Sprintf("/services/data/%s/sobjects/%s/%s/%s", apiVersion, test.objectType,
			test.externalIDField, test.externalID)
		validators := []testserver.RequestValidator{authTokenValidator, jsonContentTypeValidator,
			emptyQueryValidator, &testserver.JSONBodyValidator{Body: test.object},
			&testserver.PathValidator{Path: path}, patchMethodValidator}

		requestFunc := func() (interface{}, error) {
			return client.UpsertSObjectByExternalID(
				&UpsertSObjectByExternalIDInput{
					SObjectName:     test.objectType,
					ExternalIDField: test.externalIDField,
					ExternalID:      test.externalID,
					SObject:         test.object,
				})
		}
		successFunc := func(res interface{}) {
			out, ok := res.(*UpsertSObjectByExternalIDOutput)
			if assert.True(t, ok, assertMsg) &&
				assert.NotNil(t, res, assertMsg) {
				assert.True(t, out.Result.Success, assertMsg)
			}
		}
		handler := &testserver.JSONResponseHandler{
			StatusCode: test.statusCode,
			Body:       UpsertResult{ID: "id", Success: true, Errors: []interface{}{}},
		}

		assertRequest(t, assertMsg, server, test.errSnippet, requestFunc, successFunc,
			test.requestCount, validators, handler)
	}
}

func TestDeleteSObject(t *testing.T) {
	client, server := createClientAndServer(t)
	defer server.Stop()

	tests := []struct {
		objectType   string
		objectID     string
		statusCode   int
		requestCount int
		errSnippet   string
	}{
		{"", "", 0, 0, "invalid sobject name"},
		{"", "A", 0, 0, "invalid sobject name"},
		{"Object", "", 0, 0, "sobject id is required"},
		{"Object", "A", 204, 1, ""},
		{"Object", "A", 400, 1, "GENERIC_ERROR"},
	}

	for _, test := range tests {
		assertMsg := fmt.Sprintf("input: %v", test)
		path := fmt.Sprintf("/services/data/%s/sobjects/%s/%s", apiVersion, test.objectType,
			test.objectID)
		validators := []testserver.RequestValidator{authTokenValidator, jsonContentTypeValidator,
			emptyQueryValidator, emptyBodyValidator, &testserver.PathValidator{Path: path},
			deleteMethodValidator}

		requestFunc := func() (interface{}, error) {
			return client.DeleteSObject(&DeleteSObjectInput{
				SObjectName: test.objectType,
				SObjectID:   test.objectID,
			})
		}
		successFunc := func(res interface{}) {
			assert.Equal(t, res, &DeleteSObjectOutput{}, assertMsg)
		}
		handler := &testserver.JSONResponseHandler{
			StatusCode: test.statusCode,
			Body:       nil,
		}

		assertRequest(t, assertMsg, server, test.errSnippet, requestFunc, successFunc,
			test.requestCount, validators, handler)
	}
}

func TestQuery(t *testing.T) {
	client, server := createClientAndServer(t)
	defer server.Stop()

	tests := []struct {
		query        string
		statusCode   int
		requestCount int
		errSnippet   string
		done         bool
		totalSize    int
		nextURL      string
		records      []SObject
	}{
		{"", 0, 0, "query string is required", false, 0, "", nil},
		{"query", 200, 1, "", true, 1, "", []SObject{map[string]interface{}{"A": 1.0, "B": "two", "C": false}}},
		{"query", 400, 1, "GENERIC_ERROR", true, 1, "", []SObject{}},
		{"query", 200, 1, "", false, 3, server.URL(), []SObject{map[string]interface{}{"A": "a"}, map[string]interface{}{"A": "1"}}},
	}

	for _, test := range tests {
		assertMsg := fmt.Sprintf("input: %v", test)
		path := fmt.Sprintf("/services/data/%s/query", apiVersion)
		q, err := url.ParseQuery("q=" + test.query)
		assert.Nil(t, err, assertMsg)
		validators := []testserver.RequestValidator{authTokenValidator, jsonContentTypeValidator,
			&testserver.QueryValidator{Query: q}, emptyBodyValidator,
			&testserver.PathValidator{Path: path}, getMethodValidator}

		want := &QueryResult{
			Done:           test.done,
			NextRecordsURL: test.nextURL,
			TotalSize:      test.totalSize,
			Records:        test.records,
		}
		requestFunc := func() (interface{}, error) {
			return client.Query(&QueryInput{
				Query: test.query,
			})
		}
		successFunc := func(res interface{}) {
			out, ok := res.(*QueryOutput)
			if assert.True(t, ok, assertMsg) && assert.NotNil(t, out, assertMsg) {
				assert.Equal(t, want, out.Result, assertMsg)
			}
		}
		handler := &testserver.JSONResponseHandler{
			StatusCode: test.statusCode,
			Body:       want,
		}

		assertRequest(t, assertMsg, server, test.errSnippet, requestFunc, successFunc,
			test.requestCount, validators, handler)
	}
}

func TestQueryMore(t *testing.T) {
	client, server := createClientAndServer(t)
	defer server.Stop()

	tests := []struct {
		nextRecordsURL string
		statusCode     int
		requestCount   int
		errSnippet     string
		done           bool
		totalSize      int
		nextURL        string
		records        []SObject
	}{
		{"", 0, 0, "missing next records url", false, 0, "", nil},
		{"next/records/url", 200, 1, "", true, 1, "", []SObject{map[string]interface{}{"A": 1.0, "B": "two", "C": false}}},
		{"nextrecordsurl", 400, 1, "GENERIC_ERROR", true, 1, "", []SObject{}},
		{"nru", 200, 1, "", false, 3, server.URL(), []SObject{map[string]interface{}{"A": "a"}, map[string]interface{}{"A": "1"}}},
	}

	for _, test := range tests {
		assertMsg := fmt.Sprintf("input: %v", test)
		validators := []testserver.RequestValidator{authTokenValidator, jsonContentTypeValidator,
			emptyQueryValidator, emptyBodyValidator,
			&testserver.PathValidator{Path: test.nextRecordsURL},
			getMethodValidator}

		want := &QueryResult{
			Done:           test.done,
			NextRecordsURL: test.nextURL,
			TotalSize:      test.totalSize,
			Records:        test.records,
		}
		requestFunc := func() (interface{}, error) {
			return client.QueryMore(&QueryMoreInput{
				NextRecordsURL: test.nextRecordsURL,
			})
		}
		successFunc := func(res interface{}) {
			out, ok := res.(*QueryMoreOutput)
			if assert.True(t, ok, assertMsg) &&
				assert.NotNil(t, out, assertMsg) {
				assert.Equal(t, want, out.Result, assertMsg)
			}
		}
		handler := &testserver.JSONResponseHandler{
			StatusCode: test.statusCode,
			Body:       want,
		}

		assertRequest(t, assertMsg, server, test.errSnippet, requestFunc, successFunc,
			test.requestCount, validators, handler)
	}
}

func TestUnauthorizedClient(t *testing.T) {
	client, server := createClientAndServer(t)
	defer server.Stop()

	loginHandlerFunc := server.HandlerFunc
	// server handler return 401
	server.HandlerFunc = testserver.ValidateRequestHandlerFunc(t, "", unauthorizedHandler)

	_, err := client.CreateSObject(&CreateSObjectInput{
		SObjectName: "Object",
		SObject:     map[string]interface{}{"A": "B"},
	})
	assert.NotNil(t, err, "expected client error")
	assert.Contains(t, err.Error(), "INVALID_SESSION_ID", "expected invalid session response")
	assert.Equal(t, 2, server.RequestCount, "expected 2 request (create and login)")

	// reauthenticate session
	server.HandlerFunc = loginHandlerFunc
	assert.Nil(t, client.sess.Login(), "login failed")

	server.RequestCount = 0 // reset counter
	// 1st request fails, 2nd returns access token, others return upsert result
	server.HandlerFunc = testserver.ValidateRequestHandlerFunc(t, "",
		&testserver.ConsecutiveResponseHandler{
			Handlers: []testserver.ResponseHandler{
				unauthorizedHandler,
				&testserver.JSONResponseHandler{
					StatusCode: http.StatusOK,
					Body: session.RequestToken{
						AccessToken: accessToken,
						InstanceURL: server.URL(),
					},
				},
				&testserver.JSONResponseHandler{
					StatusCode: http.StatusCreated,
					Body:       UpsertResult{"id", true, nil},
				},
			},
		})
	_, err = client.CreateSObject(&CreateSObjectInput{
		SObjectName: "Object",
		SObject:     map[string]interface{}{"A": "B"},
	})
	assert.Nil(t, err, "client request should've succeeded")
	// 3 requests (create POST and login POST and retry create POST)
	assert.Equal(t, 3, server.RequestCount, "expected 3 requests (create, login, retry)")
}

func assertRequest(t *testing.T, assertMsg string, server *testserver.Server, wantErr string,
	invokeFunc func() (interface{}, error), successFunc func(interface{}),
	expectedRequestCount int, validators []testserver.RequestValidator,
	respHandler *testserver.JSONResponseHandler) {
	shouldErr := wantErr != ""
	// set server response
	if shouldErr {
		respHandler.Body = genericErr
	}
	server.HandlerFunc = testserver.ValidateRequestHandlerFunc(t, assertMsg, respHandler, validators...)

	// invoke request
	server.RequestCount = 0 // reset counter
	out, err := invokeFunc()

	// assertions
	assert.Equal(t, expectedRequestCount, server.RequestCount, assertMsg)
	if shouldErr {
		if assert.Error(t, err, assertMsg) {
			assert.Contains(t, err.Error(), wantErr, assertMsg)
		}
	} else {
		assert.Nil(t, err, assertMsg)
		if successFunc != nil {
			successFunc(out)
		}
	}
}
