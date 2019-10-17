# Azure Blob [Preview version]

Azure Blob output plugin exports zipped telegraf data to Azure Blob. It caches data in memory and pushes them in regular intervals, configurable via the `flushInterval` variable.

## Authentication

You can authenticate to Azure Blob Storage via an account name/account key combination or via a SAS Url, which should have appropriate permissions.

## Configuration

You can configure the flush interval (seconds) as well as the Blob Container that the zip files will be created. The files have the format `startTime-endTime-machineName`. Times are in UTC and the machineName can be set via configuration, since telegraf could be running in a container but you'd like the host's hostname to be there.

```toml
[[outputs.azure_blob]]
  ## You need to have either an account/account key combination or a SAS URL
  ## Azure Blob account
  # blobAccount = "myblobaccount"
  ## Azure Blob account key
  # blobAccountKey = "myblobaccountkey"
  ## Azure Blob SAS URL
  # blobAccountSasURL = "YOUR_SAS_URL"
  ## Azure Blob container name
  # blobContainerName = "telegrafcontainer"
  ## Flush interval in seconds
  # flushInterval = 300
  ## Machine name that is sending the data
  # machineName = "myhostname"
  ## Data format to output.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
  # data_format = "json"
```