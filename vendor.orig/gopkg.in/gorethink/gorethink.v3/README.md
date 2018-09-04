# GoRethink - RethinkDB Driver for Go

[![GitHub tag](https://img.shields.io/github/tag/GoRethink/gorethink.svg?style=flat)](https://github.com/GoRethink/gorethink/releases)
[![GoDoc](https://godoc.org/github.com/GoRethink/gorethink?status.svg)](https://godoc.org/github.com/GoRethink/gorethink)
[![Build status](https://travis-ci.org/GoRethink/gorethink.svg?branch=master)](https://travis-ci.org/GoRethink/gorethink)
<!-- [![No Maintenance Intended](http://unmaintained.tech/badge.svg)](http://unmaintained.tech/) -->

[Go](http://golang.org/) driver for [RethinkDB](http://www.rethinkdb.com/)

![GoRethink Logo](https://raw.github.com/wiki/gorethink/gorethink/gopher-and-thinker-s.png "Golang Gopher and RethinkDB Thinker")

Current version: v3.0.5 (RethinkDB v2.3)

<!-- This project is no longer maintained, for more information see the [v3.0.0 release](https://github.com/gorethink/gorethink/releases/tag/v3.0.0)-->

Please note that this version of the driver only supports versions of RethinkDB using the v0.4 protocol (any versions of the driver older than RethinkDB 2.0 will not work).

If you need any help you can find me on the [RethinkDB slack](http://slack.rethinkdb.com/) in the #gorethink channel.

## Installation

```
go get gopkg.in/gorethink/gorethink.v3
```

Replace `v3` with `v2` or `v1` to use previous versions.

## Example

[embedmd]:# (example_test.go go)
```go
package gorethink_test

import (
	"fmt"
	"log"

	r "gopkg.in/gorethink/gorethink.v3"
)

func Example() {
	session, err := r.Connect(r.ConnectOpts{
		Address: url,
	})
	if err != nil {
		log.Fatalln(err)
	}

	res, err := r.Expr("Hello World").Run(session)
	if err != nil {
		log.Fatalln(err)
	}

	var response string
	err = res.One(&response)
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println(response)

	// Output:
	// Hello World
}
```

## Connection

### Basic Connection

Setting up a basic connection with RethinkDB is simple:

[embedmd]:# (example_connect_test.go go /func ExampleConnect\(\) {/ /^}/)
```go
func ExampleConnect() {
	var err error

	session, err = r.Connect(r.ConnectOpts{
		Address: url,
	})
	if err != nil {
		log.Fatalln(err.Error())
	}
}
```

See the [documentation](http://godoc.org/github.com/gorethink/gorethink#Connect) for a list of supported arguments to Connect().

### Connection Pool

The driver uses a connection pool at all times, by default it creates and frees connections automatically. It's safe for concurrent use by multiple goroutines.

To configure the connection pool `InitialCap`, `MaxOpen` and `Timeout` can be specified during connection. If you wish to change the value of `InitialCap` or `MaxOpen` during runtime then the functions `SetInitialPoolCap` and `SetMaxOpenConns` can be used.

[embedmd]:# (example_connect_test.go go /func ExampleConnect_connectionPool\(\) {/ /^}/)
```go
func ExampleConnect_connectionPool() {
	var err error

	session, err = r.Connect(r.ConnectOpts{
		Address:    url,
		InitialCap: 10,
		MaxOpen:    10,
	})
	if err != nil {
		log.Fatalln(err.Error())
	}
}
```

### Connect to a cluster

To connect to a RethinkDB cluster which has multiple nodes you can use the following syntax. When connecting to a cluster with multiple nodes queries will be distributed between these nodes.

[embedmd]:# (example_connect_test.go go /func ExampleConnect_cluster\(\) {/ /^}/)
```go
func ExampleConnect_cluster() {
	var err error

	session, err = r.Connect(r.ConnectOpts{
		Addresses: []string{url},
		//  Addresses: []string{url1, url2, url3, ...},
	})
	if err != nil {
		log.Fatalln(err.Error())
	}
}
```

When `DiscoverHosts` is true any nodes are added to the cluster after the initial connection then the new node will be added to the pool of available nodes used by GoRethink. Unfortunately the canonical address of each server in the cluster **MUST** be set as otherwise clients will try to connect to the database nodes locally. For more information about how to set a RethinkDB servers canonical address set this page http://www.rethinkdb.com/docs/config-file/.

## User Authentication

To login with a username and password you should first create a user, this can be done by writing to the `users` system table and then grant that user access to any tables or databases they need access to. This queries can also be executed in the RethinkDB admin console.

```go
err := r.DB("rethinkdb").Table("users").Insert(map[string]string{
    "id": "john",
    "password": "p455w0rd",
}).Exec(session)
...
err = r.DB("blog").Table("posts").Grant("john", map[string]bool{
    "read": true,
    "write": true,
}).Exec(session)
...
```

Finally the username and password should be passed to `Connect` when creating your session, for example:

```go
session, err := r.Connect(r.ConnectOpts{
    Address: "localhost:28015",
    Database: "blog",
    Username: "john",
    Password: "p455w0rd",
})
```

Please note that `DiscoverHosts` will not work with user authentication at this time due to the fact that RethinkDB restricts access to the required system tables.

## Query Functions

This library is based on the official drivers so the code on the [API](http://www.rethinkdb.com/api/) page should require very few changes to work.

To view full documentation for the query functions check the [API reference](https://github.com/gorethink/gorethink/wiki/Go-ReQL-command-reference) or [GoDoc](http://godoc.org/github.com/gorethink/gorethink#Term)

Slice Expr Example
```go
r.Expr([]interface{}{1, 2, 3, 4, 5}).Run(session)
```
Map Expr Example
```go
r.Expr(map[string]interface{}{"a": 1, "b": 2, "c": 3}).Run(session)
```
Get Example
```go
r.DB("database").Table("table").Get("GUID").Run(session)
```
Map Example (Func)
```go
r.Expr([]interface{}{1, 2, 3, 4, 5}).Map(func (row Term) interface{} {
    return row.Add(1)
}).Run(session)
```
Map Example (Implicit)
```go
r.Expr([]interface{}{1, 2, 3, 4, 5}).Map(r.Row.Add(1)).Run(session)
```
Between (Optional Args) Example
```go
r.DB("database").Table("table").Between(1, 10, r.BetweenOpts{
    Index: "num",
    RightBound: "closed",
}).Run(session)
```

For any queries which use callbacks the function signature is important as your function needs to be a valid GoRethink callback, you can see an example of this in the map example above. The simplified explanation is that all arguments must be of type `r.Term`, this is because of how the query is sent to the database (your callback is not actually executed in your Go application but encoded as JSON and executed by RethinkDB). The return argument can be anything you want it to be (as long as it is a valid return value for the current query) so it usually makes sense to return `interface{}`. Here is an example of a callback for the conflict callback of an insert operation:

```go
r.Table("test").Insert(doc, r.InsertOpts{
    Conflict: func(id, oldDoc, newDoc r.Term) interface{} {
        return newDoc.Merge(map[string]interface{}{
            "count": oldDoc.Add(newDoc.Field("count")),
        })
    },
})
```

### Optional Arguments

As shown above in the Between example optional arguments are passed to the function as a struct. Each function that has optional arguments as a related struct. This structs are named in the format FunctionNameOpts, for example BetweenOpts is the related struct for Between.

#### Cancelling queries

For query cancellation use `Context` argument at `RunOpts`. If `Context` is `nil` and `ReadTimeout` or `WriteTimeout` is not 0 from `ConnectionOpts`, `Context` will be formed by summation of these timeouts.

For unlimited timeouts for `Changes()` pass `context.Background()`.

## Results

Different result types are returned depending on what function is used to execute the query.

- `Run` returns a cursor which can be used to view all rows returned.
- `RunWrite` returns a WriteResponse and should be used for queries such as Insert, Update, etc...
- `Exec` sends a query to the server and closes the connection immediately after reading the response from the database. If you do not wish to wait for the response then you can set the `NoReply` flag.

Example:

```go
res, err := r.DB("database").Table("tablename").Get(key).Run(session)
if err != nil {
    // error
}
defer res.Close() // Always ensure you close the cursor to ensure connections are not leaked
```

Cursors have a number of methods available for accessing the query results

- `Next` retrieves the next document from the result set, blocking if necessary.
- `All` retrieves all documents from the result set into the provided slice.
- `One` retrieves the first document from the result set.

Examples:

```go
var row interface{}
for res.Next(&row) {
    // Do something with row
}
if res.Err() != nil {
    // error
}
```

```go
var rows []interface{}
err := res.All(&rows)
if err != nil {
    // error
}
```

```go
var row interface{}
err := res.One(&row)
if err == r.ErrEmptyResult {
    // row not found
}
if err != nil {
    // error
}
```

## Encoding/Decoding
When passing structs to Expr(And functions that use Expr such as Insert, Update) the structs are encoded into a map before being sent to the server. Each exported field is added to the map unless

  - the field's tag is "-", or
  - the field is empty and its tag specifies the "omitempty" option.

Each fields default name in the map is the field name but can be specified in the struct field's tag value. The "gorethink" key in
the struct field's tag value is the key name, followed by an optional comma
and options. Examples:

```go
// Field is ignored by this package.
Field int `gorethink:"-"`
// Field appears as key "myName".
Field int `gorethink:"myName"`
// Field appears as key "myName" and
// the field is omitted from the object if its value is empty,
// as defined above.
Field int `gorethink:"myName,omitempty"`
// Field appears as key "Field" (the default), but
// the field is skipped if empty.
// Note the leading comma.
Field int `gorethink:",omitempty"`
// When the tag name includes an index expression
// a compound field is created
Field1 int `gorethink:"myName[0]"`
Field2 int `gorethink:"myName[1]"`
```

**NOTE:** It is strongly recommended that struct tags are used to explicitly define the mapping between your Go type and how the data is stored by RethinkDB. This is especially important when using an `Id` field as by default RethinkDB will create a field named `id` as the primary key (note that the RethinkDB field is lowercase but the Go version starts with a capital letter).

When encoding maps with non-string keys the key values are automatically converted to strings where possible, however it is recommended that you use strings where possible (for example `map[string]T`).

If you wish to use the `json` tags for GoRethink then you can call `SetTags("gorethink", "json")` when starting your program, this will cause GoRethink to check for `json` tags after checking for `gorethink` tags. By default this feature is disabled. This function will also let you support any other tags, the driver will check for tags in the same order as the parameters.

### Pseudo-types

RethinkDB contains some special types which can be used to store special value types, currently supports are binary values, times and geometry data types. GoRethink supports these data types natively however there are some gotchas:
 - Time types: To store times in RethinkDB with GoRethink you must pass a `time.Time` value to your query, due to the way Go works type aliasing or embedding is not support here
 - Binary types: To store binary data pass a byte slice (`[]byte`) to your query
 - Geometry types: As Go does not include any built-in data structures for storing geometry data GoRethink includes its own in the `github.com/gorethink/gorethink/types` package, Any of the types (`Geometry`, `Point`, `Line` and `Lines`) can be passed to a query to create a RethinkDB geometry type.

### Compound Keys

RethinkDB unfortunately does not support compound primary keys using multiple fields however it does support compound keys using an array of values. For example if you wanted to create a compound key for a book where the key contained the author ID and book name then the ID might look like this `["author_id", "book name"]`. Luckily GoRethink allows you to easily manage these keys while keeping the fields separate in your structs. For example:

```go
type Book struct {
  AuthorID string `gorethink:"id[0]"`
  Name     string `gorethink:"id[1]"`
}
// Creates the following document in RethinkDB
{"id": [AUTHORID, NAME]}
```

### References

Sometimes you may want to use a Go struct that references a document in another table, instead of creating a new struct which is just used when writing to RethinkDB you can annotate your struct with the reference tag option. This will tell GoRethink that when encoding your data it should "pluck" the ID field from the nested document and use that instead.

This is all quite complicated so hopefully this example should help. First lets assume you have two types `Author` and `Book` and you want to insert a new book into your database however you dont want to include the entire author struct in the books table. As you can see the `Author` field in the `Book` struct has some extra tags, firstly we have added the `reference` tag option which tells GoRethink to pluck a field from the `Author` struct instead of inserting the whole author document. We also have the `gorethink_ref` tag which tells GoRethink to look for the `id` field in the `Author` document, without this tag GoRethink would instead look for the `author_id` field.

```go
type Author struct {
    ID      string  `gorethink:"id,omitempty"`
    Name    string  `gorethink:"name"`
}

type Book struct {
    ID      string  `gorethink:"id,omitempty"`
    Title   string  `gorethink:"title"`
    Author  Author `gorethink:"author_id,reference" gorethink_ref:"id"`
}
```

The resulting data in RethinkDB should look something like this:

```json
{
    "author_id": "author_1",
    "id":  "book_1",
    "title":  "The Hobbit"
}
```

If you wanted to read back the book with the author included then you could run the following GoRethink query:

```go
r.Table("books").Get("1").Merge(func(p r.Term) interface{} {
    return map[string]interface{}{
        "author_id": r.Table("authors").Get(p.Field("author_id")),
    }
}).Run(session)
```

You are also able to reference an array of documents, for example if each book stored multiple authors you could do the following:

```go
type Book struct {
    ID       string  `gorethink:"id,omitempty"`
    Title    string  `gorethink:"title"`
    Authors  []Author `gorethink:"author_ids,reference" gorethink_ref:"id"`
}
```

```json
{
    "author_ids": ["author_1", "author_2"],
    "id":  "book_1",
    "title":  "The Hobbit"
}
```

The query for reading the data back is slightly more complicated but is very similar:

```go
r.Table("books").Get("book_1").Merge(func(p r.Term) interface{} {
    return map[string]interface{}{
        "author_ids": r.Table("authors").GetAll(r.Args(p.Field("author_ids"))).CoerceTo("array"),
    }
})
```

### Custom `Marshaler`s/`Unmarshaler`s

Sometimes the default behaviour for converting Go types to and from ReQL is not desired, for these situations the driver allows you to implement both the [`Marshaler`](https://godoc.org/github.com/gorethink/gorethink/encoding#Marshaler) and [`Unmarshaler`](https://godoc.org/github.com/gorethink/gorethink/encoding#Unmarshaler) interfaces. These interfaces might look familiar if you are using to using the `encoding/json` package however instead of dealing with `[]byte` the interfaces deal with `interface{}` values (which are later encoded by the `encoding/json` package when communicating with the database).

An good example of how to use these interfaces is in the [`types`](https://github.com/gorethink/gorethink/blob/master/types/geometry.go#L84-L106) package, in this package the `Point` type is encoded as the `GEOMETRY` pseudo-type instead of a normal JSON object.

## Logging

By default the driver logs are disabled however when enabled the driver will log errors when it fails to connect to the database. If you would like more verbose error logging you can call `r.SetVerbose(true)`.

Alternatively if you wish to modify the logging behaviour you can modify the logger provided by `github.com/sirupsen/logrus`. For example the following code completely disable the logger:

```go
// Enabled
r.Log.Out = os.Stderr
// Disabled
r.Log.Out = ioutil.Discard
```

## Mocking

The driver includes the ability to mock queries meaning that you can test your code without needing to talk to a real RethinkDB cluster, this is perfect for ensuring that your application has high unit test coverage.

To write tests with mocking you should create an instance of `Mock` and then setup expectations using `On` and `Return`. Expectations allow you to define what results should be returned when a known query is executed, they are configured by passing the query term you want to mock to `On` and then the response and error to `Return`, if a non-nil error is passed to `Return` then any time that query is executed the error will be returned, if no error is passed then a cursor will be built using the value passed to `Return`. Once all your expectations have been created you should then execute you queries using the `Mock` instead of a `Session`.

Here is an example that shows how to mock a query that returns multiple rows and the resulting cursor can be used as normal.

```go
func TestSomething(t *testing.T) {
	mock := r.NewMock()
	mock.On(r.Table("people")).Return([]interface{}{
		map[string]interface{}{"id": 1, "name": "John Smith"},
		map[string]interface{}{"id": 2, "name": "Jane Smith"},
	}, nil)

	cursor, err := r.Table("people").Run(mock)
	if err != nil {
		t.Errorf("err is: %v", err)
	}

	var rows []interface{}
	err = cursor.All(&rows)
	if err != nil {
		t.Errorf("err is: %v", err)
	}

	// Test result of rows

	mock.AssertExpectations(t)
}
```

The mocking implementation is based on amazing https://github.com/stretchr/testify library, thanks to @stretchr for their awesome work!

## Benchmarks

Everyone wants their project's benchmarks to be speedy. And while we know that rethinkDb and the gorethink driver are quite fast, our primary goal is for our benchmarks to be correct. They are designed to give you, the user, an accurate picture of writes per second (w/s). If you come up with a accurate test that meets this aim, submit a pull request please.

Thanks to @jaredfolkins for the contribution.

| Type    |  Value   |
| --- | --- |
| **Model Name** | MacBook Pro |
| **Model Identifier** | MacBookPro11,3 |
| **Processor Name** | Intel Core i7 |
| **Processor Speed** | 2.3 GHz |
| **Number of Processors** | 1 |
| **Total Number of Cores** | 4 |
| **L2 Cache (per Core)** | 256 KB |
| **L3 Cache** | 6 MB |
| **Memory** | 16 GB |

```bash
BenchmarkBatch200RandomWrites                20                              557227775                     ns/op
BenchmarkBatch200RandomWritesParallel10      30                              354465417                     ns/op
BenchmarkBatch200SoftRandomWritesParallel10  100                             761639276                     ns/op
BenchmarkRandomWrites                        100                             10456580                      ns/op
BenchmarkRandomWritesParallel10              1000                            1614175                       ns/op
BenchmarkRandomSoftWrites                    3000                            589660                        ns/op
BenchmarkRandomSoftWritesParallel10          10000                           247588                        ns/op
BenchmarkSequentialWrites                    50                              24408285                      ns/op
BenchmarkSequentialWritesParallel10          1000                            1755373                       ns/op
BenchmarkSequentialSoftWrites                3000                            631211                        ns/op
BenchmarkSequentialSoftWritesParallel10      10000                           263481                        ns/op
```

## Examples

Many functions have examples and are viewable in the godoc, alternatively view some more full features examples on the [wiki](https://github.com/gorethink/gorethink/wiki/Examples).

Another good place to find examples are the tests, almost every term will have a couple of tests that demonstrate how they can be used.

## Further reading

- [GoRethink Goes 1.0](https://www.compose.io/articles/gorethink-goes-1-0/)
- [Go, RethinkDB & Changefeeds](https://www.compose.io/articles/go-rethinkdb-and-changefeeds-part-1/)
- [Build an IRC bot in Go with RethinkDB changefeeds](http://rethinkdb.com/blog/go-irc-bot/)

## License

Copyright 2013 Daniel Cannon

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

## Donations

[![Donations](https://pledgie.com/campaigns/29517.png "Donations")](https://pledgie.com/campaigns/29517)
