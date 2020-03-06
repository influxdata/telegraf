package common

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// ParseConnectionString parses the given connection string into a key-value map,
// returns an error if at least one of required keys is missing.
func ParseConnectionString(cs string, require ...string) (map[string]string, error) {
	m := map[string]string{}
	for _, s := range strings.Split(cs, ";") {
		if s == "" {
			continue
		}
		kv := strings.SplitN(s, "=", 2)
		if len(kv) != 2 {
			return nil, errors.New("malformed connection string")
		}
		m[kv[0]] = kv[1]
	}
	for _, k := range require {
		if s := m[k]; s == "" {
			return nil, fmt.Errorf("%s is required", k)
		}
	}
	return m, nil
}

// NewSharedAccessKey creates new shared access key for subsequent token generation.
func NewSharedAccessKey(hostname, policy, key string) *SharedAccessKey {
	return &SharedAccessKey{
		HostName:            hostname,
		SharedAccessKeyName: policy,
		SharedAccessKey:     key,
	}
}

// SharedAccessKey is SAS token generator.
type SharedAccessKey struct {
	HostName            string
	SharedAccessKeyName string
	SharedAccessKey     string
}

// Token generates a shared access signature for the named resource and lifetime.
func (c *SharedAccessKey) Token(
	resource string, lifetime time.Duration,
) (*SharedAccessSignature, error) {
	return NewSharedAccessSignature(
		resource, c.SharedAccessKeyName, c.SharedAccessKey, time.Now().Add(lifetime),
	)
}

// NewSharedAccessSignature initialized a new shared access signature
// and generates signature fields based on the given input.
func NewSharedAccessSignature(
	resource, policy, key string, expiry time.Time,
) (*SharedAccessSignature, error) {
	sig, err := mksig(resource, key, expiry)
	if err != nil {
		return nil, err
	}
	return &SharedAccessSignature{
		Sr:  resource,
		Sig: sig,
		Se:  expiry,
		Skn: policy,
	}, nil
}

func mksig(sr, key string, se time.Time) (string, error) {
	b, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		return "", err
	}
	h := hmac.New(sha256.New, b)
	if _, err := fmt.Fprintf(h, "%s\n%d", url.QueryEscape(sr), se.Unix()); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(h.Sum(nil)), nil
}

// SharedAccessSignature is a shared access signature instance.
type SharedAccessSignature struct {
	Sr  string
	Sig string
	Se  time.Time
	Skn string
}

// String converts the signature to a token string.
func (sas *SharedAccessSignature) String() string {
	s := "SharedAccessSignature " +
		"sr=" + url.QueryEscape(sas.Sr) +
		"&sig=" + url.QueryEscape(sas.Sig) +
		"&se=" + url.QueryEscape(strconv.FormatInt(sas.Se.Unix(), 10))
	if sas.Skn != "" {
		s += "&skn=" + url.QueryEscape(sas.Skn)
	}
	return s
}
