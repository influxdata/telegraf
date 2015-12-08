# expvar plugin

The expvar plugin collects InfluxDB-style expvar data from JSON endpoints.

With a configuration of:

```toml
[plugins.expvar]
  [[plugins.expvar.services]]
    name = "produce"
    urls = [
      "http://127.0.0.1:8086/debug/vars",
      "http://192.168.2.1:8086/debug/vars"
    ]
```

And if 127.0.0.1 responds with this JSON:

```json
{
  "k1": {
    "name": "fruit",
    "tags": {
      "kind": "apple"
    },
    "values": {
      "inventory": 371,
      "sold": 112
    }
  },
  "k2": {
    "name": "fruit",
    "tags": {
      "kind": "banana"
    },
    "values": {
      "inventory": 1000,
      "sold": 403
    }
  }
}
```

And if 192.168.2.1 responds like so:

```json
{
  "k3": {
    "name": "transactions",
    "tags": {},
    "values": {
      "total": 100,
      "balance": 184.75
    }
  }
}
```

Then the collected metrics will be:

```
expvar_produce_fruit,expvar_url='http://127.0.0.1:8086/debug/vars',kind='apple' inventory=371.0,sold=112.0
expvar_produce_fruit,expvar_url='http://127.0.0.1:8086/debug/vars',kind='banana' inventory=1000.0,sold=403.0

expvar_produce_transactions,expvar_url='http://192.168.2.1:8086/debug/vars' total=100.0,balance=184.75
```

There are two important details to note about the collected metrics:

1. Even though the values in JSON are being displayed as integers, the metrics are reported as floats.
JSON encoders usually don't print the fractional part for round floats.
Because you cannot change the type of an existing field in InfluxDB, we assume all numbers are floats.

2. The top-level keys' names (in the example above, `"k1"`, `"k2"`, and `"k3"`) are not considered when recording the metrics.
