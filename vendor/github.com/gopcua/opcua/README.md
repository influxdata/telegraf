<p align="center">
   <img width="50%" src="https://raw.githubusercontent.com/gopcua/opcua/master/gopher.png">
</p>

<p align="center">
  Artwork by <a href="https://twitter.com/ashleymcnamara">Ashley McNamara</a> -
  Inspired by <a href="http://reneefrench.blogspot.co.uk/">Renee French</a> -
  Taken from <a href="https://gopherize.me">https://gopherize.me</a> by <a href="https://twitter.com/matryer">Mat Ryer</a>
</p>

<h1 align="center">OPCUA</h1>

A native Go implementation of the OPC/UA Binary Protocol.

You need go1.13 or higher. We test with the current and previous Go version.

[![CircleCI](https://circleci.com/gh/gopcua/opcua.svg?style=shield)](https://circleci.com/gh/gopcua/opcua)
[![GitHub](https://github.com/gopcua/opcua/workflows/gopuca/badge.svg)](https://github.com/gopcua/opcua/actions)
[![GoDoc](https://godoc.org/github.com/gopcua/opcua?status.svg)](https://godoc.org/github.com/gopcua/opcua)
[![GolangCI](https://golangci.com/badges/github.com/gopcua/opcua.svg)](https://golangci.com/r/github.com/gopcua/opcua)
[![License](https://img.shields.io/github/license/mashape/apistatus.svg)](https://github.com/gopcua/opcua/blob/master/LICENSE)
[![Version](https://img.shields.io/github/tag/gopcua/opcua.svg?color=blue&label=version)](https://github.com/gopcua/opcua/releases)

## Quickstart

```sh
# make sure you have go1.13 or higher

# install library
go get -u github.com/gopcua/opcua

# get current date and time 'ns=0;i=2258'
go run examples/datetime/datetime.go -endpoint opc.tcp://localhost:4840

# read the server version
go run examples/read/read.go -endpoint opc.tcp://localhost:4840 -node 'ns=0;i=2261'

# get the current date time using different security and authentication modes
go run examples/crypto/*.go -endpoint opc.tcp://localhost:4840 -cert path/to/cert.pem -key path/to/key.pem -sec-policy Basic256 -sec-mode SignAndEncrypt

# checkout examples/ for more examples...
```

## Disclaimer

We are still actively working on this project and the APIs will change.

We have started to tag the code to support go modules and reproducible builds
but there is still no guarantee of API stability.

However, you can safely assume that we are aiming to make the APIs as
stable as possible. :)

The [Current State](https://github.com/gopcua/opcua/wiki/Current-State) was moved
to the [Wiki](https://github.com/gopcua/opcua/wiki).

## Your Help is Appreciated

If you are looking for ways to contribute you can

 * test the high-level client against real OPC/UA servers
 * add functions to the client or tell us which functions you need for `gopcua` to be useful
 * work on the security layer, server and other components
 * and last but not least, file issues, review code and write/update documentation

Also, if the library is already useful please spread the word as a motivation.

## Authors

The [Gopcua Team](https://github.com/gopcua/opcua/graphs/contributors).

If you need to get in touch with us directly you may find us on [Keybase.io](https://keybase.io)
but try to create an issue first.

## Supported Features

The current focus is on the OPC UA Binary protocol over TCP. No other protocols are supported at this point.

| Categories     | Features                         | Supported | Notes |
|----------------|----------------------------------|-----------|-------|
| Encoding       | OPC UA Binary                    | Yes       |       |
|                | OPC UA JSON                      |           | not planned |
|                | OPC UA XML                       |           | not planned |
| Transport      | UA-TCP UA-SC UA Binary           | Yes       |       |
|                | OPC UA HTTPS                     |           | not planned |
|                | SOAP-HTTP WS-SC UA Binary        |           | not planned |
|                | SOAP-HTTP WS-SC UA XML           |           | not planned |
|                | SOAP-HTTP WS-SC UA XML-UA Binary |           | not planned |
| Encryption     | None                             | Yes       |       |
|                | Basic128Rsa15                    | Yes       |       |
|                | Basic256                         | Yes       |       |
|                | Basic256Sha256                   | Yes       |       |
| Authentication | Anonymous                        | Yes       |       |
|                | User Name Password               | Yes       |       |
|                | X509 Certificate                 | Yes       |       |

### Services

The current set of supported services is only for the high-level client.

| Service Set                 | Service                       | Supported | Notes        |
|-----------------------------|-------------------------------|-----------|--------------|
| Discovery Service Set       | FindServers                   |           |              |
|                             | FindServersOnNetwork          |           |              |
|                             | GetEndpoints                  | Yes       |              |
|                             | RegisterServer                |           |              |
|                             | RegisterServer2               |           |              |
| Secure Channel Service Set  | OpenSecureChannel             | Yes       |              |
|                             | CloseSecureChannel            | Yes       |              |
| Session Service Set         | CreateSession                 | Yes       |              |
|                             | CloseSession                  | Yes       |              |
|                             | ActivateSession               | Yes       |              |
|                             | Cancel                        |           |              |
| Node Management Service Set | AddNodes                      |           |              |
|                             | AddReferences                 |           |              |
|                             | DeleteNodes                   |           |              |
|                             | DeleteReferences              |           |              |
| View Service Set            | Browse                        | Yes       |              |
|                             | BrowseNext                    | Yes       |              |
|                             | TranslateBrowsePathsToNodeIds |           |              |
|                             | RegisterNodes                 | Yes       |              |
|                             | UnregisterNodes               | Yes       |              |
| Query Service Set           | QueryFirst                    |           |              |
|                             | QueryNext                     |           |              |
| Attribute Service Set       | Read                          | Yes       |              |
|                             | Write                         | Yes       |              |
|                             | HistoryRead                   | Yes       |              |
|                             | HistoryUpdate                 |           |              |
| Method Service Set          | Call                          | Yes       |              |
| MonitoredItems Service Set  | CreateMonitoredItems          | Yes       |              |
|                             | DeleteMonitoredItems          | Yes       |              |
|                             | ModifyMonitoredItems          |           |              |
|                             | SetMonitoringMode             |           |              |
|                             | SetTriggering                 |           |              |
| Subscription Service Set    | CreateSubscription            | Yes       |              |
|                             | ModifySubscription            |           |              |
|                             | SetPublishingMode             |           |              |
|                             | Publish                       | Yes       |              |
|                             | Republish                     |           |              |
|                             | DeleteSubscriptions           | Yes       |              |
|                             | TransferSubscriptions         |           |              |

## License

[MIT](https://github.com/gopcua/opcua/blob/master/LICENSE)
