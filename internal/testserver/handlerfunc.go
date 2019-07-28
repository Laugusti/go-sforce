package testserver

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

// ValidateAndSetResponseHandler runs each validator again the request, then sets the response body and status
func ValidateAndSetResponseHandler(t *testing.T, assertMessage string, body interface{}, statusCode int,
	validators ...RequestValidator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		for _, v := range validators {
			v.Validate(t, r, assertMessage)
		}
		// write response
		if body == nil {
			w.WriteHeader(statusCode)
		}
		b, err := json.Marshal(body)
		assert.Nil(t, err, assertMessage)
		w.WriteHeader(statusCode)
		_, err = w.Write(b)
		assert.Nil(t, err, assertMessage)
	}
}

// StaticJSONHandler creates reponse by marshalling the value as a json.
func StaticJSONHandler(t *testing.T, body interface{}, statusCode int) http.HandlerFunc {
	return ValidateAndSetResponseHandler(t, "", body, statusCode)
}

// ValidateJSONBodyHandler validates the request body, then sets the response body and status.
func ValidateJSONBodyHandler(t *testing.T, wantedRequestBody, respBody interface{},
	responseCode int, assertMessage string) http.HandlerFunc {
	return ValidateAndSetResponseHandler(t, "", respBody, responseCode, &JSONBodyValidator{wantedRequestBody})
}
