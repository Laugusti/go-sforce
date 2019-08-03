package session

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"

	"github.com/Laugusti/go-sforce/credentials"
)

func ExampleSession_Login() {
	// test login server that responds with access token.
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(RequestToken{
			AccessToken: "access_token",
			InstanceURL: "instance_url",
			ID:          "id",
			TokenType:   "token_type",
			IssuedAt:    "issued_at",
			Signature:   "signature",
		})
	}))
	defer s.Close()

	// create unauthenticated session
	username, password := "user@example.com", "P@s$w0rD"
	clientID, clientSecret := "test1n3289s32", `testjfa9a"afa8"'132%$#@@`
	sess := Must(New(s.URL, "42.0", credentials.New(username, password, clientID, clientSecret)))
	fmt.Println(sess.HasToken())

	// try login
	err := sess.Login()
	if err != nil {
		log.Fatalf("login failed: %v", err)
	}

	fmt.Println(sess.HasToken())
	fmt.Println(sess.AccessToken())
	fmt.Println(sess.InstanceURL())

	// Output:
	// false
	// true
	// access_token
	// instance_url
}
