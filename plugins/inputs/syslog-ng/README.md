# syslog-ng Input Plugin

This plugin gathers stats from [syslog-ng](https://www.syslog-ng.com/) -
an enhanced log daemon.

## Configuration

```toml
# A plugin to collect stats from the syslog-ng daemon
[[inputs.syslog-ng]]
  ## If running as a restricted user you can prepend sudo for additional access:
  # use_sudo = false

  ## The default location of the syslog-ng-ctl binary can be overridden with:
  # binary = "/usr/local/sbin/syslog-ng-ctl"

  ## The default timeout of 1s can be overridden with:
  # timeout = "1s"
```

### Permissions

It's important to note that this plugin references syslog-ng-ctl, which
may require additional permissions to execute successfully.  Depending
on the user/group permissions of the telegraf user executing this
plugin, you may need to alter the group membership, set facls, or use
sudo.

**Group membership**:

```bash
$ groups telegraf
telegraf : telegraf

$ usermod -a -G syslog-ng telegraf

$ groups telegraf
telegraf : telegraf syslog-ng
```

**Sudo privileges (Recommended)**:
If you use this method, you will need the following in your telegraf config:

```toml
[[inputs.syslog-ng]]
  use_sudo = true
```

You will also need to update your sudoers file:

```bash
$ visudo
# Add the following line:
Cmnd_Alias SYSLOGNGCTL = /usr/local/sbin/syslog-ng-ctl
telegraf  ALL=(ALL) NOPASSWD: SYSLOGNGCTL
Defaults!SYSLOGNGCTL !logfile, !syslog, !pam_session
```

Please use the solution you see as most appropriate.

## Metrics

This is the full list of stats provided by syslog-ng. In the output, the
pascal-case in the syslog-ng-ctl stat name are replaced by snake case
(see
<https://www.syslog-ng.com/technical-documents/doc/syslog-ng-open-source-edition/3.16/administration-guide/80>
for details).

- syslog-ng
  - tags:
    - type
    - source_name
  - fields:
    - number
