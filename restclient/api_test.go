package restclient

import (
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
	unauthorizedHandler = &testserver.JSONResponseHandler{
		StatusCode: http.StatusUnauthorized,
		Body: APIError{
			Message:   "Session expired or invalid",
			ErrorCode: "INVALID_SESSION_ID"},
	}

	// api error
	genericErr = APIError{Message: "Generic API error", ErrorCode: "GENERIC_ERROR"}

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
	sess := session.Must(session.New(s.URL(), apiVersion, credentials.New("user", "pass", "cid", "csecret")))
	if err := sess.Login(); err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 1, s.RequestCount, "expected single request (login)")
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
		assertMsg := fmt.Sprintf("input: %v", test)
		path := fmt.Sprintf("/services/data/%s/sobjects/%s", apiVersion, test.objectType)
		validators := []testserver.RequestValidator{authTokenValidator, jsonContentTypeValidator,
			emptyQueryValidator, &testserver.JSONBodyValidator{Body: test.object},
			&testserver.PathValidator{Path: path}, postMethodValidator}
		requestFunc := func() (interface{}, error) {
			return client.CreateSObject(test.objectType, test.object)
		}
		successFunc := func(res interface{}) {
			if assert.NotNil(t, res, assertMsg) {
				assert.True(t, res.(*UpsertResult).Success, assertMsg)
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
		{"", "", 0, 0, "sobject name is required", nil},
		{"", "A", 0, 0, "sobject name is required", nil},
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
			return client.GetSObject(test.objectType, test.objectID)
		}
		successFunc := func(res interface{}) {
			if assert.NotNil(t, res, assertMsg) {
				assert.Equal(t, test.wantedObject, res, assertMsg)
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
		assertMsg := fmt.Sprintf("input: %v", test)
		path := fmt.Sprintf("/services/data/%s/sobjects/%s/%s/%s", apiVersion, test.objectType,
			test.externalIDField, test.externalID)
		validators := []testserver.RequestValidator{authTokenValidator, jsonContentTypeValidator,
			emptyQueryValidator, emptyBodyValidator, &testserver.PathValidator{Path: path},
			getMethodValidator}

		requestFunc := func() (interface{}, error) {
			return client.GetSObjectByExternalID(test.objectType,
				test.externalIDField, test.externalID)
		}
		successFunc := func(res interface{}) {
			if assert.NotNil(t, res, assertMsg) {
				assert.Equal(t, test.wantedObject, res, assertMsg)
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
		assertMsg := fmt.Sprintf("input: %v", test)
		path := fmt.Sprintf("/services/data/%s/sobjects/%s/%s", apiVersion, test.objectType,
			test.objectID)
		validators := []testserver.RequestValidator{authTokenValidator, jsonContentTypeValidator,
			emptyQueryValidator, &testserver.JSONBodyValidator{Body: test.object},
			&testserver.PathValidator{Path: path}, patchMethodValidator}

		requestFunc := func() (interface{}, error) {
			return client.UpsertSObject(test.objectType, test.objectID, test.object)
		}
		successFunc := func(res interface{}) {
			if assert.NotNil(t, res, assertMsg) {
				assert.True(t, res.(*UpsertResult).Success, assertMsg)
				assert.Equal(t, test.objectID, res.(*UpsertResult).ID, assertMsg)
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
		{"", "", "", nil, 0, 0, "sobject name is required"},
		{"", "A", "", nil, 0, 0, "sobject name is required"},
		{"", "A", "a", nil, 0, 0, "sobject name is required"},
		{"", "A", "a", map[string]interface{}{"A": "one", "B": 2}, 0, 0, "sobject name is required"},
		{"Object", "", "", nil, 0, 0, "external id field is required"},
		{"Object", "", "a", nil, 0, 0, "external id field is required"},
		{"Object", "", "a", map[string]interface{}{"A": "one", "B": 2}, 0, 0, "external id field is required"},
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
			return client.UpsertSObjectByExternalID(test.objectType, test.externalIDField,
				test.externalID, test.object)
		}
		successFunc := func(res interface{}) {
			if assert.NotNil(t, res, assertMsg) {
				assert.True(t, res.(*UpsertResult).Success, assertMsg)
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
		{"", "", 0, 0, "sobject name is required"},
		{"", "A", 0, 0, "sobject name is required"},
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
			return nil, client.DeleteSObject(test.objectType, test.objectID)
		}
		successFunc := func(res interface{}) {
			assert.Nil(t, res, assertMsg)
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
			return client.Query(test.query)
		}
		successFunc := func(res interface{}) {
			assert.Equal(t, want, res, assertMsg)
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
		queryResult  *QueryResult
		statusCode   int
		requestCount int
		errSnippet   string
		done         bool
		totalSize    int
		nextURL      string
		records      []SObject
	}{
		{nil, 0, 0, "missing next records url", false, 0, "", nil},
		{&QueryResult{}, 0, 0, "missing next records url", false, 0, "", nil},
		{&QueryResult{NextRecordsURL: "next/records/url"}, 200, 1, "", true, 1, "", []SObject{map[string]interface{}{"A": 1.0, "B": "two", "C": false}}},
		{&QueryResult{NextRecordsURL: "nextrecordsurl"}, 400, 1, "GENERIC_ERROR", true, 1, "", []SObject{}},
		{&QueryResult{NextRecordsURL: "nru"}, 200, 1, "", false, 3, server.URL(), []SObject{map[string]interface{}{"A": "a"}, map[string]interface{}{"A": "1"}}},
	}

	for _, test := range tests {
		assertMsg := fmt.Sprintf("input: %v", test)
		path := ""
		if test.queryResult != nil {
			path = test.queryResult.NextRecordsURL
		}
		validators := []testserver.RequestValidator{authTokenValidator, jsonContentTypeValidator,
			emptyQueryValidator, emptyBodyValidator,
			&testserver.PathValidator{Path: path}, getMethodValidator}

		want := &QueryResult{
			Done:           test.done,
			NextRecordsURL: test.nextURL,
			TotalSize:      test.totalSize,
			Records:        test.records,
		}
		requestFunc := func() (interface{}, error) {
			return client.QueryMore(test.queryResult)
		}
		successFunc := func(res interface{}) {
			assert.Equal(t, want, res, assertMsg)
		}
		handler := &testserver.JSONResponseHandler{
			StatusCode: test.statusCode,
			Body:       want,
		}

		assertRequest(t, assertMsg, server, test.errSnippet, requestFunc, successFunc,
			test.requestCount, validators, handler)
	}
}

func TestFullQuery(t *testing.T) {
	client, server := createClientAndServer(t)
	defer server.Stop()

	want := &QueryResult{
		Done:           true,
		NextRecordsURL: "",
		TotalSize:      3,
		Records: []SObject{
			map[string]interface{}{"A": "a"},
			map[string]interface{}{"B": true},
			map[string]interface{}{"C": 3.0},
		},
	}

	validators := []testserver.RequestValidator{authTokenValidator, jsonContentTypeValidator,
		emptyBodyValidator, getMethodValidator}

	handlers := []testserver.ResponseHandler{
		&testserver.JSONResponseHandler{
			StatusCode: http.StatusOK,
			Body: &QueryResult{
				Done:           false,
				NextRecordsURL: "batch2",
				TotalSize:      3,
				Records:        []SObject{map[string]interface{}{"A": "a"}},
			},
		},
		&testserver.JSONResponseHandler{
			StatusCode: http.StatusOK,
			Body: &QueryResult{
				Done:           false,
				NextRecordsURL: "batch3",
				TotalSize:      3,
				Records:        []SObject{map[string]interface{}{"B": true}},
			},
		},
		&testserver.JSONResponseHandler{
			StatusCode: http.StatusOK,
			Body: &QueryResult{
				Done:           true,
				NextRecordsURL: "",
				TotalSize:      3,
				Records:        []SObject{map[string]interface{}{"C": 3.0}},
			},
		},
	}

	server.HandlerFunc = testserver.ValidateRequestHandlerFunc(t, "",
		&testserver.ConsecutiveResponseHandler{Handlers: handlers},
		validators...)

	res, err := client.FullQuery("query")
	assert.Nil(t, err)
	assert.Equal(t, want, res)
}

func TestUnauthorizedClient(t *testing.T) {
	client, server := createClientAndServer(t)
	defer server.Stop()
	// server handler return 401
	server.HandlerFunc = testserver.ValidateRequestHandlerFunc(t, "", unauthorizedHandler)

	_, err := client.CreateSObject("Object", map[string]interface{}{"A": "B"})
	assert.NotNil(t, err, "expected client error")
	assert.Contains(t, err.Error(), "INVALID_SESSION_ID", "expected invalid session response")
	assert.Equal(t, 2, server.RequestCount, "expected 2 request (create and login)")

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
	_, err = client.CreateSObject("Object", map[string]interface{}{"A": "B"})
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
	res, err := invokeFunc()

	// assertions
	assert.Equal(t, expectedRequestCount, server.RequestCount, assertMsg)
	if shouldErr {
		if assert.Error(t, err, assertMsg) {
			assert.Contains(t, err.Error(), wantErr, assertMsg)
		}
		assert.Nil(t, res, assertMsg)
	} else {
		assert.Nil(t, err, assertMsg)
		if successFunc != nil {
			successFunc(res)
		}
	}
}
