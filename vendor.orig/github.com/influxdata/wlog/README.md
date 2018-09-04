# wlog
Simple log level based Go logger.
Provides an io.Writer that filters log messages based on a log level prefix.
Valid log levels are: DEBUG, INFO, WARN, ERROR, OFF.
Log messages need to begin with a L! where L is one of D, I, W, or E.


## Usage

Create a *log.Logger via wlog.New:

```go
package main

import (
    "log"
    "os"

    "github.com/influxdata/wlog"
)

func main() {
    var logger *log.Logger
    logger = wlog.New(os.Stderr, "prefix", log.LstdFlags)
    logger.Println("I! initialized logger")
}
```

Create a *log.Logger explicitly using wlog.Writer:

```go
package main

import (
    "log"
    "os"

    "github.com/influxdata/wlog"
)

func main() {
    var logger *log.Logger
    logger = log.New(wlog.NewWriter(os.Stderr), "prefix", log.LstdFlags)
    logger.Println("I! initialized logger")
}
```

Prefix log messages with a log level char and the `!` delimiter.

```go
logger.Println("D! this is a debug log")
logger.Println("I! this is an info log")
logger.Println("W! this is a warn log")
logger.Println("E! this is an error log")
```


The log level can be changed via the SetLevel or the SetLevelFromName functions.


```go
package main

import (
    "log"
    "os"

    "github.com/influxdata/wlog"
)

func main() {
    var logger *log.Logger
    logger = wlog.New(os.Stderr, "prefix", log.LstdFlags)
    wlog.SetLevel(wlog.DEBUG)
    logger.Println("D! initialized logger")
    wlog.SetLevelFromName("INFO")
    logger.Println("D! this message will be dropped")
    logger.Println("I! this message will be printed")
}
```

