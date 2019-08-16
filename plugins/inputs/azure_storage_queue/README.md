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
  
  ## Set to false to disable peeking age of oldest message (executes faster)
  # peek_oldest_message_age = true
```

### Metrics
- azure_storage_queues
  - tags:
    - name
    - storage_account
  - fields:
    - size (integer, count)
    - oldest_message_age_ns (integer, nanoseconds) Age of message at the head of the queue.
      Requires `peek_oldest_message_age` to be configured to `true`.
      
### Example Output

```
azure_storage_queues,name=myqueue,storage_account=mystorageaccount oldest_message_age=799714900i,size=7i 1565970503000000000
azure_storage_queues,name=myemptyqueue,storage_account=mystorageaccount size=0i 1565970502000000000
```