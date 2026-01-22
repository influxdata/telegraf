# Azure Queue Storage Input Plugin

This plugin gathers queue sizes from the [Azure Queue Storage][azure_queues]
service, storing a large numbers of messages.

⭐ Telegraf v1.13.0
🏷️ cloud
💻 all

[azure_queues]: https://learn.microsoft.com/en-us/azure/storage/queues

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

Plugins support additional global and plugin configuration settings for tasks
such as modifying metrics, tags, and fields, creating aliases, and configuring
plugin ordering. See [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Gather Azure Storage Queue metrics
[[inputs.azure_storage_queue]]
  ## Azure Storage Account name and shared access key (required)
  account_name = "mystorageaccount"
  account_key = "storageaccountaccesskey"

  ## Disable peeking age of oldest message (faster)
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

```text
azure_storage_queues,queue=myqueue,account=mystorageaccount oldest_message_age=799714900i,size=7i 1565970503000000000
azure_storage_queues,queue=myemptyqueue,account=mystorageaccount size=0i 1565970502000000000
```
