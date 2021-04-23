# Spectrum Scale (GPFS) Input Plugin

[Spectrum Scale](https://www.ibm.com/ca-en/marketplace/scale-out-file-and-object-storage) input plugin gathers metrics from GPFS.

This plugin requires read-write access to Spectrum Scale monitoring socket (`/var/mmfs/mmpmon/mmpmonSocket`).
This can be accomplished by granting telegraf root privileges, but an appropriate ACL should be sufficient.
Tested against versions 4.2+ and 5.0+

### Configuration:

```toml
# SampleConfig
[[inputs.spectrum_scale]]
  # An array of Spectrum Scale (GPFS) sensors
  #
  # These will be monitored through the mmpmon socket
  sensors = ["fis", "nsd_ds", "vfss"]
  socketLocation = "/var/mmfs/mmpmon/mmpmonSocket"

```

#### `sensors`

List of spectrum scale subsystems to query. Examples include: `fis`, `nsd_ds`, and `vfss`

#### `socketLocation`

Location of the mmpmon monitoring socket. By default, `/var/mmfs/mmpmon/mmpmonSocket`

