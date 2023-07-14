# OAuth2 Secret-store Plugin

The `oauth2` plugin allows to retrieve and maintain secrets from various OAuth2
services such as [Auth0][auth0], [AzureAD][azuread] or others (see
[Configuration section](#configuration)).
Tokens that are expired or are about to expire will be automatically renewed
by this secret-store, so other plugins referencing those tokens can then use
them to perform their API calls without hassle.

**Please note:** This plugin only supports the *2-legged client credentials*
flow.

You can use Telegraf to test token retrieval. Run

```shell
telegraf secrets help
```

to get more information on how to do access secrets with Telegraf.

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
# Secret-store to retrieve and maintain tokens from various OAuth2 services
[[secretstores.oauth2]]
  ## Unique identifier for the secret-store.
  ## This id can later be used in plugins to reference the secrets
  ## in this secret-store via @{<id>:<secret_key>} (mandatory)
  id = "secretstore"

  ## Service to retrieve the token(s) from
  ## Currently supported services are "custom", "auth0" and "AzureAD"
  # service = "custom"

  ## Setting to overwrite the queried token-endpoint
  ## This setting is optional for some serices but mandatory for others such
  ## as "custom" or "auth0". Please check the documentation at
  ## https://github.com/influxdata/telegraf/blob/master/plugins/secretstores/oauth2/README.md
  # token_endpoint = ""

  ## Tenant ID for the AzureAD service
  # tenant_id = ""

  ## Minimal remaining time until the token expires
  ## If a token expires less than the set duration in the future, the token is
  ## renewed. This is useful to avoid race-condition issues where a token is
  ## still valid, but isn't when the request reaches the API endpoint of
  ## your service using the token.
  # token_expiry_margin = "1s"

  ## Section for defining a token secret
  [[secretstores.oauth2.token]]
    ## Unique secret-key used for referencing the token via @{<id>:<secret_key>}
    key = ""
    ## Client-ID and secret for the 2-legged OAuth flow
    client_id = ""
    client_secret = ""
    ## Scopes to send in the request
    # scopes = []

    ## Additional (optional) parameters to include in the token request
    ## This might for example include the "audience" parameter required for
    ## auth0.
    # [secretstores.oauth2.token.parameters]
    #     audience = ""
```

All services allow multiple `[[secretstores.oauth2.token]]` sections to be
specified to define different tokens for the secret store. Please make sure to
specify `key`s that are **unique** within the secret-store instance as those
are used to reference the tokens/secrets later.

The `oauth2` secret-store supports various services that might differ in the
required or allowed settings as listed below. All of the services accept
optional `scopes` and optional `parameter` settings if not stated otherwise.

Please **replace the placeholders** in the minumal example configurations below
and add `scopes` and/or `parameters` if required.

### Auth0

To use the [Auth0 service][auth0] for retrieving the token you need to set the
`token_endpoint` to your application's endpoint. Furthermore, specifying the
`audience` parameter is required. An example configuration look like

```toml
[[secretstores.oauth2]]
  id = "secretstore"
  service = "auth0"
  token_endpoint = "https://YOUR_DOMAIN/oauth/token"

  [[secretstores.oauth2.token]]
    key = "mytoken"
    client_id = "YOUR_CLIENT_ID"
    client_secret = "YOUR_CLIENT_SECRET"

    [secretstores.oauth2.token.parameters]
        audience = "YOUR_API_IDENTIFIER"
```

### AzureAD

To use the [AzureAD service][azuread] for retrieving the token you need to set
the `tenant_id` and provide a valid `scope`. An example configuration look like

```toml
[[secretstores.oauth2]]
  id = "secretstore"
  service = "AzureAD"
  tenant_id = "YOUR_TENANT_ID"

  [[secretstores.oauth2.token]]
    key = "mytoken"
    client_id = "YOUR_CLIENT_ID"
    client_secret = "YOUR_CLIENT_SECRET"
    scopes = ["YOUR_CLIENT_ID/.default"]
```

### Custom service

If your service is not listed above, you can still use it setting
`service = "custom"` as well as the `token_endpoint`. Please make sure your
service is configured for the *2-legged client credentials* OAuth2 flow!

[auth0]: https://auth0.com
[azuread]: https://azure.microsoft.com/en/products/active-directory
