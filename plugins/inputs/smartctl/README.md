# smartctl JSON Input Plugin

This plugin collects [Self-Monitoring, Analysis and Reporting Technology][smart]
information for storage devices information using the
[`smartmontools`][smartmon] package. Contrary to the
[smart plugin][smart_plugin], this plugin does not use the [`nvme-cli`][nvmecli]
package to collect additional information about NVMe devices.

> [!NOTE]
> This plugin requires [`smartmontools`][smartmon] to be installed on your
> system. The `smartctl` command must to be executable by Telegraf and must
> supporting JSON output. JSON output was added in v7.0 and improved in
> subsequent releases

⭐ Telegraf v1.31.0
🏷️ hardware, system
💻 all

[smart]: https://en.wikipedia.org/wiki/Self-Monitoring,_Analysis_and_Reporting_Technology
[smart_plugin]: /plugins/inputs/smart/README.md
[nvmecli]: https://github.com/linux-nvme/nvme-cli

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

Plugins support additional global and plugin configuration settings for tasks
such as modifying metrics, tags, and fields, creating aliases, and configuring
plugin ordering. See [CONFIGURATION.md][CONFIGURATION.md] for more details.

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
    ## `smartctl --scan-open`. Only one option is allowed at a time. If set, include
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

### Permissions

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

## Troubleshooting

This plugin uses the following commands to determine devices and collect
metrics:

- `smartctl --json --scan-open`
- `smartctl --json --all $DEVICE --device $TYPE --nocheck=$NOCHECK`

Please include the output of the above two commands for all devices that are
having issues.

## Metrics

- smartctl
  - tags
    - model (model name of the storage device)
    - name (device id in the system)
    - serial (serial number of the device)
    - type (device type like SATA etc)
    - wwn (world wide number of the device)
  - fields
    - depending on the device information

- smartctl_attributes
  - tags
    - model (model name of the storage device)
    - name (name of the attribute)
    - serial (serial number of the device)
    - type (device type like SATA etc)
    - wwn (world wide number of the device)
  - fields
    - raw_value (integer)
    - threshold (integer)
    - value (integer)
    - worst (integer)

## Example Output

```text
smartctl,model=SanDisk\ pSSD,name=/dev/sda,serial=06c9f4c44,type=sat,wwn=5001b4409f6c444c capacity=15693664256i,firmware="3",health_ok=true,logical_block_size=512i,power_on_hours=11i,temperature=0i 1711480345675066854
smartctl_attributes,model=SanDisk\ pSSD,name=Reallocated_Sector_Ct,serial=06c9f4c44,type=sat,wwn=5001b4409f6c444c raw_value=0i,threshold=0i,value=100i,worst=100i 1711480345675066854
smartctl_attributes,model=SanDisk\ pSSD,name=Power_On_Hours,serial=06c9f4c44,type=sat,wwn=5001b4409f6c444c raw_value=11i,threshold=0i,value=100i,worst=100i 1711480345675066854
smartctl_attributes,model=SanDisk\ pSSD,name=Power_Cycle_Count,serial=06c9f4c44,type=sat,wwn=5001b4409f6c444c raw_value=223i,threshold=0i,value=100i,worst=100i 1711480345675066854
smartctl_attributes,model=SanDisk\ pSSD,name=Program_Fail_Count,serial=06c9f4c44,type=sat,wwn=5001b4409f6c444c raw_value=0i,threshold=0i,value=100i,worst=100i 1711480345675066854
smartctl_attributes,model=SanDisk\ pSSD,name=Erase_Fail_Count,serial=06c9f4c44,type=sat,wwn=5001b4409f6c444c raw_value=0i,threshold=0i,value=100i,worst=100i 1711480345675066854
smartctl_attributes,model=SanDisk\ pSSD,name=Avg_Write/Erase_Count,serial=06c9f4c44,type=sat,wwn=5001b4409f6c444c raw_value=3i,threshold=0i,value=100i,worst=100i 1711480345675066854
smartctl_attributes,model=SanDisk\ pSSD,name=Unexpect_Power_Loss_Ct,serial=06c9f4c44,type=sat,wwn=5001b4409f6c444c raw_value=114i,threshold=0i,value=100i,worst=100i 1711480345675066854
smartctl_attributes,model=SanDisk\ pSSD,name=Reported_Uncorrect,serial=06c9f4c44,type=sat,wwn=5001b4409f6c444c raw_value=0i,threshold=0i,value=100i,worst=100i 1711480345675066854
smartctl_attributes,model=SanDisk\ pSSD,name=Perc_Write/Erase_Count,serial=06c9f4c44,type=sat,wwn=5001b4409f6c444c raw_value=10i,threshold=0i,value=100i,worst=100i 1711480345675066854
smartctl_attributes,model=SanDisk\ pSSD,name=Perc_Avail_Resrvd_Space,serial=06c9f4c44,type=sat,wwn=5001b4409f6c444c raw_value=0i,threshold=5i,value=100i,worst=100i 1711480345675066854
smartctl_attributes,model=SanDisk\ pSSD,name=Perc_Write/Erase_Ct_BC,serial=06c9f4c44,type=sat,wwn=5001b4409f6c444c raw_value=0i,threshold=0i,value=100i,worst=100i 1711480345675066854
smartctl_attributes,model=SanDisk\ pSSD,name=Total_LBAs_Written,serial=06c9f4c44,type=sat,wwn=5001b4409f6c444c raw_value=10171055i,threshold=0i,value=100i,worst=100i 1711480345675066854
smartctl_attributes,model=SanDisk\ pSSD,name=Total_LBAs_Read,serial=06c9f4c44,type=sat,wwn=5001b4409f6c444c raw_value=94845144i,threshold=0i,value=100i,worst=100i 1711480345675066854
```
