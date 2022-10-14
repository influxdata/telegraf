# ublox

## Configuration

```toml @sample.conf
[[inputs.simple]]
    ublox_pty = "/tmp/ptyGPSRO_tlg"
```

## Data

Provide `ublox-data` with next fields:

- `active`  - boolean, true if ublox device provide correct data
- `lon`     - floating, longitude
- `lat`     - floating, latitude
- `heading` - floating
- `pdop`    - unsigned int (16)
