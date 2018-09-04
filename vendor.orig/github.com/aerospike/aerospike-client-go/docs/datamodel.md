# Data Model

<!--
################################################################################
binmap
################################################################################
-->
<a name="binmap"></a>

## BinMap

BinMap is a type defined as map[string]interface{} to facilitate declaring bin data.

```go
  bins := BinMap{
    "name"          : "Abu Rayhan Biruni",
    "contribution"  : "accurately calculated the radious of earth in 11th century",
    "citation"      : "https://en.wikipedia.org/wiki/History_of_geodesy#Biruni",
  }
```

<!--
################################################################################
record
################################################################################
-->
<a name="record"></a>

## Record

A record is how the data is represented and stored in the database. A record is represented as a `struct`.

Fields are:
- `Bins` — Bins and their values are represented as a BinMap (map[string]interface{})
- `Key` — Associated Key pointer
- `Node` — Database node from which the record was retrieved from.
- `Expiration` — TimeToLive of the record in seconds. Shows in how many seconds the data will be erased if not updated.
- `Generation` — Record generation (number of times the record has been updated).

The keys of the Bins are the names of the fields (bins) of a record. The values for each field can either be u/int/8,16,32,64, string, Array or Map.

Note: Arrays and Maps can contain an array or a map as a value in them. In other words, nesting of complex values is allowed.

Records are returned as a result of `Get` operations. To write back their values, one needs to pass their Bins field to the `Put` method.

Simple example of a Read, Change, Update operation:

```go
  // define a client to connect to
  client, err := NewClient("127.0.0.1", 3000)
  panicOnError(err)

  key, err := NewKey("test", "demo", "key") // key can be of any supported type
  panicOnError(err)

  // define some bins
  bins := BinMap{
    "bin1": 42, // you can pass any supported type as bin value
    "bin2": "An elephant is a mouse with an operating system",
    "bin3": []interface{}{"Go", 2009},
  }

  // write the bins
  writePolicy := NewWritePolicy(0, 0)
  err = client.Put(writePolicy, key, bins)
  panicOnError(err)

  // read it back!
  readPolicy := NewPolicy()
  rec, err := client.Get(readPolicy, key)
  panicOnError(err)

  // change data
  v := rec.Bins["bin1"].(int)
  v += 1
  rec.Bins["bin1"] = v

  // update
  err = client.Put(nil, key, rec.Bins)
```

<!--
################################################################################
recordset
################################################################################
-->
<a name="recordset"></a>

## Recordset

A recordset is the result of a scan or query operation against the database. Records are retrieved one by one from the database, and delivered to the user via `Records` channel.
To prevent too much memory use, the operation will block if the Records channel is full.
Errors are returned on `Errors` channel. If an error is of type NodeError, it will contain the Node, ResultCode and Error message of the error.

Recordsets can be closed at any time to cancel the operation.

- `Records` — The resulting records channel.
- `Errors` – The error channel.

```go
  // scan the whole cluster
  recordset, err := client.ScanAll(nil, "test", "demo")

  for res := range recordset.Results() {
    if res.Err != nil {
        // you may be able to find out on which node the error occurred
        if ne, ok := err.(NodeError); ok {
          node := ne.Node
          // do something
        }
    }

    // process record
    fmt.Println(res.Record)
  }
```

<!--
################################################################################
key
################################################################################
-->
<a name="key"></a>

## NewKey(ns, set string, key interface{})

A record is addressable via its key. A key is a struct containing:

- `ns` — The namespace of the key. Must be a String.
- `set` – The set of the key. Must be a String.
- `key` – The value of the key. Can be of any supported types.

Example:

```go
  key, err := NewKey("test", "demo", "key") // key can be of any supported type
  panicOnError(err)
```

<!--
################################################################################
bin
################################################################################
-->
<a name="bin"></a>

## NewBin(name string, value interface{}) Value

Bins are analogous to fields in relational databases.

- `name` — Bin name. Must be a String.
- `value` – The value of the key. Can be of any supported type.

Example:

```go
  bin1 := NewBin("name", "Aerospike") // string value
  bin2 := NewBin("maxTPS", 1000000) // number value
  bin3 := NewBin("notes",
    map[interface{}]interface{}{
      "age": 5,
      666: "not allowed in",
      "clients": []string{"go", "c", "java", "python", "node", "erlang"},
    }) // go wild!
```

<!--
################################################################################
statement
################################################################################
-->
<a name="statement"></a>

## NewStatement(ns string, set string, binNames ...string) *Statement

A statement indicates which records should be affected by the query. Limits are set by Filter objects. If no filters are set, the query will be a ScanAll.

- `ns`            — The namespace. Must be a String.
- `set`           – Set name. Must be a String.
- `binNames`      – (optional) name of bins which will be affected.

The following optional attributes can also be changed in the statement struct:

- `IndexName`     —  Query index name. If not set, the server will determine the index from the filter's bin name.
- `Filters`       — Optional query filters.  Currently, only one filter is allowed by the server on a secondary index lookup.

```go
  stm := NewStatement("namespace", "set", "binName")

  // Use one of the following

  // SQL Eq: select binName from ns.set where binName == 42
  stm.Addfilter(NewEqualFilter("binName", 42))

  // OR

  // SQL Eq: select binName from ns.set where binName between 0 and 100
  stm.Addfilter(NewRangeFilter("binName", 0, 100))

  // send the query with default policy
  recordset, err := client.Query(nil, stm)

  // consume recordset and check errors
  for res := recordset.Results() {
    if res.Err != nil {
      // handle error
      panic(res.Err)
    }

    // process record
    fmt.Println(res.Record)
  }
```

<!--
################################################################################
filter
################################################################################
-->
<a name="filter"></a>

## NewEqualFilter(binName string, value interface{}) *Filter

Create equality filter for query.

- `binName`       — Name of bin which is being targeted. Must be a String.
- `value`         – Value which needs to be matched. should be either integer or string

## NewRangeFilter(binName string, begin int64, end int64) *Filter

Create range filter for query. String ranges are not supported.

- `binName`       — Name of bin which is being targeted. Must be a String.
- `begin`         – Lower bound of the range. It is included in the range.
- `end`           – Upper bound of the range. It is included in the range.

Refer to statement for examples.
