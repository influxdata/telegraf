# Exec Plugin

The exec plugin can execute arbitrary commands which return flattened
JSON.

For example, if you have a json-returning command called mycollector, you could
setup the exec plugin with:

```
[[exec.commands]]
command = "/usr/bin/mycollector --output=json"
name = "mycollector"
```

The name is used as a prefix for the measurements.
