# LVM Input Plugin

The Logical Volume Management (LVM) input plugin collects information about
physical volumes, volume groups, and logical volumes.

## Configuration

The `lvm` command requires elevated permissions. If the user has configured
sudo with the ability to run these commands, then set the `use_sudo` to true.

```toml
# Read metrics about LVM physical volumes, volume groups, logical volumes.
[[inputs.lvm]]
  ## Use sudo to run LVM commands
  use_sudo = false
```

### Using sudo

If your account does not already have the ability to run commands
with passwordless sudo then updates to the sudoers file are required. Below
is an example to allow the requires LVM commands:

First, use the `visudo` command to start editing the sudoers file. Then add
the following content, where `<username>` is the username of the user that
needs this access:

```text
Cmnd_Alias LVM = /usr/sbin/pvs *, /usr/sbin/vgs *, /usr/sbin/lvs *
<username>  ALL=(root) NOPASSWD: LVM
Defaults!LVM !logfile, !syslog, !pam_session
```

## Metrics

Metrics are broken out by physical volume (pv), volume group (vg), and logical
volume (lv):

- lvm_physical_vol
  - tags
    - path
    - vol_group
  - fields
    - size
    - free
    - used
    - used_percent
- lvm_vol_group
  - tags
    - name
  - fields
    - size
    - free
    - used_percent
    - physical_volume_count
    - logical_volume_count
    - snapshot_count
- lvm_logical_vol
  - tags
    - name
    - vol_group
  - fields
    - size
    - data_percent
    - meta_percent

## Example Output

The following example shows a system with the root partition on an LVM group
as well as with a Docker thin-provisioned LVM group on a second drive:

```shell
> lvm_physical_vol,path=/dev/sda2,vol_group=vgroot free=0i,size=249510756352i,used=249510756352i,used_percent=100 1631823026000000000
> lvm_physical_vol,path=/dev/sdb,vol_group=docker free=3858759680i,size=128316342272i,used=124457582592i,used_percent=96.99277612525741 1631823026000000000
> lvm_vol_group,name=vgroot free=0i,logical_volume_count=1i,physical_volume_count=1i,size=249510756352i,snapshot_count=0i,used_percent=100 1631823026000000000
> lvm_vol_group,name=docker free=3858759680i,logical_volume_count=1i,physical_volume_count=1i,size=128316342272i,snapshot_count=0i,used_percent=96.99277612525741 1631823026000000000
> lvm_logical_vol,name=lvroot,vol_group=vgroot data_percent=0,metadata_percent=0,size=249510756352i 1631823026000000000
> lvm_logical_vol,name=thinpool,vol_group=docker data_percent=0.36000001430511475,metadata_percent=1.3300000429153442,size=121899057152i 1631823026000000000
```
