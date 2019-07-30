# go-sforce
go-sforce is a simplified Saleforce client for the Go programming language.

## Installing
Use `go get` to retrieve the client to add to your `GOPATH` workspace:
```
go get github.com/Laugusti/go-sforce
```
Go Modules support coming soon.
## Reference Documentation
* [Package Reference](https://godoc.org/github.com/Laugusti/go-sforce/)
## Overview of Packages
* credentials - Provides OAuth credentials need to make API requests.
* session - Used to request an access token from the Salesfore API using the OAuth credentials.
* restclient - Client for the Salesforce Rest API.
## Usage
### Create a session
1. Create unauthenticated session
```
username, password := "user@example.com", "P@s$w0rD"
clientID, clientSecret := "test1n3289s32", `testjfa9a"afa8"'132%$#@@`
sess := session.Must(session.New("http://login.example.com", "42.0",
	credentials.New(username, password, clientID, clientSecret)))
```
2. Optionally request an access token before passing to client
```
err := sess.Login()
if err != nil {
	log.Fatal(err)
}
```
### Rest API client
1.  Create rest client from a session
```
restClient := restClient.New(sess)
```
2. Supported Methods
- CreateSObject - Used to creates a SObject in Salesforce using the object type.
- GetSObject - Used to retrieve a SObject using the object type and Salesforce id.
- GetSObjectByExternalID - Used to retrieve a SObject using the object type, external id field, and external id.
- UpsertSObject - Used to upsert (update/insert) a SObject using the object type and Salesforce id.
- UpsertSObjectByExternalID - Used to upsert a SObject using the object type, external id field, and external id.
- DeleteSObject - Used to delete a SObject using the object type and Salesforce id.
- Query - Used to execute a SOQL query in Salesforce.
- QueryMore - Used to get the remaining result of a SOQL query.
- FullQuery - Not supported by the Salesforce API. This exeuctes any and all necessary QueryMore for a query.

### Bulk API client
```
Coming soon...
```
