package testserver

import (
	"net/http"
	"testing"
)

// ValidateAndSetResponseHandler runs each validator again the request, then sets the response body and status
func ValidateAndSetResponseHandler(t *testing.T, assertMessage string, handler ResponseHandler,
	validators ...RequestValidator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		for _, v := range validators {
			v.Validate(t, r, assertMessage)
		}
		handler.Handle(t, w, assertMessage)
	}
}

// StaticJSONHandler creates reponse by marshalling the value as a json.
func StaticJSONHandler(t *testing.T, body interface{}, statusCode int) http.HandlerFunc {
	return ValidateAndSetResponseHandler(t, "", &JSONBodyResponseHandler{statusCode, body})
}
