package testserver

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

// StaticJSONHandler creates reponse by marshalling the value as a json.
func StaticJSONHandler(v interface{}, statusCode int) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		b, err := json.Marshal(v)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(statusCode)
		_, _ = w.Write(b)
	}
}

// ValidateJSONBodyHandler validates the request body, then sets the response body and status.
func ValidateJSONBodyHandler(t *testing.T, wantedRequestBody, respBody interface{},
	responseCode int, assertMessage string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if wantedRequestBody != nil {
			// get expected body as map
			want, err := jsonObjectToMap(wantedRequestBody)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			// get request body as map
			got := make(map[string]interface{})
			err = json.NewDecoder(r.Body).Decode(&got)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			// assert body matches expected
			assert.Equal(t, want, got, assertMessage)
		}
		// write response
		b, err := json.Marshal(respBody)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(responseCode)
		_, _ = w.Write(b)
	}
}

func jsonObjectToMap(v interface{}) (map[string]interface{}, error) {
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(v)
	if err != nil {
		return nil, err
	}
	m := make(map[string]interface{})
	err = json.NewDecoder(&buf).Decode(&m)
	return m, err
}
