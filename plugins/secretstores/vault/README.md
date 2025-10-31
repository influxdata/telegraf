# HashiCorp Vault Secret-Store Plugin

The `vault` plugin allows to utilize secrets stored in a HashiCorp
vault server via the Vault API. It supports authentication via AppRole.

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

```toml @sample.conf
# Secret-store to access Vault Secrets
[[secretstores.vault]]
  ## Unique identifier for the secretstore.
  ## This id can later be used in plugins to reference the secrets
  ## in this secret-store via @{<id>:<secret_key>} (mandatory)
  id = "vault_secretstore"

  ## Address of the Vault server
  address = "localhost:8200"

  ## Mount Path of the KV secrets engine
  mount_path = ""

  ## Path to the desired secrets within the KV secrets engine
  secret_path = ""

  ## Whether to use the older KV v1 secrets engine.
  ## By default will use the v2 engine.
  # use_kv_v1 = false

  [[secretstores.vault.approle]]
    ## The Role ID for AppRole Authentication, a UUID string
    role_id = ""

    ## Whether the Secret ID is configured to be response wrapped or not
    # response_wrapped = false

    ## The Secret ID for AppRole Authentication
    ## Only one of the following three options should be set. If multiple
    ## are set, the precedence is: secret_file > secret_env > secret_id
    # secret_file = ""
    # secret_env = ""
    # secret_id = ""
```
