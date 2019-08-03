package request

import (
	"net/url"
	"path"
)

func joinURL(baseURL string, paths ...string) (string, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}
	u.Path = path.Join(u.Path, path.Join(paths...))
	return u.String(), nil
}

// isInSlice returns true if the value is in the slice.
func isInSlice(value int, slice []int) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}
