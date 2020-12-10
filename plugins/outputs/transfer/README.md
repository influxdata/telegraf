# Transfer
This plugin is used to transfer files

## Configuration
```
[[outputs.transfer]]
  flush_interval = "10s"
  remove_source = 1
  concurrency = 50
  source = '{{.source}}'
  [outputs.transfer.tagdrop]
    data_format = ['vqtcsv']
  [[outputs.transfer.entry]]
    destination = [
      'file://{{.dir_prefix}}/<<dir>>/{{.relative}}/{{.filename}}'
    ]
    error = 'file://{{.dir_prefix}}/<<dir>>/{{.relative}}/{{.filename}}'
    verbose = 0
    retries = 10
    retry_wait = "1s"
    [outputs.transfer.entry.fieldpass]
      source = [ 'file_source' ] 
```