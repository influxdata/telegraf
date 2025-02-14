package tls

import (
	"crypto/tls"
	"sync"
)

var tlsVersionMap = map[string]uint16{
	"TLS10": tls.VersionTLS10,
	"TLS11": tls.VersionTLS11,
	"TLS12": tls.VersionTLS12,
	"TLS13": tls.VersionTLS13,
}

var tlsCipherMapInit sync.Once
var tlsCipherMapSecure map[string]uint16
var tlsCipherMapInsecure map[string]uint16

func init() {
	tlsCipherMapInit.Do(func() {
		// Initialize the secure suites
		suites := tls.CipherSuites()
		tlsCipherMapSecure = make(map[string]uint16, len(suites))
		for _, s := range suites {
			tlsCipherMapSecure[s.Name] = s.ID
		}

		suites = tls.InsecureCipherSuites()
		tlsCipherMapInsecure = make(map[string]uint16, len(suites))
		for _, s := range suites {
			tlsCipherMapInsecure[s.Name] = s.ID
		}
	})
}
