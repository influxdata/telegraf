# Postfix Input Plugin

This plugin collects metrics on a local [Postfix][postfix] instance reporting
the length, size and age of the active, hold, incoming, maildrop, and deferred
[queues][queues].

⭐ Telegraf v1.5.0
🏷️ server
💻 freebsd, linux, macos, solaris

[postfix]: https://www.postfix.org/
[queues]: https://www.postfix.org/QSHAPE_README.html#queues

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

Plugins support additional global and plugin configuration settings for tasks
such as modifying metrics, tags, and fields, creating aliases, and configuring
plugin ordering. See [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Measure postfix queue statistics
# This plugin ONLY supports non-Windows
[[inputs.postfix]]
  ## Postfix queue directory. If not provided, telegraf will try to use
  ## 'postconf -h queue_directory' to determine it.
  # queue_directory = "/var/spool/postfix"
```

### Permissions

Telegraf will need read access to the files in the queue directory.  You may
need to alter the permissions of these directories to provide access to the
telegraf user.

This can be setup either using standard unix permissions or with Posix ACLs,
you will only need to use one method:

Unix permissions:

```sh
sudo chgrp -R telegraf /var/spool/postfix/{active,hold,incoming,deferred}
sudo chmod -R g+rXs /var/spool/postfix/{active,hold,incoming,deferred}
sudo usermod -a -G postdrop telegraf
sudo chmod g+r /var/spool/postfix/maildrop
```

Posix ACL:

```sh
sudo setfacl -Rm g:telegraf:rX /var/spool/postfix/
sudo setfacl -dm g:telegraf:rX /var/spool/postfix/
```

## Metrics

- postfix_queue
  - tags:
    - queue
  - fields:
    - length (integer)
    - size (integer, bytes)
    - age (integer, seconds)

## Example Output

```text
postfix_queue,queue=active length=3,size=12345,age=9
postfix_queue,queue=hold length=0,size=0,age=0
postfix_queue,queue=maildrop length=1,size=2000,age=2
postfix_queue,queue=incoming length=1,size=1020,age=0
postfix_queue,queue=deferred length=400,size=76543210,age=3600
```
