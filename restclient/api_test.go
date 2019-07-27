package restclient

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Laugusti/go-sforce/credentials"
	"github.com/Laugusti/go-sforce/session"
)

func TestCreateSObject(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.Copy(w, strings.NewReader(`{}`))
	}))
	defer s.Close()
	client := &Client{
		sess:       session.Must(session.New(s.URL, "version", credentials.New("user", "pass", "cid", "csecret"))),
		httpClient: s.Client(),
	}
	_, err := client.CreateSObject("Item", map[string]interface{}{
		"Field1": "one",
		"Field2": 2,
	})
	if err != nil {
		t.Errorf("TestCreateSObject => create failed: %v", err)
	}
}
