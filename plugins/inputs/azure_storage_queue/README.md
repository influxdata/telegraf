# Telegraf Input Plugin: Azure Storage Queue

This plugin gathers sizes of Azure Storage Queues.

### Configuration:

```toml
# Description
[[inputs.azure_storage_queue]]
  ## Required Azure Storage Account name
  azure_storage_account_name = "mystorageaccount"

  ## Required Azure Storage Account access key
  azure_storage_account_key = "storageaccountaccesskey"
  
  ## Uncomment to disable peeking age of oldest message (saves time)
  # peek_oldest_message_age = false
```

### Measurements & Fields:

- azure_storage_queues:
  - size
  - oldest_message_age - Age of message at the head of the queue, in seconds

### Tags:

- azure_storage_queues:
  - name
  - storage_account