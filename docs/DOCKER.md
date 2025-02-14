# Docker Images

Telegraf is available as an [Official image][] on DockerHub. Official images
are a curated set of Docker Images that also automatically get security updates
from Docker, follow a set of best practices, and are available via a shortcut
syntax which omits the organization.

InfluxData maintains Debian and Alpine based images across the last three
minor releases. To pull the latest Telegraf images:

```shell
# latest Debian-based image
docker pull telegraf
# latest Alpine-based image
docker pull telegraf:alpine
```

See the [Telegraf DockerHub][] page for complete details on available images,
versions, and tags.

[official image]: https://docs.docker.com/trusted-content/official-images/
[Telegraf DockerHub]: https://hub.docker.com/_/telegraf

## Nightly Images

[Nightly builds][] are available and are generated from the master branch each
day at around midnight UTC. The artifacts include both binary packages, RPM &
DEB packages, as well as nightly Docker images that are hosted on [quay.io][].

[Nightly builds]: /docs/NIGHTLIES.md
[quay.io]: https://quay.io/repository/influxdb/telegraf-nightly?tab=tags&tag=latest

## Dockerfiles

The [Dockerfiles][] for these images are available for users to use as well.

[Dockerfiles]: https://github.com/influxdata/influxdata-docker

## Lockable Memory

Telegraf does require the ability to use lockable memory when running by default. In some
deployments for Docker a container may not have enough lockable memory, which
results in the following warning:

```text
W! Insufficient lockable memory 64kb when 72kb is required. Please increase the limit for Telegraf in your Operating System!
```

or this error:

```text
panic: could not acquire lock on 0x7f7a8890f000, limit reached? [Err: cannot allocate memory]
```

Users have two options:

1. Increase the ulimit in the container. The user does this with the `ulimit -l`
  command. To both see and set the value. For docker, there is a `--ulimit` flag
  that could be used, like `--ulimit memlock=8192:8192` as well.
2. Add the `--unprotected` flag to the command arguments to not use locked
  memory and instead store secrets in unprotected memory. This is less secure
  as secrets could find their way into paged out memory and can be written to
  disk unencrypted, therefore this is opt-in. For docker look at updating the
  `CMD` used to include this flag.
