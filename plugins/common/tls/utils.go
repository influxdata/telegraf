package tls

import (
	"fmt"
)

// ParseCiphers returns a `[]uint16` by received `[]string` key that represents ciphers from crypto/tls.
// If some of ciphers in received list doesn't exists  ParseCiphers returns nil with error
func ParseCiphers(ciphers []string) ([]uint16, error) {
	suites := []uint16{}

	for _, cipher := range ciphers {
		v, ok := tlsCipherMap[cipher]
		if !ok {
			return nil, fmt.Errorf("unsupported cipher %q", cipher)
		}
		suites = append(suites, v)
	}

	return suites, nil
}

// ParseTLSVersion returns a `uint16` by received version string key that represents tls version from crypto/tls.
// If version isn't supported ParseTLSVersion returns 0 with error
func ParseTLSVersion(version string) (uint16, error) {
	if v, ok := tlsVersionMap[version]; ok {
		return v, nil
	}
	return 0, fmt.Errorf("unsupported version %q", version)
}
