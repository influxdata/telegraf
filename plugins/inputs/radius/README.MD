# Radius Input Plugin

The Radius plugin collects radius response times.

### Configuration:

toml @sample.conf

### Measurements & Fields:

- radius
  - responsetime (float)

### Tags:

- All measurements have the following tags:
    - port
    - server

### Example Output:

```
radius,port=1812,server=debian-stretch-radius responsetime=0.011 1502489900000000000
```
