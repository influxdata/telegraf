# ublox

## Configuration

```toml @sample.conf
[[inputs.ublox]]
    ublox_pty = "/tmp/ptyGPSRO_tlg"
```

## Data

Provide `ublox-data` with next fields:

- `active`            - boolean, true if ublox device provide correct data
- `lon`               - floating, longitude
- `lat`               - floating, latitude
- `heading`           - floating
- `horizontal_acc`    - floating
- `heading`           - floating
- `heading_of_motion` - floating
- `heading_acc`       - floating
- `heading_is_valid`  - boolean
- `speed`             - floating
- `speed_acc`         - floating
- `pdop`              - unsigned int (16)
- `hdop`              - unsigned int (16)
- `sat_num`           - unsigned int (8)
- `fix_type`          - unsigned int (8)

- `fusion_mode`             - unsigned int (8)
- `system_gps_time_diff_ms` - int (64)

- `sw_version` - string
- `hw_version` - string
- `fw_version` - string


Provide `ublox-data-sensors`

Fields:
- `s_status1`  - byte
- `s_status2`  - byte
- `s_freq`     - byte
- `s_faults`   - byte

Tags:
- `name`       - string
