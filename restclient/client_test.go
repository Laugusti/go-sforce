package restclient

import (
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/Laugusti/go-sforce/credentials"
	"github.com/Laugusti/go-sforce/internal/testserver"
	"github.com/Laugusti/go-sforce/session"
	"github.com/stretchr/testify/assert"
)

func TestBuildRequest(t *testing.T) {
	s := testserver.New(t)
	defer s.Stop()
	s.HandlerFunc = testserver.StaticJSONHandlerFunc(t, session.RequestToken{
		AccessToken: accessToken,
		InstanceURL: s.URL(),
	}, 200)

	client := &Client{
		sess:       session.Must(session.New(s.URL(), "version", credentials.New("user", "pass", "cid", "csecret"))),
		httpClient: s.Client(),
	}

	tests := []struct {
		apiPath  string
		rawQuery string
		body     string
	}{
		{"", "", ""},
		{"/path", "", ""},
		{"", "a=1,2b=3", "{}"},
		{"/api/path", "raw=query", "body"},
	}

	for _, test := range tests {
		req, err := client.buildRequest(test.apiPath, test.rawQuery, http.MethodGet, strings.NewReader(test.body))

		assert.Nil(t, err)
		assert.Contains(t, req.Header.Get("Authorization"), accessToken)
		assert.Equal(t, "application/json", req.Header.Get("Content-Type"))
		b, err := ioutil.ReadAll(req.Body)
		assert.Nil(t, err)
		assert.Equal(t, []byte(test.body), b)
		assert.Equal(t, test.apiPath, req.URL.Path)
		assert.Equal(t, test.rawQuery, req.URL.RawQuery)
	}
}
