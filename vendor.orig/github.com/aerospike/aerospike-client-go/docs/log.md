#log

Various log levels available to log from the Aerospike API.
Default is set to OFF.

```go
  import asl "github.com/aerospike/aerospike-client-go/logger"

  asl.Logger.SetLevel(asl.OFF)
```

You can set the Logger to any object that supports log.Logger interface.

## Log levels:

##### ERROR

##### WARN

##### INFO

##### DEBUG

##### OFF
