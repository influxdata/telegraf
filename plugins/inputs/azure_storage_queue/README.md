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
```

### Measurements & Fields:

- azure_storage_queues:
  - size

### Tags:

- azure_storage_queues:
  - name
  - storage_account