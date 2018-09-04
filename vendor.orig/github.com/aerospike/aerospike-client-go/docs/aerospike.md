# Aerospike Package

- [Usage](#usage)
- [Structs](#structs)
  - [policies](#Policies)
  - [logger](#logger)
- [Functions](#functions)
  - [NewClient()](#client)
  - [NewKey()](#key)


<a name="usage"></a>
## Usage

The aerospike package can be imported into your project via:

```go
  import as "github.com/aerospike/aerospike-client-go"
```

<a name="structs"></a>
## Structs

<!--
################################################################################
Policies
################################################################################
-->
<a name="Policies"></a>

### Policies

Policies contain the allowed values for operation conditions for each of the [client](client.md) operations.

For details, see [Policies Object](policies.md)


<!--
################################################################################
Log
################################################################################
-->
<a name="Log"></a>

### Log

Log is a collection of the various logging levels available in Aerospike. This logging levels can be used to modify the granularity of logging from the API.
Default level is LOG_ERR.

```go
    as.Logger.SetLevel(as.INFO)
```

For details, see [Logger Object](log.md)

<a name="client"></a>

### client(host string, port int): *Client

Creates a new [client](client.md) with the provided configuration.

Parameters:

- `name` – Host name or IP to connect to.
- `port` – Host port.

Returns a new client object.

Example:

```go
  client, err := as.NewClient("127.0.0.1", 3000)
```

For detals, see [Client Class](client.md).

<!--
################################################################################
key
################################################################################
-->
<a name="key"></a>

### NewKey(ns, set string, key interface{}): *

Creates a new [key object](datamodel.md#key) with the provided arguments.

Parameters:

- `ns` – The namespace for the key.
- `set` – The set for the key.
- `key` – The value for the key.

Returns a new key.

Example:

```go
  key := as.Key("test", "demo", 123)
```

For details, see [Key Object](datamodel.md#key).

