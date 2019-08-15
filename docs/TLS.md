# TLS settings

TLS for output plugin will be used if you provide options `tls_cert` and `tls_key`.
Settings that can be used to configure TLS:

- `tls_cert` - path to certificate. Type: `string`. Ex. `tls_cert = "/etc/ssl/telegraf.crt"`
- `tls_key` - path to key. Type: `string`, Ex. `tls_key = "/etc/ssl/telegraf.key"`
- `tls_allowed_cacerts` - Set one or more allowed client CA certificate file names to enable mutually authenticated TLS connections. Type: `list`. Ex. `tls_allowed_cacerts = ["/etc/telegraf/clientca.pem"]`
- `tls_cipher_suites`- Define list of ciphers that will be supported. If wasn't defined default will be used. Type: `list`. Ex. `tls_cipher_suites = ["TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305", "TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305", "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256", "TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256", "TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384", "TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384", "TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256", "TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA", "TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256", "TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA", "TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA", "TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA", "TLS_RSA_WITH_AES_128_GCM_SHA256", "TLS_RSA_WITH_AES_256_GCM_SHA384", "TLS_RSA_WITH_AES_128_CBC_SHA256", "TLS_RSA_WITH_AES_128_CBC_SHA", "TLS_RSA_WITH_AES_256_CBC_SHA"]`
- `tls_min_version` - Minimum TLS version that is acceptable. If wasn't defined default (TLS 1.0) will be used. Type: `string`. Ex. `tls_min_version = "TLS11"`
- `tls_max_version` - Maximum SSL/TLS version that is acceptable. If not set, then the maximum version supported is used, which is currently TLS 1.2 (for go < 1.12) or TLS 1.3 (for go == 1.12). Ex. `tls_max_version = "TLS12"`

tls ciphers are supported:
- TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305
- TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305
- TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256
- TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256
- TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384
- TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384
- TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256
- TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA
- TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256
- TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA
- TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA
- TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA
- TLS_RSA_WITH_AES_128_GCM_SHA256
- TLS_RSA_WITH_AES_256_GCM_SHA384
- TLS_RSA_WITH_AES_128_CBC_SHA256
- TLS_RSA_WITH_AES_128_CBC_SHA
- TLS_RSA_WITH_AES_256_CBC_SHA
- TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA
- TLS_RSA_WITH_3DES_EDE_CBC_SHA
- TLS_RSA_WITH_RC4_128_SHA
- TLS_ECDHE_RSA_WITH_RC4_128_SHA
- TLS_ECDHE_ECDSA_WITH_RC4_128_SHA
- TLS_AES_128_GCM_SHA256 (only if version go1.12 was used for make build)
- TLS_AES_256_GCM_SHA384 (only if version go1.12 was used for make build)
- TLS_CHACHA20_POLY1305_SHA256 (only if version go1.12 was used for make build)

TLS versions are supported:
- TLS10
- TLS11
- TLS12
- TLS13 (only if version go1.12 was used for make build)
