# Azure Blob [PREVIEW]

Azure Blob output plugin exports telegraf data to Azure Blob. It caches data in memory and pushes them in regular intervals, configurable via the `flushInterval` variable.