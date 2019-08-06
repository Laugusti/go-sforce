package cmd

import (
	"fmt"
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
