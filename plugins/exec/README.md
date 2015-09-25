# Exec Plugin

The exec plugin can execute arbitrary commands which return flattened
JSON.

For example, if you have a json-returning command called mycollector, you could
setup the exec plugin with:

```
[[exec.commands]]
command = "/usr/bin/mycollector --output=json"
name = "mycollector"
interval = 10
```

The name is used as a prefix for the measurements.

The interval is used to determine how often a particular command should be run. Each
time the exec plugin runs, it will only run a particular command if it has been at least
`interval` seconds since the exec plugin last ran the command.
