# OpenSMTPD Input Plugin

This plugin gathers statistics from [OpenSMTPD][opensmtp] using the `smtpctl`
binary.

> [!NOTE]
> The `smtpctl` binary must be present on the system and executable by Telegraf.
> The plugin supports using `sudo` for execution.

⭐ Telegraf v1.5.0
🏷️ server, network
💻 all

[opensmtp]: https://www.opensmtpd.org/

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# A plugin to collect stats from Opensmtpd - a validating, recursive, and caching DNS resolver
 [[inputs.opensmtpd]]
   ## If running as a restricted user you can prepend sudo for additional access:
   #use_sudo = false

   ## The default location of the smtpctl binary can be overridden with:
   binary = "/usr/sbin/smtpctl"

   # The default timeout of 1s can be overridden with:
   #timeout = "1s"
```

### Permissions

It's important to note that this plugin references `smtpctl`, which may require
additional permissions to execute successfully. Depending on the user/group
permissions of the telegraf user executing this plugin, you may need to alter
the group membership, set facls, or use sudo.

#### Group membership (recommended)

```bash
$ groups telegraf
telegraf : telegraf

$ usermod -a -G opensmtpd telegraf

$ groups telegraf
telegraf : telegraf opensmtpd
```

#### Sudo privileges

If you use this method, you will need the following in your telegraf config:

```toml
[[inputs.opensmtpd]]
  use_sudo = true
```

You will also need to update your sudoers file:

```bash
$ visudo
# Add the following line:
Cmnd_Alias SMTPCTL = /usr/sbin/smtpctl
telegraf  ALL=(ALL) NOPASSWD: SMTPCTL
Defaults!SMTPCTL !logfile, !syslog, !pam_session
```

Please use the solution you see as most appropriate.

## Metrics

This is the full list of statistics provided by smtpctl and potentially
collected by telegraf depending of your smtpctl configuration.

- smtpctl
    bounce_envelope
    bounce_message
    bounce_session
    control_session
    mda_envelope
    mda_pending
    mda_running
    mda_user
    mta_connector
    mta_domain
    mta_envelope
    mta_host
    mta_relay
    mta_route
    mta_session
    mta_source
    mta_task
    mta_task_running
    queue_bounce
    queue_evpcache_load_hit
    queue_evpcache_size
    queue_evpcache_update_hit
    scheduler_delivery_ok
    scheduler_delivery_permfail
    scheduler_delivery_tempfail
    scheduler_envelope
    scheduler_envelope_expired
    scheduler_envelope_incoming
    scheduler_envelope_inflight
    scheduler_ramqueue_envelope
    scheduler_ramqueue_message
    scheduler_ramqueue_update
    smtp_session
    smtp_session_inet4
    smtp_session_local
    uptime

## Example Output

```text
opensmtpd,host=localhost scheduler_delivery_tempfail=822,mta_host=10,mta_task_running=4,queue_bounce=13017,scheduler_delivery_permfail=51022,mta_relay=7,queue_evpcache_size=2,scheduler_envelope_expired=26,bounce_message=0,mta_domain=7,queue_evpcache_update_hit=848,smtp_session_local=12294,bounce_envelope=0,queue_evpcache_load_hit=4389703,scheduler_ramqueue_update=0,mta_route=3,scheduler_delivery_ok=2149489,smtp_session_inet4=2131997,control_session=1,scheduler_envelope_incoming=0,uptime=10346728,scheduler_ramqueue_envelope=2,smtp_session=0,bounce_session=0,mta_envelope=2,mta_session=6,mta_task=2,scheduler_ramqueue_message=2,mta_connector=7,mta_source=1,scheduler_envelope=2,scheduler_envelope_inflight=2 1510220300000000000
```
