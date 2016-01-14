# Exec Plugin

The exec plugin can execute arbitrary commands which output JSON. Then it flattens JSON and finds
all numeric values, treating them as floats.

For example, if you have a json-returning command called mycollector, you could
setup the exec plugin with:

```
[[inputs.exec]]
  command = "/usr/bin/mycollector --output=json"
  name_suffix = "_mycollector"
  interval = 10
```

The name suffix is appended to exec as "exec_name_suffix" to identify the input stream.

The interval is used to determine how often a particular command should be run. Each
time the exec plugin runs, it will only run a particular command if it has been at least
`interval` seconds since the exec plugin last ran the command.


# Sample

Let's say that we have a command with the name_suffix "_mycollector", which gives the following output:
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

The collected metrics will be stored as field values under the same measurement "exec_mycollector":
```
 exec_mycollector a=0.5,b_c="some text",b_d=0.1,b_e=5 1452815002357578567
```
