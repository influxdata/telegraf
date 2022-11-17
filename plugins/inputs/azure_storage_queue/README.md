# Azure Storage Queue Input Plugin

This plugin gathers sizes of Azure Storage Queues.

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md

## Configuration

```toml @sample.conf
# Gather Azure Storage Queue metrics
[[inputs.azure_storage_queue]]
  ## Required Azure Storage Account name
  account_name = "mystorageaccount"

  ## Required Azure Storage Account access key
  account_key = "storageaccountaccesskey"

  ## Set to false to disable peeking age of oldest message (executes faster)
  # peek_oldest_message_age = true
```

## Metrics

- azure_storage_queues
  - tags:
    - queue
    - account
  - fields:
    - size (integer, count)
    - oldest_message_age_ns (integer, nanoseconds) Age of message at the head
      of the queue. Requires `peek_oldest_message_age` to be configured
      to `true`.

## Example Output

```shell
azure_storage_queues,queue=myqueue,account=mystorageaccount oldest_message_age=799714900i,size=7i 1565970503000000000
azure_storage_queues,queue=myemptyqueue,account=mystorageaccount size=0i 1565970502000000000
```
