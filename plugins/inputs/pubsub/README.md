# GCP PubSub Input Plugin

The GCP PubSub plugin ingests metrics from [Google Cloud PubSub][pubsub]
and creates metrics using one of the supported [input data formats][].


### Configuration

This section contains the default TOML to configure the plugin.  You can
generate it using `telegraf --usage pubsub`.

```toml
[[inputs.pubsub]]
## TODO(emilymye) Copy from pubsub.go
```

[pubsub]: https://cloud.google.com/pubsub
[input data formats]: /docs/DATA_FORMATS_INPUT.md
