# HTTP Secret-store Plugin

The `http` plugin allows to query secrets from an HTTP endpoint. The secrets
can be transmitted plain-text or in an encrypted fashion.

To manage your secrets of this secret-store, you should use Telegraf. Run

```shell
telegraf secrets help
```

to get more information on how to do this.

## Usage <!-- @/docs/includes/secret_usage.md -->

Secrets defined by a store are referenced with `@{<store-id>:<secret_key>}`
the Telegraf configuration. Only certain Telegraf plugins and options of
support secret stores. To see which plugins and options support
secrets, see their respective documentation (e.g.
`plugins/outputs/influxdb/README.md`). If the plugin's README has the
`Secret-store support` section, it will detail which options support secret
store usage.

## Configuration

```toml @sample.conf
# Read secrets from a HTTP endpoint
[[secretstores.http]]
  ## Unique identifier for the secret-store.
  ## This id can later be used in plugins to reference the secrets
  ## in this secret-store via @{<id>:<secret_key>} (mandatory)
  id = "secretstore"

  ## URLs from which to read the secrets
  url = "http://localhost/secrets"

  ## Optional HTTP headers
  # headers = {"X-Special-Header" = "Special-Value"}

  ## Optional Token for Bearer Authentication via
  ## "Authorization: Bearer <token>" header
  # token = "your-token"

  ## Optional Credentials for HTTP Basic Authentication
  # username = "username"
  # password = "pa$$word"

  ## OAuth2 Client Credentials. The options 'client_id', 'client_secret', and 'token_url' are required to use OAuth2.
  # client_id = "clientid"
  # client_secret = "secret"
  # token_url = "https://indentityprovider/oauth2/v1/token"
  # scopes = ["urn:opc:idm:__myscopes__"]

  ## HTTP Proxy support
  # use_system_proxy = false
  # http_proxy_url = ""

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Minimal TLS version to accept by the client
  # tls_min_version = "TLS12"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false

  ## Optional Cookie authentication
  # cookie_auth_url = "https://localhost/authMe"
  # cookie_auth_method = "POST"
  # cookie_auth_username = "username"
  # cookie_auth_password = "pa$$word"
  # cookie_auth_headers = { Content-Type = "application/json", X-MY-HEADER = "hello" }
  # cookie_auth_body = '{"username": "user", "password": "pa$$word", "authenticate": "me"}'
  ## When unset or set to zero the authentication will only happen once
  ## and will never renew the cookie. Set to a suitable duration if you
  ## require cookie renewal!
  # cookie_auth_renewal = "0s"

  ## Amount of time allowed to complete the HTTP request
  # timeout = "5s"

  ## List of success status codes
  # success_status_codes = [200]

  ## JSONata expression to transform the server response into a
  ##   { "secret name": "secret value", ... }
  ## form. See https://jsonata.org for more information and a playground.
  # transformation = ''

  ## Cipher used to decrypt the secrets.
  ## In case your secrets are transmitted in an encrypted form, you need
  ## to specify the cipher used and provide the corresponding configuration.
  ## Please refer to https://github.com/influxdata/telegraf/blob/master/plugins/secretstores/http/README.md
  ## for supported values.
  # cipher = "none"

  ## AES cipher parameters
  # [secretstores.http.aes]
  #   ## Key (hex-encoded) and initialization-vector (IV) for the decryption.
  #   ## In case the key (and IV) is derived from a password, the values can
  #   ## be omitted.
  #   key = ""
  #   init_vector = ""
  #
  #   ## Parameters for password-based-key derivation.
  #   ## These parameters must match the encryption side to derive the same
  #   ## key on both sides!
  #   # kdf_algorithm = "PBKDF2-HMAC-SHA256"
  #   # password = ""
  #   # salt = ""
  #   # iterations = 0
```

A collection of secrets is queried from the `url` endpoint. The plugin currently
expects JSON data in a flat key-value form and means to convert arbitrary JSON
to that form (see [transformation section](#transformation)).
Furthermore, the secret data can be transmitted in an encrypted
format, see [encryption section](#encryption) for details.

## Transformation

Secrets are currently expected to be JSON data in the following flat key-value
form

```json
{
    "secret name A": "secret value A",
    ...
    "secret name X": "secret value X"
}
```

If your HTTP endpoint provides JSON data in a different format, you can use
the `transformation` option to apply a [JSONata expression](https://jsonata.org)
(version v1.5.4) to transform the server answer to the above format.

## Encryption

### Plain text

Set `cipher` to `none` if the secrets are transmitted as plain-text. No further
options are required.

### Advanced Encryption Standard (AES)

Currently the following AES ciphers are supported

- `AES128/CBC`: 128-bit key in _CBC_ block mode without padding
- `AES128/CBC/PKCS#5`: 128-bit key in _CBC_ block mode with _PKCS#5_ padding
- `AES128/CBC/PKCS#7`: 128-bit key in _CBC_ block mode with _PKCS#7_ padding
- `AES192/CBC`: 192-bit key in _CBC_ block mode without padding
- `AES192/CBC/PKCS#5`: 192-bit key in _CBC_ block mode with _PKCS#5_ padding
- `AES192/CBC/PKCS#7`: 192-bit key in _CBC_ block mode with _PKCS#7_ padding
- `AES256/CBC`: 256-bit key in _CBC_ block mode without padding
- `AES256/CBC/PKCS#5`: 256-bit key in _CBC_ block mode with _PKCS#5_ padding
- `AES256/CBC/PKCS#7`: 256-bit key in _CBC_ block mode with _PKCS#7_ padding

Additional to the cipher, you need to provide the encryption `key` and
initialization vector `init_vector` to be able to decrypt the data.
In case you are using password-based key derivation, `key`
(and possibly `init_vector`) can be omitted. Take a look at the
[password-based key derivation section](#password-based-key-derivation).

### Password-based key derivation

Alternatively to providing a `key` (and `init_vector`) the key (and vector)
can be derived from a given password. Currently the following algorithms are
supported for `kdf_algorithm`:

- `PBKDF2-HMAC-SHA256` for `key` only, no `init_vector` created

You also need to provide the `password` to derive the key from as well as the
`salt` and `iterations` used.
__Please note:__ All parameters must match the encryption side to derive the
same key in Telegraf!
