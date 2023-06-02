package processors

import (
	"io/ioutil"
	"net/http"
	"os"
)

const (
	RootEnvVar     = "SPM_ROOT"
	DefaultRootDir = "/opt/spm/"
)

// GetRootDir returns the root dir of Sematext Agent installation, if it is present. Otherwise empty string.
func GetRootDir() string {
	if dir := os.Getenv(RootEnvVar); dir != "" {
		if exists(dir) {
			return dir
		}
	}

	if exists(DefaultRootDir) {
		return DefaultRootDir
	}

	return ""
}

func exists(dir string) bool {
	if _, err := os.Stat(dir); err != nil {
		return false
	}
	return true
}

// response reads the response from the HTTP body.
func response(r *http.Response) string {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return ""
	}
	return string(body)
}
