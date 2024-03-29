package request

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/Laugusti/go-sforce/internal/testserver"
	"github.com/Laugusti/go-sforce/sforce/credentials"
	"github.com/Laugusti/go-sforce/sforce/session"
	"github.com/stretchr/testify/assert"
)

func TestDoWithResult(t *testing.T) {
	s := testserver.New(t)
	defer s.Stop()

	// create session
	sess := session.Must(session.New(
		s.URL(),
		"version",
		credentials.New("user", "pass", "cid", "csecret"),
	))
	sess.HTTPClient = s.Client()
	// login
	s.HandlerFunc = testserver.StaticJSONHandlerFunc(t, http.StatusOK,
		session.RequestToken{
			InstanceURL: s.URL(),
		})
	assert.Nil(t, sess.Login())

	// use response variable as response
	var response string
	s.HandlerFunc = func(w http.ResponseWriter, h *http.Request) {
		_, _ = w.Write([]byte(response))
	}
	tests := []struct {
		shouldErr bool
		response  string
		respType  ResultType
	}{
		{true, ``, JSONResult},
		{true, ``, XMLResult},
		{false, `{"key":"value"}`, JSONResult},
		{false, `<root><key>value</key></root>`, XMLResult},
		{true, `{"key":"value"}`, XMLResult},
		{true, `<root><key>value</key></root>`, JSONResult},
	}

	for _, test := range tests {
		assertMsg := fmt.Sprintf("input: %v", test)
		// set response
		response = test.response

		var got interface{}
		req := New(sess, &Operation{Method: "POST"},
			NewResultExpectation(test.respType, http.StatusOK),
			&got,
		)

		err := req.Send()
		if test.shouldErr {
			assert.NotNil(t, err, assertMsg)
		} else {
			assert.Nil(t, err, assertMsg)
		}
	}
}
