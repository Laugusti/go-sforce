package restclient

import (
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Laugusti/go-sforce/credentials"
	"github.com/Laugusti/go-sforce/session"
)

func TestBuildRequest(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(w, strings.NewReader(`{"instance_url":"url", "access_token": "token"}`))
	}))
	defer s.Close()
	client := &Client{
		sess:       session.Must(session.New(s.URL, "version", credentials.New("user", "pass", "cid", "csecret"))),
		httpClient: s.Client(),
	}
	req, err := client.buildRequest("apiPath", "raw=query", "GET", strings.NewReader("body"))
	if err != nil {
		t.Errorf("TestBuildRequest => %v", err)
	}

	if auth := req.Header.Get("Authorization"); auth != "Bearer token" {
		t.Errorf("TestBuildRequest => wrong auth: want %q, got %q", "Bearer token", auth)
	}
	if b, _ := ioutil.ReadAll(req.Body); string(b) != "body" {
		t.Errorf("TestBuildRequest => wrong body: want %q, got %q", "body", string(b))
	}
	if p := req.URL.Path; p != "url/apiPath" {
		t.Errorf("TestBuildRequest => wrong path: want %q, got %q", "url/apiPath", p)
	}
	if q := req.URL.RawQuery; q != "raw=query" {
		t.Errorf("TestBuildRequest => wrong query: want %q, got %q", "raw=query", q)
	}
}
