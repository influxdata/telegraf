package internal

import (
	"bufio"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"
	"unicode"
)

const alphanum string = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

// Duration just wraps time.Duration
type Duration struct {
	Duration time.Duration
}

// UnmarshalTOML parses the duration from the TOML config file
func (d *Duration) UnmarshalTOML(b []byte) error {
	dur, err := time.ParseDuration(string(b[1 : len(b)-1]))
	if err != nil {
		return err
	}

	d.Duration = dur

	return nil
}

var NotImplementedError = errors.New("not implemented yet")

// ReadLines reads contents from a file and splits them by new lines.
// A convenience wrapper to ReadLinesOffsetN(filename, 0, -1).
func ReadLines(filename string) ([]string, error) {
	return ReadLinesOffsetN(filename, 0, -1)
}

// ReadLines reads contents from file and splits them by new line.
// The offset tells at which line number to start.
// The count determines the number of lines to read (starting from offset):
//   n >= 0: at most n lines
//   n < 0: whole file
func ReadLinesOffsetN(filename string, offset uint, n int) ([]string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return []string{""}, err
	}
	defer f.Close()

	var ret []string

	r := bufio.NewReader(f)
	for i := 0; i < n+int(offset) || n < 0; i++ {
		line, err := r.ReadString('\n')
		if err != nil {
			break
		}
		if i < int(offset) {
			continue
		}
		ret = append(ret, strings.Trim(line, "\n"))
	}

	return ret, nil
}

// RandomString returns a random string of alpha-numeric characters
func RandomString(n int) string {
	var bytes = make([]byte, n)
	rand.Read(bytes)
	for i, b := range bytes {
		bytes[i] = alphanum[b%byte(len(alphanum))]
	}
	return string(bytes)
}

// GetTLSConfig gets a tls.Config object from the given certs, key, and CA files.
// you must give the full path to the files.
// If all files are blank and InsecureSkipVerify=false, returns a nil pointer.
func GetTLSConfig(
	SSLCert, SSLKey, SSLCA string,
	InsecureSkipVerify bool,
) (*tls.Config, error) {
	if SSLCert == "" && SSLKey == "" && SSLCA == "" && !InsecureSkipVerify {
		return nil, nil
	}

	t := &tls.Config{
		InsecureSkipVerify: InsecureSkipVerify,
	}

	if SSLCA != "" {
		caCert, err := ioutil.ReadFile(SSLCA)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("Could not load TLS CA: %s",
				err))
		}

		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)
		t.RootCAs = caCertPool
	}

	if SSLCert != "" && SSLKey != "" {
		cert, err := tls.LoadX509KeyPair(SSLCert, SSLKey)
		if err != nil {
			return nil, errors.New(fmt.Sprintf(
				"Could not load TLS client key/certificate: %s",
				err))
		}

		t.Certificates = []tls.Certificate{cert}
		t.BuildNameToCertificate()
	}

	// will be nil by default if nothing is provided
	return t, nil
}

// SnakeCase converts the given string to snake case following the Golang format:
// acronyms are converted to lower-case and preceded by an underscore.
func SnakeCase(in string) string {
	runes := []rune(in)
	length := len(runes)

	var out []rune
	for i := 0; i < length; i++ {
		if i > 0 && unicode.IsUpper(runes[i]) && ((i+1 < length && unicode.IsLower(runes[i+1])) || unicode.IsLower(runes[i-1])) {
			out = append(out, '_')
		}
		out = append(out, unicode.ToLower(runes[i]))
	}

	return string(out)
}

// Glob will test a string pattern, potentially containing globs, against a
// subject string. The result is a simple true/false, determining whether or
// not the glob pattern matched the subject text.
//
// Adapted from https://github.com/ryanuber/go-glob/blob/master/glob.go
// thanks Ryan Uber!
func Glob(pattern, measurement string) bool {
	// Empty pattern can only match empty subject
	if pattern == "" {
		return measurement == pattern
	}

	// If the pattern _is_ a glob, it matches everything
	if pattern == "*" {
		return true
	}

	parts := strings.Split(pattern, "*")

	if len(parts) == 1 {
		// No globs in pattern, so test for match
		return pattern == measurement
	}

	leadingGlob := strings.HasPrefix(pattern, "*")
	trailingGlob := strings.HasSuffix(pattern, "*")
	end := len(parts) - 1

	for i, part := range parts {
		switch i {
		case 0:
			if leadingGlob {
				continue
			}
			if !strings.HasPrefix(measurement, part) {
				return false
			}
		case end:
			if len(measurement) > 0 {
				return trailingGlob || strings.HasSuffix(measurement, part)
			}
		default:
			if !strings.Contains(measurement, part) {
				return false
			}
		}

		// Trim evaluated text from measurement as we loop over the pattern.
		idx := strings.Index(measurement, part) + len(part)
		measurement = measurement[idx:]
	}

	// All parts of the pattern matched
	return true
}
