# HashiCorp Vault Secret-Store Plugin

The `vault` plugin allows to utilize secrets stored in a
[HashiCorp Vault][vault] server via the Vault API. It supports authentication
via a pre-obtained Vault token or via the AppRole method.

⭐ Telegraf v1.37.0
🏷️ secrets
💻 all

## Usage <!-- @/docs/includes/secret_usage.md -->

Secrets defined by a store are referenced with `@{<store-id>:<secret_key>}`
the Telegraf configuration. Only certain Telegraf plugins and options of
support secret stores. To see which plugins and options support
secrets, see their respective documentation (e.g.
`plugins/outputs/influxdb/README.md`). If the plugin's README has the
`Secret-store support` section, it will detail which options support secret
store usage.

## Configuration

### Authentication

When authenticating with a `token`, the token may be provided directly or
chained from another secret-store (e.g. `@{other_store:vault_token}`). This
lets you obtain a token through any mechanism another secret-store can
produce (OAuth2, file, environment, etc.) and hand it to this plugin. Token
renewal is the responsibility of the supplying source.

When authenticating with `approle`, the plugin logs in with the configured
Role ID and Secret ID and starts a lifetime watcher to keep the token
renewed.

```toml @sample.conf
# Secret-store to access Vault Secrets
[[secretstores.vault]]
  ## Unique identifier for the secretstore.
  ## This id can later be used in plugins to reference the secrets
  ## in this secret-store via @{<id>:<secret_key>} (mandatory)
  id = "vault_secretstore"

  ## Address of the Vault server
  address = "localhost:8200"

  ## Mount path of the KV secrets engine.
  ## This is the path where the KV secrets engine is enabled. For example, if
  ## your full secret path in the Vault CLI is "secret/data/myapp/database",
  ## then mount_path = "secret".
  mount_path = ""

  ## Path to the secret within the KV secrets engine.
  ## This is the path to your specific secret under the mount point. For example,
  ## if your full secret path is "secret/data/myapp/database", then
  ## secret_path = "myapp/database". Note that the "/data/" segment in KV v2
  ## paths is handled automatically and should not be included.
  secret_path = ""

  ## Secret store engine to use.
  ## Supports 'kv-v1' and 'kv-v2' engines.
  ## By default will use the kv-v2 engine.
  # engine = "kv-v2"

  ## Authentication
  ## Exactly one of "token" or "approle" must be configured. Use "token" to
  ## pass an already-obtained Vault token (directly or via another
  ## secret-store, e.g. @{other_store:vault_token}). Use "approle" to have
  ## Telegraf authenticate via the AppRole method and manage token renewal.

  ## Vault token used to authenticate with the server
  # token = ""

  # [secretstores.vault.approle]
  #   ## The Role ID for AppRole Authentication, a UUID string
  #   role_id = ""
  #
  #   ## Whether the Secret ID is configured to be response wrapped or not
  #   # response_wrapped = false
  #
  #   ## The Secret ID for AppRole Authentication
  #   secret = ""
```

[vault]: https://www.hashicorp.com/en/products/vault
