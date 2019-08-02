package restclient

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"

	"github.com/Laugusti/go-sforce/credentials"
	"github.com/Laugusti/go-sforce/session"
)

func ExampleClient_CreateSObject() {
	loginURL, tearDown := testSalesforceServer(http.StatusCreated, UpsertResult{
		ID:      "00xTEST00123",
		Success: true,
		Errors:  []interface{}{},
	})
	defer tearDown()

	// create unauthenticated session
	username, password := "user@example.com", "P@s$w0rD"
	clientID, clientSecret := "test1n3289s32", `testjfa9a"afa8"'132%$#@@`
	sess := session.Must(session.New(loginURL, "42.0",
		credentials.New(username, password, clientID, clientSecret)))

	// create salesforce client
	client := New(sess)

	sobj := NewSObjectBuilder().
		NewField("Name", "test").
		NewField("Field1", "Value").
		NewDottedField("Parent.ExternalId", "parentId").
		MustBuild()

	res, err := client.CreateSObject("Object", sobj)
	if err != nil {
		log.Fatalf("create failed: %v", err)
	}

	fmt.Println(res)

	// Output:
	// &{00xTEST00123 true []}
}

func ExampleClient_GetSObject() {
	loginURL, tearDown := testSalesforceServer(http.StatusOK, NewSObjectBuilder().
		NewField("Id", "salesforceId").
		NewField("Name", "test").
		NewField("Field1", "Value").
		MustBuild(),
	)
	defer tearDown()

	// create unauthenticated session
	username, password := "user@example.com", "P@s$w0rD"
	clientID, clientSecret := "test1n3289s32", `testjfa9a"afa8"'132%$#@@`
	sess := session.Must(session.New(loginURL, "42.0",
		credentials.New(username, password, clientID, clientSecret)))

	// create salesforce client
	client := New(sess)

	sobj, err := client.GetSObject("Object", "salesforceId")
	if err != nil {
		log.Fatalf("get failed: %v", err)
	}

	fmt.Println(sobj)

	// Output:
	// map[Field1:Value Id:salesforceId Name:test]
}

func ExampleClient_FullQuery() {
	loginURL, tearDown := testSalesforceServer(http.StatusOK, QueryResult{
		TotalSize:      3,
		Done:           true,
		NextRecordsURL: "",
		Records: []SObject{
			NewSObjectBuilder().NewField("Id", "id1").MustBuild(),
			NewSObjectBuilder().NewField("Id", "id2").MustBuild(),
			NewSObjectBuilder().NewField("Id", "id3").MustBuild(),
		},
	},
	)
	defer tearDown()

	// create unauthenticated session
	username, password := "user@example.com", "P@s$w0rD"
	clientID, clientSecret := "test1n3289s32", `testjfa9a"afa8"'132%$#@@`
	sess := session.Must(session.New(loginURL, "42.0",
		credentials.New(username, password, clientID, clientSecret)))

	// create salesforce client
	client := New(sess)

	res, err := client.FullQuery("select Id from Object")
	if err != nil {
		log.Fatalf("query failed: %v", err)
	}

	fmt.Println(res)

	// Output:
	// &{3 true  [map[Id:id1] map[Id:id2] map[Id:id3]]}
}

func testSalesforceServer(statusCode int, body interface{}) (string, func()) {
	instanceServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(statusCode)
		_ = json.NewEncoder(w).Encode(body)
	}))
	loginServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(session.RequestToken{
			InstanceURL: instanceServer.URL,
		})

	}))
	return loginServer.URL, func() { instanceServer.Close(); loginServer.Close() }
}
