// +build go1.12

package tls

import "crypto/tls"

func init() {
	tlsVersionMap["TLS13"] = tls.VersionTLS13
	tlsCipherMap["TLS_AES_128_GCM_SHA256"] = tls.TLS_AES_128_GCM_SHA256
	tlsCipherMap["TLS_AES_256_GCM_SHA384"] = tls.TLS_AES_256_GCM_SHA384
	tlsCipherMap["TLS_CHACHA20_POLY1305_SHA256"] = tls.TLS_CHACHA20_POLY1305_SHA256
}
