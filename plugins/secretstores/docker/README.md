# Docker Secrets Secret-Store Plugin

The `docker` plugin allows to utilize credentials and secrets mounted by
Docker during container runtime. The secrets are mounted as files
under the `/run/secrets` directory within the container.

> NOTE: This plugin can ONLY read the mounted secrets from Docker and NOT set them.

## Configuration

```toml @sample.conf
# File based Docker Secrets secret-store
[[secretstores.docker]]
  ## Unique identifier for the secretstore.
  ## This id can later be used in plugins to reference the secrets
  ## in this secret-store via @{<id>:<secret_key>} (mandatory)
  id = "docker_secretstore"

  ## Default Path to directory where docker stores the secrets file
  ## Current implementation in docker compose v2 only allows the following
  ## value for the path where the secrets are mounted at runtime
  # path = "/run/secrets"
```

Each Secret mentioned within a Compose service's `secrets` parameter will be
available as file under the `/run/secrets/<secret-name>` within the container.
