# Client Class

The `Client` class provides operations which can be performed on an Aerospike
database cluster. In order to get an instance of the Client class, you need
to call `NewClient()`:

```go
  client, err := as.NewClient("127.0.0.1", 3000)
```

To customize a Client with a ClientPolicy:

```go
  clientPolicy := as.NewClientPolicy()
  clientPolicy.ConnectionQueueSize = 64
  clientPolicy.LimitConnectionsToQueueSize = true
  clientPolicy.Timeout = 50 * time.Millisecond

  client, err := as.NewClientWithPolicy(clientPolicy, "127.0.0.1", 3000)
```

*Notice*: Examples in the section are only intended to illuminate simple use cases without too much distraction. Always follow good coding practices in production. Always check for errors.

With a new client, you can use any of the methods specified below. You need only *ONE* client object. This object is goroutine-friendly, and pools its resources internally.

- [Methods](#methods)
  - [Add()](#add)
  - [Append()](#append)
  - [Close()](#close)
  - [Delete()](#delete)
  - [Exists()](#exists)
  - [BatchExists()](#batchexists)
  - [Get()](#get)
  - [GetHeader()](#getheader)
  - [BatchGet()](#batchget)
  - [BatchGetHeader()](#batchgetheader)
  - [IsConnected()](#isConnected)
  - [Operate()](#operate)
  - [Prepend()](#prepend)
  - [Put()](#put)
  - [PutBins()](#putbins)
  - [Touch()](#touch)
  - [ScanAll()](#scanall)
  - [ScanNode()](#scannode)
  - [CreateIndex()](#createindex)
  - [DropIndex()](#dropindex)
  - [RegisterUDF()](#registerudf)
  - [RegisterUDFFromFile()](#registerudffromfile)
  - [Execute()](#execute)
  - [ExecuteUDF()](#executeudf)
  - [Query()](#query)


<a name="methods"></a>
## Methods

<!--
################################################################################
add()
################################################################################
-->
<a name="add"></a>

### Add(policy *WritePolicy, key *Key, bins BinMap) error

Using the provided key, adds values to the mentioned bins.
Bin value types should be of type `integer` for the command to have any effect.

Parameters:

- `policy`      – (optional) A [Write Policy object](policies.md#WritePolicy) to use for this operation.
                Pass `nil` for default values.
- `key`         – A [Key object](datamodel.md#key), used to locate the record in the cluster.
- `bins`        – A [BinMap](datamodel.md#binmap) used for specifying the fields and value.

Example:
```go
  key := NewKey("test", "demo", 123)

  bins = BinMap {
    "e": 2,
    "pi": 3,
  }

  err := client.Add(nil, key, bins)
```

<!--
################################################################################
append()
################################################################################
-->
<a name="append"></a>

### Append(policy *WritePolicy, key *Key, bins BinMap) error

Using the provided key, appends provided values to the mentioned bins.
Bin value types should be of type `string` or `[]byte` for the command to have any effect.

Parameters:

- `policy`      – (optional) A [Write Policy object](policies.md#WritePolicy) to use for this operation.
                Pass `nil` for default values.
- `key`         – A [Key object](datamodel.md#key), used to locate the record in the cluster.
- `bins`        – A [BinMap](datamodel.md#binmap) used for specifying the fields and value.

Example:
```go
  key := NewKey("test", "demo", 123)

  bins = BinMap {
    "story": ", and lived happily ever after...",
  }

  err := client.Append(nil, key, bins)
```

<!--
################################################################################
close()
################################################################################
-->
<a name="close"></a>

### Close()

Closes the client connection to the cluster.

Example:
```go
  client.Close()
```

<!--
################################################################################
remove()
################################################################################
-->
<a name="delete"></a>

### Delete(policy *WritePoicy, key *Key) (existed bool, err error)

Removes a record with the specified key from the database cluster.

Parameters:

- `policy`      – (optional) The [delete Policy object](policies.md#RemovePolicy) to use for this operation.
- `key`         – A [Key object](datamodel.md#key) used for locating the record to be removed.

returned values:

- `existed`         – Boolean value that indicates if the Key existed.

Example:
```go
  key := NewKey("test", "demo", 123)

  if existed, err := client.Delete(nil, key); existed {
    // do something
  }
```

<!--
################################################################################
exists()
################################################################################
-->
<a name="exists"></a>

### Exists(policy *BasePolicy, key *Key) (bool, error)

Using the key provided, checks for the existence of a record in the database cluster .

Parameters:

- `policy`      – (optional) The [BasePolicy object](policies.md#BasePolicy) to use for this operation.
                  Pass `nil` for default values.
- `key`         – A [Key object](datamodel.md#key), used to locate the record in the cluster.

Example:

```go
  key := NewKey("test", "demo", 123)

  if exists, err := client.Exists(nil, key) {
    // do something
  }
```

<!--
################################################################################
batchexists()
################################################################################
-->
<a name="batchexists"></a>

### BatchExists(policy *BasePolicy, keys []*Key) ([]bool, error)

Using the keys provided, checks for the existence of records in the database cluster in one request.

Parameters:

- `policy`      – (optional) The [BasePolicy object](policies.md#BasePolicy) to use for this operation.
                  Pass `nil` for default values.
- `keys`         – A [Key array](datamodel.md#key), used to locate the records in the cluster.

Example:

```go
  key1 := NewKey("test", "demo", 123)
  key2 := NewKey("test", "demo", 42)

  existanceArray, err := client.Exists(nil, []*Key{key1, key2}) {
    // do something
  }
```

<!--
################################################################################
get()
################################################################################
-->
<a name="get"></a>

### Get(policy *BasePolicy, key *Key, bins ...string) (*Record, error)

Using the key provided, reads a record from the database cluster .

Parameters:

- `policy`      – (optional) The [BasePolicy object](policies.md#BasePolicy) to use for this operation.
                  Pass `nil` for default values.
- `key`         – A [Key object](datamodel.md#key), used to locate the record in the cluster.
- `bins`        – (optional) Bins to retrieve. Will retrieve all bins if not provided.

Example:

```go
  key := NewKey("test", "demo", 123)

  rec, err := client.Get(nil, key) // reads all the bins
```

<!--
################################################################################
getheader()
################################################################################
-->
<a name="getheader"></a>

### GetHeader(policy *BasePolicy, key *Key) (*Record, error)

Using the key provided, reads *ONLY* record metadata from the database cluster. Record metadata includes record generation and Expiration (TTL from the moment of retrieval, in seconds)

```record.Bins``` will always be empty in resulting ```record```.

Parameters:

- `policy`      – (optional) The [BasePolicy object](policies.md#BasePolicy) to use for this operation.
                  Pass `nil` for default values.
- `key`         – A [Key object](datamodel.md#key), used to locate the record in the cluster.

Example:

```go
  key := NewKey("test", "demo", 123)

  rec, err := client.GetHeader(nil, key) // No bins will be retrieved
```

<!--
################################################################################
batchget()
################################################################################
-->
<a name="batchget"></a>

### BatchGet(policy *BasePolicy, keys *[]Key, bins ...string) ([]*Record, error)

Using the keys provided, reads all relevant records from the database cluster in a single request.

Parameters:

- `policy`      – (optional) The [BasePolicy object](policies.md#BasePolicy) to use for this operation.
                  Pass `nil` for default values.
- `keys`         – A [Key array](datamodel.md#key), used to locate the record in the cluster.
- `bins`        – (optional) Bins to retrieve. Will retrieve all bins if not provided.

Example:

```go
  key1 := NewKey("test", "demo", 123)
  key2 := NewKey("test", "demo", 42)

  recs, err := client.BatchGet(nil, []*Key{key1, key2}) // reads all the bins
```

<!--
################################################################################
batchgetheader()
################################################################################
-->
<a name="batchgetheader"></a>

### BatchGetHeader(policy *BasePolicy, keys *[]Key) ([]*Record, error)

Using the keys provided, reads all relevant record metadata from the database cluster in a single request.

```record.Bins``` will always be empty in resulting ```record```.

Parameters:

- `policy`      – (optional) The [BasePolicy object](policies.md#BasePolicy) to use for this operation.
                  Pass `nil` for default values.
- `keys`         – A [Key array](datamodel.md#key), used to locate the record in the cluster.

Example:

```go
  key1 := NewKey("test", "demo", 123)
  key2 := NewKey("test", "demo", 42)

  recs, err := client.BatchGetHeader(nil, []*Key{key1, key2}) // reads all the bins
```
<!--
################################################################################
idConnected()
################################################################################
-->
<a name="isConnected"></a>

### IsConnected() bool

Checks if the client is connected to the cluster.

<!--
################################################################################
prepend()
################################################################################
-->
<a name="prepend"></a>

### Prepend(policy *WritePolicy, key *Key, bins BinMap) error

Using the provided key, prepends provided values to the mentioned bins.
Bin value types should be of type `string` or `[]byte` for the command to have any effect.

Parameters:

- `policy`      – (optional) A [Write Policy object](policies.md#WritePolicy) to use for this operation.
                Pass `nil` for default values.
- `key`         – A [Key object](datamodel.md#key), used to locate the record in the cluster.
- `bins`        – A [BinMap](datamodel.md#binmap) used for specifying the fields and value.

Example:
```go
  key := NewKey("test", "demo", 123)

  bins = BinMap {
    "story": "Long ago, in a galaxy far far away, ",
  }

  err := client.Prepend(nil, key, bins)
```

<!--
################################################################################
put()
################################################################################
-->
<a name="put"></a>

### Put(policy *WritePolicy, key *Key, bins BinMap) error

Writes a record to the database cluster. If the record exists, it modifies the record with bins provided.
To remove a bin, set its value to `nil`.

#### Node: Under the hood, Put converts BinMap to []Bins and uses ```PutBins```. Use PutBins to avoid unnecessary memory allocation and iteration.

Parameters:

- `policy`      – (optional) A [Write Policy object](policies.md#WritePolicy) to use for this operation.
                Pass `nil` for default values.
- `key`         – A [Key object](datamodel.md#key), used to locate the record in the cluster.
- `bins`        – A [BinMap map](datamodel.md#binmap) used for specifying the fields to store.

Example:
```go
  key := NewKey("test", "demo", 123)

  bins := BinMap {
    "a": "Lack of skill dictates economy of style.",
    "b": 123,
    "c": []int{1, 2, 3},
    "d": map[string]interface{}{"a": 42, "b": "An elephant is mouse with an operating system."},
  }

  err := client.Put(nil, key, bins)
```

<!--
################################################################################
putbins()
################################################################################
-->
<a name="putbins"></a>

### PutBins(policy *WritePolicy, key *Key, bins ...*Bin) error

Writes a record to the database cluster. If the record exists, it modifies the record with bins provided.
To remove a bin, set its value to `nil`.

Parameters:

- `policy`      – (optional) A [Write Policy object](policies.md#WritePolicy) to use for this operation.
                Pass `nil` for default values.
- `key`         – A [Key object](datamodel.md#key), used to locate the record in the cluster.
- `bins`        – A [Bin array](datamodel.md#bin) used for specifying the fields to store.

Example:
```go
  key := NewKey("test", "demo", 123)

  bin1 := NewBin("a", "Lack of skill dictates economy of style.")
  bin2 := NewBin("b", 123)
  bin3 := NewBin("c", []int{1, 2, 3})
  bin4 := NewBin("d", map[string]interface{}{"a": 42, "b": "An elephant is mouse with an operating system."})

  err := client.PutBins(nil, key, bin1, bin2, bin3, bin4)
```

<!--
################################################################################
touch()
################################################################################
-->
<a name="touch"></a>

### Touch(policy *WritePolicy, key *Key) error

Create record if it does not already exist.
If the record exists, the record's time to expiration will be reset to the policy's expiration.

Parameters:

- `policy`      – (optional) A [Write Policy object](policies.md#WritePolicy) to use for this operation.
                Pass `nil` for default values.
- `key`         – A [Key object](datamodel.md#key), used to locate the record in the cluster.

Example:
```go
  key := NewKey("test", "demo", 123)

  err := client.Touch(NewWritePolicy(0, 5), key)
```

<!--
################################################################################
scanall()
################################################################################
-->
<a name="scanall"></a>

### ScanAll(policy *ScanPolicy, namespace string, setName string, binNames ...string) (*Recordset, error)

Performs a full Scan on all nodes in the cluster, and returns the results in a [Recordset object](datamodel.md#recordset)


Parameters:

- `policy`      – (optional) A [Scan Policy object](policies.md#ScanPolicy) to use for this operation.
                Pass `nil` for default values.
- `namespace`         – Namespace to perform the scan on.
- `setName`         – Name of the Set to perform the scan on.
- `binNames`         – Name of bins to retrieve. If not passed, all bins will be retrieved.

Refer to [Recordset object](datamodel.md#recordset) documentation for details on how to retrieve the data.

Example:
```go
  // scan the whole cluster
  recordset, err := client.ScanAll(nil, "test", "demo")

  for res := range recordset.Results() {
    if res.Err != nil {
      // handle error; or close the recordset and break
    }
    
  // process record
  fmt.Println(res.Record)
  }
```

<!--
################################################################################
scannode()
################################################################################
-->
<a name="scannode"></a>

### ScanNode(policy *ScanPolicy, node *Node, namespace string, setName string, binNames ...string) (*Recordset, error)

Performs a full Scan *on a specific node* in the cluster, and returns the results in a [Recordset object](datamodel.md#recordset)

It works the same as ScanAll() method.

<!--
################################################################################
createindex()
################################################################################
-->
<a name="createindex"></a>

### CreateIndex(policy *WritePolicy, namespace string, setName string, indexName string, binName string, indexType IndexType) (*IndexTask, error)

Creates a secondary index. IndexTask will return a IndexTask object which can be used to determine if the operation is completed.

Parameters:

- `policy`      – (optional) A [Write Policy object](policies.md#WritePolicy) to use for this operation.
                Pass `nil` for default values.
- `namespace`         – Namespace
- `setName`         – Name of the Set
- `indexName`         – Name of index
- `binName`         – Bin name to create the index on
- `indexType`         – STRING or NUMERIC

Example:

```go
  idxTask, err := client.CreateIndex(nil, "test", "demo", "indexName", "binName", NUMERIC)
  panicOnErr(err)

  // wait until index is created.
  // OnComplete() channel will return nil on success and an error on errors
  err = <- idxTask.OnComplete()
  if err != nil {
    panic(err)
  }
```

<!--
################################################################################
dropindex()
################################################################################
-->
<a name="dropindex"></a>
### DropIndex(  policy *WritePolicy,  namespace string,  setName string,  indexName string) error

Drops an index.

Parameters:

- `policy`      – (optional) A [Write Policy object](policies.md#WritePolicy) to use for this operation.
                Pass `nil` for default values.
- `namespace`         – Namespace
- `setName`           – Name of the Set.
- `indexName`         – Name of index

```go
  err := client.DropIndex(nil, "test", "demo", "indexName")
```

<!--
################################################################################
registerudf()
################################################################################
-->
<a name="registerudf"></a>

### RegisterUDF(policy *WritePolicy, udfBody []byte, serverPath string, language Language) (*RegisterTask, error)

Registers the given UDF on the server.

Parameters:

- `policy`      – (optional) A [Write Policy object](policies.md#WritePolicy) to use for this operation.
                Pass `nil` for default values.
- `udfBody`     – UDF source code
- `serverPath`  – Path on which the UDF should be put on the server-side
- `language`    – Only 'LUA' is currently supported


Example:

```go
  const udfBody = `function testFunc1(rec)
     local ret = map()                     -- Initialize the return value (a map)

     local x = rec['bin1']               -- Get the value from record bin named "bin1"

     rec['bin2'] = (x / 2)               -- Set the value in record bin named "bin2"

     aerospike:update(rec)                -- Update the main record

     ret['status'] = 'OK'                   -- Populate the return status
     return ret                             -- Return the Return value and/or status
  end`

  regTask, err := client.RegisterUDF(nil, []byte(udfBody), "udf1.lua", LUA)
  panicOnErr(err)

  // wait until UDF is created
  err = <-regTask.OnComplete()
  if err != nil {
    panic(err)
  }
```

<!--
################################################################################
registerudffromfile()
################################################################################
-->
<a name="registerudffromfile"></a>

### RegisterUDFFromFile(policy *WritePolicy, clientPath string, serverPath string, language Language) (*RegisterTask, error)

Read the UDF source code from a file and registers it on the server.

Parameters:

- `policy`      – (optional) A [Write Policy object](policies.md#WritePolicy) to use for this operation.
                Pass `nil` for default values.
- `clientPath`  – full file path for UDF source code
- `serverPath`  – Path on which the UDF should be put on the server-side
- `language`    – Only 'LUA' is currently supported


Example:

```go
  regTask, err := client.RegisterUDFFromFile(nil, "/path/udf.lua", "udf1.lua", LUA)
  panicOnErr(err)

  // wait until UDF is created
  err = <- regTask.OnComplete()
  if err != nil {
    panic(err)
  }
```

<!--
################################################################################
execute()
################################################################################
-->
<a name="execute"></a>

### Execute(policy *WritePolicy, key *Key, packageName string, functionName string, args ...Value) (interface{}, error)

Executes a UDF on a record with the given key, and returns the results.

Parameters:

- `policy`       – (optional) A [Write Policy object](policies.md#WritePolicy) to use for this operation.
                Pass `nil` for default values.
- `packageName`  – server path to the UDF
- `functionName` – UDF name
- `args`         – (optional) UDF arguments

Example:

Considering the UDF registered in RegisterUDF example above:

```go
    res, err := client.Execute(nil, key, "udf1", "testFunc1")

    // res will be a: map[interface{}]interface{}{"status": "OK"}
```
<!--
################################################################################
executeudf()
################################################################################
-->
<a name="executeudf"></a>

### ExecuteUDF(policy *QueryPolicy,  statement *Statement,  packageName string,  functionName string,  functionArgs ...Value) (*ExecuteTask, error)

Executes a UDF on all records which satisfy filters set in the statement. If there are filters, it will run on all records in the database.

Parameters:

- `policy`       – (optional) A [Query Policy object](policies.md#QueryPolicy) to use for this operation.
                Pass `nil` for default values.
- `statement`    – [Statement object](datamodel.md#statement) to narrow down records.
- `packageName`  – server path to the UDF
- `functionName` – UDF name
- `functionArgs` – (optional) UDF arguments

Example:

Considering the UDF registered in RegisterUDF example above:

```go
  statement := NewStatement("namespace", "set")
  exTask, err := client.ExecuteUDF(nil, statement, "udf1", "testFunc1")
  panicOnErr(err)

  // wait until UDF is run on all records
  err = <- exTask.OnComplete()
  if err != nil {
    panic(err)
  }
```

<!--
################################################################################
query()
################################################################################
-->
<a name="query"></a>

### Query(policy *QueryPolicy, statement *Statement) (*Recordset, error)

Performs a query on the cluster, and returns the results in a [Recordset object](datamodel.md#recordset)


Parameters:

- `policy`       – (optional) A [Query Policy object](policies.md#QueryPolicy) to use for this operation.
                Pass `nil` for default values.
- `statement`    – [Statement object](datamodel.md#statement) to narrow down records.

Refer to [Recordset object](datamodel.md#recordset) documentation for details on how to retrieve the data.


Example:

```go
  stm := NewStatement("namespace", "set")
  stm.Addfilter(NewRangeFilter("binName", value1, value2))

  recordset, err := client.Query(nil, stm)

  // consume recordset and check errors
  for res := recordset.Results() {
    if res.Err != nil {
      // handle error, or close the recordset and break
    }

    // process record
    fmt.Println(res.Record)
  }
```
