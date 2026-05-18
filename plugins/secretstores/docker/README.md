# Docker Secret Store Plugin

This plugin allows to access [Docker secrets][docker_secrets] mounted by the
engine during container runtime. The secrets are accessible as files under
`/run/secrets` from within the container.

⭐ Telegraf v1.27.0
🏷️ containers
💻 all

[docker_secrets]: https://docs.docker.com/engine/swarm/secrets/

## Usage <!-- @/docs/includes/secret_usage.md -->

Secrets defined by a store are referenced with `@{<store-id>:<secret_key>}`
the Telegraf configuration. Only certain Telegraf plugins and options of
support secret stores. To see which plugins and options support
secrets, see their respective documentation (e.g.
`plugins/outputs/influxdb/README.md`). If the plugin's README has the
`Secret store support` section, it will detail which options support secret
store usage.

## Configuration

```toml @sample.conf
# Secret store to access docker secrets
[[secretstores.docker]]
  ## Unique identifier for the secret store.
  ## This id can later be used in plugins to reference the secrets
  ## in this secret store via @{<id>:<secret_key>} (mandatory)
  id = "docker_secretstore"

  ## Default Path to directory where docker stores the secrets file
  ## Current implementation in docker compose v2 only allows the following
  ## value for the path where the secrets are mounted at runtime
  # path = "/run/secrets"

  ## Allow dynamic secrets that are updated during runtime of telegraf
  ## Dynamic Secrets work only with `file` or `external` configuration
  ## in `secrets` section of the `docker-compose.yml` file
  # dynamic = false
```

Each Secret mentioned within a Compose service's `secrets` parameter will be
available as file under the `/run/secrets/<secret-name>` within the container.

It is possible to let Telegraf pick changed secret values into plugins by
setting `dynamic` to `true`. This feature will work only for Docker secrets
provided via `file` and `external` settings within the `docker-compose.yml` file
([see documentation][spec]) instead of using `environment` variables.

[spec]: https://github.com/compose-spec/compose-spec/blob/master/09-secrets.md

## Example Compose File

```yaml
services:
  telegraf:
    image: docker.io/telegraf:latest
    container_name: dockersecret_telegraf
    user: "${USERID}" # Required to access the /run/secrets directory in container
    secrets:
      - secret_for_plugin
    volumes:
      - /path/to/telegrafconf/host:/etc/telegraf/telegraf.conf:ro

secrets:
  secret_for_plugin:
    environment: TELEGRAF_PLUGIN_CREDENTIAL
```

here the `TELEGRAF_PLUGIN_CREDENTIAL` exists in a `.env` file in the same directory
as the `docker-compose.yml`. An example of the `.env` file can be as follows:

```env
TELEGRAF_PLUGIN_CREDENTIAL=superSecretStuff
# determine this value by executing `id -u` in terminal
USERID=1000
```

## Additional Information

This plugin only supports reading the secrets, it cannot create or modify them.
