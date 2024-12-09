# Javascript Object Signing and Encryption Secret-store Plugin

The `jose` plugin allows to manage and store secrets locally
protected by the [Javascript Object Signing and Encryption][jose] algorithm.

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
# File based Javascript Object Signing and Encryption based secret-store
[[secretstores.jose]]
  ## Unique identifier for the secret-store.
  ## This id can later be used in plugins to reference the secrets
  ## in this secret-store via @{<id>:<secret_key>} (mandatory)
  id = "secretstore"

  ## Directory for storing the secrets
  path = "/etc/telegraf/secrets"

  ## Password to access the secrets.
  ## If no password is specified here, Telegraf will prompt for it at startup time.
  # password = ""
```

Each secret is stored in an individual file in the subdirectory specified
using the `path` parameter. To access the secrets, a password is required.
This password can be specified using the `password` parameter containing a
string, an environment variable or as a reference to a secret in another
secret store. If `password` is not specified in the config, you will be
prompted for the password at startup.

__Please note:__ All secrets in this secret store are encrypted using
the same password. If you need individual passwords for each `jose`
secret, please use multiple instances of this plugin.

[jose]: https://github.com/dvsekhvalnov/jose2go
