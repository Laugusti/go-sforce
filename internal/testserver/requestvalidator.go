package testserver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"strings"
)

// RequestValidator is an interface with a single method to validate a http request.
// Validators should return a descriptive error if the validation fails.
type RequestValidator interface {
	Validate(*http.Request) error
}

func deepCompare(name string, want, got interface{}) error {
	if !reflect.DeepEqual(want, got) {
		return fmt.Errorf("%s failed: want %q, got %q", name, want, got)
	}
	return nil
}

// HeaderValidator validates headers on the request.
type HeaderValidator struct {
	Key   string
	Value string
}

// Validate implements the RequestValidator interface.
func (v *HeaderValidator) Validate(r *http.Request) error {
	return deepCompare("HeaderValidator", v.Value, r.Header.Get(v.Key))
}

// JSONBodyValidator validates the request body as JSON.
type JSONBodyValidator struct {
	Body interface{}
}

// Validate implements the RequestValidator interface.
func (v *JSONBodyValidator) Validate(r *http.Request) error {
	if v.Body == nil {
		// this case should not happen when validating server requests
		if r.Body == nil {
			return nil
		}
		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return fmt.Errorf("JSONBodyValidator failed: could not read request body: %v", err)
		}

		if len(b) != 0 {
			return fmt.Errorf("JSONBodyValidator failed: want empty body, got %q", string(b))
		}
		return nil
	}
	//get expected body as map
	want, err := jsonObjectToMap(v.Body)
	if err != nil {
		return fmt.Errorf("JSONBodyValidator failed: could not convert object to map: %v", err)
	}
	// get request body as map
	got := make(map[string]interface{})
	if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
		return fmt.Errorf("JSONBodyValidator failed: could not unmarshal request body: %v", err)
	}
	// assert body matches
	return deepCompare("JSONBodyValidator", want, got)
}

// PathValidator validates the request path.
type PathValidator struct {
	Path string
}

// Validate implements the RequestValidator interface.
func (v *PathValidator) Validate(r *http.Request) error {
	path := r.URL.Path
	if !strings.HasPrefix(v.Path, "/") {
		path = strings.TrimPrefix(r.URL.Path, "/")
	}
	return deepCompare("PathValidator", v.Path, path)
}

// QueryValidator validates the request query.
type QueryValidator struct {
	Query url.Values
}

// Validate implements the RequestValidator interface.
func (v *QueryValidator) Validate(r *http.Request) error {
	return deepCompare("QueryValidator", v.Query, r.URL.Query())
}

// MethodValidator validates the request method.
type MethodValidator struct {
	Method string
}

// Validate implements the RequestValidator interface.
func (v *MethodValidator) Validate(r *http.Request) error {
	return deepCompare("MethodValidator", v.Method, r.Method)
}

// FormValidator validates the request form.
type FormValidator struct {
	Form url.Values
}

// Validate implements the RequestValidator interface.
func (v *FormValidator) Validate(r *http.Request) error {
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("FormValidator failed: could not parse form: %v", err)
	}
	return deepCompare("FormValidator", v.Form, r.Form)
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
