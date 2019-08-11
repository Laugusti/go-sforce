package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/Laugusti/go-sforce/sforce/session"
	"github.com/Laugusti/go-sforce/sforce/sforceerr"
)

func splitString(s, sep string) []string {
	if s == "" {
		return []string{}
	}
	s = strings.ReplaceAll(s, " ", "")
	return strings.Split(s, sep)
}

func marshalJSONToStdout(cmd string, v interface{}) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "\t")
	exitIfError(cmd, enc.Encode(v))

}

func unmarshalJSONFile(cmd, file string, v interface{}) {
	unmarshalFile(cmd, file, json.Unmarshal, v)
}

func unmarshalFile(cmd, file string, unmarshalFunc func([]byte, interface{}) error,
	v interface{}) {
	var data []byte
	if file != "-" {
		// get data from file
		b, err := ioutil.ReadFile(file)
		exitIfError(cmd, err)
		data = b
	} else {
		// get data from stdin
		b, err := ioutil.ReadAll(os.Stdin)
		exitIfError(cmd, err)
		data = b
	}
	// unmarshal data
	exitIfError(cmd, unmarshalFunc(data, v))
}

func exitIfError(operation string, err error) {
	if err == nil {
		return
	}
	switch err := err.(type) {
	case *session.LoginError:
		fmt.Fprintf(os.Stderr, "Login failed (%s): %s\n", err.ErrorCode, err.Message)
	case *sforceerr.APIError:
		fmt.Fprintf(os.Stderr, "An error occurred (%s) when calling the %s operation: %s\n",
			err.ErrorCode, operation, err.Message)
	default:
		fmt.Fprintln(os.Stderr, err)
	}
	os.Exit(1)
}
