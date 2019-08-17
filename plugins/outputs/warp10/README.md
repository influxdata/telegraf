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

### Values format

The Warp 10 output support natively number, float and boolean values. String are send as URL encoded values as well as all Influx objects. 