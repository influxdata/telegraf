# Exec Plugin

The exec plugin can execute arbitrary commands which output JSON. Then it flattens JSON and finds
all numeric values, treating them as floats.

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


# Sample

Let's say that we have a command named "mycollector", which gives the following output:
```json
{
    "a": 0.5,
    "b": {
        "c": "some text",
        "d": 0.1,
        "e": 5
    }
}
```

The collected metrics will be:
```
exec_mycollector_a value=0.5
exec_mycollector_b_d value=0.1
exec_mycollector_b_e value=5
```
