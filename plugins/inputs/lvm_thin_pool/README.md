# LVM Thin Pool Input Plugin

TO EDIT:
The lvm_thin_pool gathers size and usage data from LVM thin pools.
It runs the "lvdisplay" utility with filtered output options showing
columns "lv_size", "lv_metadata", "data_percent", "metadata_percent" and
"thin_count".

Results are tagged with the volume path.

As telegraf doesn't have the permission to run lvdisplay, one should
configure and use sudo to make the plugin work properly.

### Using sudo

You will need the following in your telegraf config:
```toml
[[inputs.lvm_thin_pool]]
  use_sudo = true
```

You will also need to update your sudoers file:
```bash
$ visudo
# Add the following line:
Cmnd_Alias LVDISPLAY = /usr/sbin/lvdisplay
telegraf  ALL=(root) NOPASSWD: LVDISPLAY
Defaults!LVDISPLAY !logfile, !syslog, !pam_session
```

### Configuration

```toml
  [[inputs.lvm_thin_pool]]
    ## Adjust your sudo settings appropriately if using this option
    use_sudo = false
    # set path to the thin pool and use it as tag
    path = "my_vg/my_thin_pool"
```

### Example Output

```
$ sudo lvdisplay -C -o lv_size,lv_metadata_size,data_percent,metadata_percent,thin_count --units m --noheadings --separator , my_vg/my_thin_pool
  20456.00m,20.00m,22.23,21.13,3
```

```
$ telegraf --config telegraf.conf --input-filter lvm_thin_pool --test --debug
* Plugin: inputs.lvm_thin_pool, Collection 1
> lvm_thin_pool,host=vagrant-1,path=my_vg/my_thin_pool data_percent=22.26,lv_metadata=20,lv_size=20456,metadata_percent=21.15,thin_count=3i 161946885400000000
```
