# Introduction

This package describes the Aerospike Go Client API in detail.


## Usage

The aerospike Go client package is the main entry point to the client API.

```go
    import as "github.com/aerospike/aerospike-client-go"
```

Before connecting to a cluster, you must import the package.

You can then generate a client object for connecting to and operating against as cluster.

```go
  client, err := as.NewClient("127.0.0.1", 3000)
```

The application will use the client object to connect to a cluster, then perform operations such as writing and reading records.
Client object is goroutine frinedly, so you can use it in goroutines without synchronization.
It caches its connections and internal state automatically for optimal performance. These settings can also be changed.

For more details on client operations, see [Client Class](client.md).

## API Reference

- [Aerospike Go Client Library Overview](aerospike.md)
- [Client Class](client.md)
- [Object Model](datamodel.md)
- [Policy Objects](policies.md)
- [Logger Object](log.md)
