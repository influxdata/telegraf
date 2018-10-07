# SystemD Input Plugin

Read unit metrics of systemd using dbus. This plugin only works on linux.

### Configuration:

```toml
[[inputs.systemd]]
  ## List of unit regex pattern
  unit_pattern = [".*"]
```

### Metrics:

- systemd
  - tags:
    - unit_name
    - unit_type
  - fields:
    - is_active (integer)
    - active_enter_timestamp (integer)
    - last_trigger_usec (integer, only timer unit)
    - n_restarts (integer, only service unit)
    - n_accepted (integer, only socket unit)
    - n_connection (integer, only socket unit)
    - n_refused (integer, only socket unit)