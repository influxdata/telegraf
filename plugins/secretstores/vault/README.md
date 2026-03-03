# HashiCorp Vault Secret-Store Plugin

The `vault` plugin allows to utilize secrets stored in a
[HashiCorp Vault][vault] server via the Vault API. It supports authentication
via AppRole.

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

  # [secretstores.vault.approle]
  #   ## The Role ID for AppRole Authentication, a UUID string
  #   role_id = ""
  #
  #   ## Whether the Secret ID is configured to be response wrapped or not
  #   # response_wrapped = false
  #
  #   ## The Secret ID for AppRole Authentication
  #   secret = ""

  # [secretstores.vault.aws_ec2]
  #   ## The Role Name for AWS EC2 authentication
  #   role_name = ""
  #
  #   ## The AWS region, defaulting to "us-east-1" if unset
  #   # region = "us-east-1"
  #
  #   ## The signature type to use, defaulting to "pkcs7"
  #   ## Allowed options: "pkcs7", "identity", "rsa2048"
  #   # signature_type = "pkcs7"

  # ## Credentials will be set using the values in the environment variables:
  # ## - AWS_ACCESS_KEY_ID
  # ## - AWS_SECRET_ACCESS_KEY
  # ## - AWS_SESSION_TOKEN
  # ## To specify a path to a credentials file instead, set:
  # ## - AWS_SHARED_CREDENTIALS_FILE
  # [secretstores.vault.aws_iam]
  #   ## The Role Name for AWS IAM authentication
  #   role_name = ""
  #
  #   ## The AWS region, defaulting to "us-east-1" if unset
  #   # region = "us-east-1"
  #
  #   ## An optional server ID header to provide, with the key
  #   ## "X-Vault-AWS-IAM-Server-ID"
  #   # server_id_header = ""

  # [secretstores.vault.azure]
  #   ## The Role Name for Azure authentication
  #   role_name = ""
  #
  #   ## The Azure Resource URL to use as the aud value on the JWT token to
  #   ## use rather than the default of Azure Public Cloud's ARM URL.
  #   ## Defaults to "https://management.azure.com/"
  #   # resource_url = "https://management.azure.com/"

  # [secretstores.vault.kubernetes]
  #   ## The Kubernetes service account role name
  #   role_name = ""
  #
  #   ## The Kubernetes service account token
  #   secret = ""

  # [secretstores.vault.userpass]
  #   ## The Vault Userpass username
  #   username = ""
  #
  #   ## The Vault Userpass password
  #   password = ""

```

[vault]: https://www.hashicorp.com/en/products/vault
