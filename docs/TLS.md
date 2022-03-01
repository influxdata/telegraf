# Transport Layer Security

There is an ongoing effort to standardize TLS options across plugins.  When
possible, plugins will provide the standard settings described below.  With the
exception of the advanced configuration available TLS settings will be
documented in the sample configuration.

## Client Configuration

For client TLS support we have the following options:

```toml
## Root certificates for verifying server certificates encoded in PEM format.
# tls_ca = "/etc/telegraf/ca.pem"

## The public and private keypairs for the client encoded in PEM format.  May
## contain intermediate certificates.
# tls_cert = "/etc/telegraf/cert.pem"
# tls_key = "/etc/telegraf/key.pem"
## Skip TLS verification.
# insecure_skip_verify = false
## Send the specified TLS server name via SNI.
# tls_server_name = "foo.example.com"
```

### Server Configuration

The server TLS configuration provides support for TLS mutual authentication:

```toml
## Set one or more allowed client CA certificate file names to
## enable mutually authenticated TLS connections.
# tls_allowed_cacerts = ["/etc/telegraf/clientca.pem"]

## Set one or more allowed DNS name to enable a whitelist
## to verify incoming client certificates.
## It will go through all available SAN in the certificate,
## if of them matches the request is accepted.
# tls_allowed_dns_names = ["client.example.org"]

## Add service certificate and key.
# tls_cert = "/etc/telegraf/cert.pem"
# tls_key = "/etc/telegraf/key.pem"
```

#### Advanced Configuration

For plugins using the standard server configuration you can also set several
advanced settings.  These options are not included in the sample configuration
for the interest of brevity.

```toml
## Define list of allowed ciphers suites.  If not defined the default ciphers
## supported by Go will be used.
##   ex: tls_cipher_suites = [
##           "TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305",
##           "TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305",
##           "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
##           "TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256",
##           "TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384",
##           "TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384",
##           "TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256",
##           "TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA",
##           "TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256",
##           "TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA",
##           "TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA",
##           "TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA",
##           "TLS_RSA_WITH_AES_128_GCM_SHA256",
##           "TLS_RSA_WITH_AES_256_GCM_SHA384",
##           "TLS_RSA_WITH_AES_128_CBC_SHA256",
##           "TLS_RSA_WITH_AES_128_CBC_SHA",
##           "TLS_RSA_WITH_AES_256_CBC_SHA"
##       ]
# tls_cipher_suites = []

## Minimum TLS version that is acceptable.
# tls_min_version = "TLS10"

## Maximum SSL/TLS version that is acceptable.
# tls_max_version = "TLS13"
```

Cipher suites for use with `tls_cipher_suites`:

- `TLS_RSA_WITH_RC4_128_SHA`
- `TLS_RSA_WITH_3DES_EDE_CBC_SHA`
- `TLS_RSA_WITH_AES_128_CBC_SHA`
- `TLS_RSA_WITH_AES_256_CBC_SHA`
- `TLS_RSA_WITH_AES_128_CBC_SHA256`
- `TLS_RSA_WITH_AES_128_GCM_SHA256`
- `TLS_RSA_WITH_AES_256_GCM_SHA384`
- `TLS_ECDHE_ECDSA_WITH_RC4_128_SHA`
- `TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA`
- `TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA`
- `TLS_ECDHE_RSA_WITH_RC4_128_SHA`
- `TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA`
- `TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA`
- `TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA`
- `TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256`
- `TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256`
- `TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256`
- `TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256`
- `TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384`
- `TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384`
- `TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305`
- `TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305`
- `TLS_AES_128_GCM_SHA256`
- `TLS_AES_256_GCM_SHA384`
- `TLS_CHACHA20_POLY1305_SHA256`

TLS versions for use with `tls_min_version` or `tls_max_version`:

- `TLS10`
- `TLS11`
- `TLS12`
- `TLS13`
