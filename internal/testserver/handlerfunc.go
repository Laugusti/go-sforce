package testserver

import (
	"net/http"
	"testing"
)

// ValidateRequestHandlerFunc runs each validator again the request, then sets the response body and status
func ValidateRequestHandlerFunc(t *testing.T, assertMessage string, handler ResponseHandler,
	validators ...RequestValidator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		for _, v := range validators {
			v.Validate(t, r, assertMessage)
		}
		handler.Handle(t, w, assertMessage)
	}
}

// StaticJSONHandlerFunc creates reponse by marshalling the value as a json.
func StaticJSONHandlerFunc(t *testing.T, body interface{}, statusCode int) http.HandlerFunc {
	return ValidateRequestHandlerFunc(t, "", &JSONResponseHandler{statusCode, body})
}
