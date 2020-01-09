# README #

Telegraph plugin to push metrics on Warp10

### Telegraph output for Warp10 ###

Execute a post http on Warp10 at every flush time configured in telegraph in order to push the metrics collected

### Config ###

Add following instruction in the config file (Output part)

```
[[outputs.warp10]]
warpUrl = "http://localhost:4242"
token = "token"
prefix = "telegraf."
timeout = "15s" 
```

To get more details on Warp 10 errors occuring when pushing data with Telegraf, you can optionaly set:

```
printErrorBody = true   ## To print the full body of the HTTP Post instead of the request status
maxStringErrorSize = 700  ## To update the maximal string size of the Warp 10 error body. By default it's set to 512.
```

### Values format

The Warp 10 output support natively number, float and boolean values. String are send as URL encoded values as well as all Influx objects.