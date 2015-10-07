Eclipse Paho MQTT Go client
===========================


This repository contains the source code for the [Eclipse Paho](http://eclipse.org/paho) MQTT Go client library. 

This code builds a library which enable applications to connect to an [MQTT](http://mqtt.org) broker to publish messages, and to subscribe to topics and receive published messages.

This library supports a fully asynchronous mode of operation.


Installation and Build
----------------------

This client is designed to work with the standard Go tools, so installation is as easy as:

```
go get git.eclipse.org/gitroot/paho/org.eclipse.paho.mqtt.golang.git
```

The client depends on Google's [websockets](http://godoc.org/code.google.com/p/go.net/websocket) package, 
also easily installed with the command:

```
go get code.google.com/p/go.net/websocket
```


Usage and API
-------------

Detailed API documentation is available by using to godoc tool, or can be browsed online
using the [godoc.org](http://godoc.org/git.eclipse.org/gitroot/paho/org.eclipse.paho.mqtt.golang.git) service.

Make use of the library by importing it in your Go client source code. For example,
```
import MQTT "git.eclipse.org/gitroot/paho/org.eclipse.paho.mqtt.golang.git"
```

Samples are available in the `/samples` directory for reference.


Runtime tracing
---------------

Tracing is enabled by assigning logs (from the Go log package) to the logging endpoints, ERROR, CRITICAL, WARN and DEBUG


Reporting bugs
--------------

Please report bugs under the "MQTT-Go" Component in [Eclipse Bugzilla](http://bugs.eclipse.org/bugs/) for the Paho Technology project. This is a very new library as of Q1 2014, so there are sure to be bugs.


More information
----------------

Discussion of the Paho clients takes place on the [Eclipse paho-dev mailing list](https://dev.eclipse.org/mailman/listinfo/paho-dev).

General questions about the MQTT protocol are discussed in the [MQTT Google Group](https://groups.google.com/forum/?hl=en-US&fromgroups#!forum/mqtt).

There is much more information available via the [MQTT community site](http://mqtt.org).
