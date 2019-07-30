package testserver

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

// ValidateRequestHandlerFunc runs each validator again the request, then sets the response body and status
func ValidateRequestHandlerFunc(t *testing.T, assertMessage string, handler ResponseHandler,
	validators ...RequestValidator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		for _, v := range validators {
			assert.Nil(t, v.Validate(r), assertMessage)
		}
		assert.Nil(t, handler.Handle(w), assertMessage)
	}
}

// StaticJSONHandlerFunc creates reponse by marshalling the value as a json.
func StaticJSONHandlerFunc(t *testing.T, statusCode int, body interface{}) http.HandlerFunc {
	return ValidateRequestHandlerFunc(t, "", &JSONResponseHandler{statusCode, body})
}
