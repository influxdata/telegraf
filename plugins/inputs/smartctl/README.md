# smartctl JSON Input Plugin

Get metrics using the command line utility `smartctl` for S.M.A.R.T.
(Self-Monitoring, Analysis and Reporting Technology) storage devices. SMART is a
monitoring system included in computer hard disk drives (HDDs), solid-state
drives (SSDs), and nVME drives that detects and reports on various indicators of
drive reliability, with the intent of enabling the anticipation of hardware
failures.

This version of the plugin requires support of the JSON flag from the `smartctl`
command. This flag was added in 7.0 (2019) and further enhanced in subsequent
releases.

See smartmontools (<https://www.smartmontools.org/>) for more information.

## smart vs smartctl

The smartctl plugin is an alternative to the smart plugin. The biggest
difference is that the smart plugin can also call `nvmectl` to collect
additional details about NVMe devices as well as some vendor specific device
information.

This plugin will also require a version of the `smartctl` command that supports
JSON output versus the smart plugin will parse the raw output.

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Read metrics from SMART storage devices using smartclt's JSON output
[[inputs.smartctl]]
    ## Optionally specify the path to the smartctl executable
    # path = "/usr/sbin/smartctl"

    ## Use sudo
    ## On most platforms used, smartctl requires root access. Setting 'use_sudo'
    ## to true will make use of sudo to run smartctl. Sudo must be configured to
    ## allow the telegraf user to run smartctl without a password.
    # use_sudo = false

    ## Devices to include or exclude
    ## By default, the plugin will use all devices found in the output of
    ## `smartctl --scan`. Only one option is allowed at a time. If set, include
    ## sets the specific devices to scan, while exclude omits specific devices.
    # devices_include = []
    # devices_exclude = []

    ## Skip checking disks in specified power mode
    ## Defaults to "standby" to not wake up disks that have stopped rotating.
    ## For full details on the options here, see the --nocheck section in the
    ## smartctl man page. Choose from:
    ##   * never: always check the device
    ##   * sleep: check the device unless it is in sleep mode
    ##   * standby: check the device unless it is in sleep or standby mode
    ##   * idle: check the device unless it is in sleep, standby, or idle mode
    # nocheck = "standby"

    ## Timeout for the cli command to complete
    # timeout = "30s"
```

## Permissions

It is important to note that this plugin references `smartctl`, which may
require additional permissions to execute successfully.  Depending on the
user/group permissions of the telegraf user executing this plugin, users may
need to use sudo.

Users need the following in the Telegraf config:

```toml
[[inputs.smart_json]]
  use_sudo = true
```

And to update the `/etc/sudoers` file to allow running smartctl:

```bash
$ visudo
# Add the following lines:
Cmnd_Alias SMARTCTL = /usr/sbin/smartctl
telegraf  ALL=(ALL) NOPASSWD: SMARTCTL
Defaults!SMARTCTL !logfile, !syslog, !pam_session
```

## Debugging Issues

This plugin uses the following commands to determine devices and collect
metrics:

* `smartctl --json --scan`
* `smartctl --json --all $DEVICE --device $TYPE --nocheck=$NOCHECK`

Please include the output of the above two commands for all devices that are
having issues.

## Metrics

## Example Output

```text
```
