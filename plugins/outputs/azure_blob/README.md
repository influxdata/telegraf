# Azure Blob [Preview version]

Azure Blob output plugin exports zipped telegraf data to Azure Blob. It caches data in memory and pushes them in regular intervals, configurable via the `flushInterval` variable.

## Authentication

You can authenticate to Azure Blob Storage via an account name/account key combination or via a SAS url. This (pre-authenticated) url should include the name of the Blob Container and should have appropriate permissions (create and write), so it should be something like this:

_https://ACCOUNTNAME.blob.core.windows.net/CONTAINERNAME?sv=2018-03-28&sr=c&sig=REDACTED_

If you use account name/account key to login, you can optionally provide the Blob Container name (in the `blobContainerName` variable).

## Configuration

You can configure the flush interval (seconds) that the zip files will be created in Blob Storage. These files have the format `machineName-endTime-startTime`. Times are in UTC whereas the machineName can be set via configuration. This is because telegraf can be running in a container but you'd like the host's hostname to be there.

```toml
[[outputs.azure_blob]]
  ## You need to have either an accountName/accountKey combination or a SAS URL
  ## SAS URL should contain the Blob Container Name and have appropriate permissions (create and write)
  ## Azure Blob account
  # blobAccount = "myblobaccount"
  ## Azure Blob account key
  # blobAccountKey = "myblobaccountkey"
  ## Azure Blob container name. Used only when authenticating via accountName. If omitted, "metrics" is used
  # blobContainerName = "telegrafcontainer"
  ## Azure Blob Container SAS URL
  # blobAccountSasURL = "YOUR_SAS_URL"
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