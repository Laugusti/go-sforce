package testserver

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
)

func jsonObjectToReadCloser(v interface{}) (io.ReadCloser, error) {
	if v == nil {
		return nil, nil
	}
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(v)
	if err != nil {
		return nil, err
	}
	return ioutil.NopCloser(&buf), nil
}
func jsonObjectToMap(v interface{}) (map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}
	r, err := jsonObjectToReadCloser(v)
	if err != nil {
		return nil, err
	}
	m := make(map[string]interface{})
	if err := json.NewDecoder(r).Decode(&m); err != nil {
		return nil, err
	}
	return m, nil
}

func jsonReaderToMap(r io.Reader) (map[string]interface{}, error) {
	m := make(map[string]interface{})
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	if len(b) == 0 {
		return nil, nil
	}
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, err
	}
	return m, err
}
