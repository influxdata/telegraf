# Change Log
All notable changes to this project will be documented in this file.
This project adheres to [Semantic Versioning](http://semver.org/).

## v1.0.0 - 2015-06-27

1.0.0 is finally here, This is the first stable production ready release of GoRethink!

![GoRethink Logo](https://raw.github.com/wiki/dancannon/gorethink/gopher-and-thinker.png "Golang Gopher and RethinkDB Thinker")

In an attempt to make this library more "idiomatic" some functions have been renamed, for the full list of changes and bug fixes see below.

### Added
 - Added more documentation.
 - Added `Shards`, `Replicas` and `PrimaryReplicaTag` optional arguments in `TableCreateOpts`.
 - Added `MultiGroup` and `MultiGroupByIndex` which are equivalent to the running `group` with the `multi` optional argument set to true.

### Changed 
 - Renamed `Db` to `DB`.
 - Renamed `DbCreate` to `DBCreate`.
 - Renamed `DbDrop` to `DBDrop`.
 - Renamed `RqlConnectionError` to `RQLConnectionError`.
 - Renamed `RqlDriverError` to `RQLDriverError`.
 - Renamed `RqlClientError` to `RQLClientError`.
 - Renamed `RqlRuntimeError` to `RQLRuntimeError`.
 - Renamed `RqlCompileError` to `RQLCompileError`.
 - Renamed `Js` to `JS`.
 - Renamed `Json` to `JSON`.
 - Renamed `Http` to `HTTP`.
 - Renamed `GeoJson` to `GeoJSON`.
 - Renamed `ToGeoJson` to `ToGeoJSON`.
 - Renamed `WriteChanges` to `ChangeResponse`, this is now a general type and can be used when dealing with changefeeds.
 - Removed depth limit when encoding values using `Expr`

### Fixed
 - Fixed issue causing errors when closing a changefeed cursor (#191)
 - Fixed issue causing nodes to remain unhealthy when host discovery is disabled (#195)
 - Fixed issue causing driver to fail when connecting to DB which did not have its canonical address set correctly (#200).
- Fixed ongoing queries not being properly stopped when closing the cursor.

### Removed
 - Removed `CacheSize` and `DataCenter` optional arguments in `TableCreateOpts`.
 - Removed `CacheSize` optional argument from `InsertOpts`

## v0.7.2 - 2015-05-05
### Added
 - Added support for connecting to a server using TLS (#179)

### Fixed
 - Fixed issue causing driver to fail to connect to servers with the HTTP admin interface disabled (#181)
 - Fixed errors in documentation (#182, #184)
 - Fixed RunWrite not closing the cursor (#185)

## v0.7.1 - 2015-04-19
### Changed
- Improved logging of connection errors.

### Fixed
- Fixed bug causing empty times to be inserted into the DB even when the omitempty tag was set.
- Fixed node status refresh loop leaking goroutines.

## v0.7.0 - 2015-03-30

This release includes support for RethinkDB 2.0 and connecting to clusters. To connect to a cluster you should use the new `Addresses` field in `ConnectOpts`, for example:

```go
session, err := r.Connect(r.ConnectOpts{
    Addresses: []string{"localhost:28015", "localhost:28016"},
})
if err != nil {
    log.Fatalln(err.Error())
}
```

Also added was the ability to read from a cursor using a channel, this is especially useful when using changefeeds. For more information see this [gist](https://gist.github.com/dancannon/2865686d163ed78bbc3c)

```go
cursor, err := r.Table("items").Changes()
ch := make(chan map[string]interface{})
cursor.Listen(ch)
```

For more details checkout the [README](https://github.com/dancannon/gorethink/blob/master/README.md) and [godoc](https://godoc.org/github.com/dancannon/gorethink). As always if you have any further questions send me a message on [Gitter](https://gitter.im/dancannon/gorethink).

- Added the ability to connect to multiple nodes, queries are then distributed between these nodes. If a node stops responding then queries stop being sent to this node.
- Added the `DiscoverHosts` optional argument to `ConnectOpts`, when this value is `true` the driver will listen for new nodes added to the cluster.
- Added the `Addresses` optional argument to `ConnectOpts`, this allows the driver to connect to multiple nodes in a cluster.
- Added the `IncludeStates` optional argument to `Changes`.
- Added `MinVal` and `MaxVal` which represent the smallest and largest possible values.
- Added the `Listen` cursor helper function which publishes database results to a channel.
- Added support for optional  arguments for the `Wait` function.
- Added the `Type` function to the `Cursor`, by default this value will be "Cursor" unless using a changefeed.
- Changed the `IndexesOf` function to `OffsetsOf` .
- Changed driver to use the v0.4 protocol (used to use v0.3).
- Fixed geometry tests not properly checking the expected results.
- Fixed bug causing nil pointer panics when using an `Unmarshaler`
- Fixed dropped millisecond precision if given value is too old

## v0.6.3 - 2015-03-04
### Added
- Add `IdentifierFormat` optarg to `TableOpts` (#158)

### Fixed
- Fix struct alignment for ARM and x86-32 builds (#153)
- Fix sprintf format for geometry error message (#157)
- Fix duplicate if block (#159)
- Fix incorrect assertion in decoder tests

## v0.6.2 - 2015-02-15

- Fixed `writeQuery` being too small when sending large queries

## v0.6.1 - 2015-02-13

- Reduce GC by using buffers when reading and writing
- Fixed encoding `time.Time` ignoring millseconds
- Fixed pointers in structs that implement the `Marshaler`/`Unmarshaler` interfaces being ignored

## v0.6.0 - 2015-01-01

There are some major changes to the driver with this release that are not related to the RethinkDB v1.16 release. Please have a read through them:
- Improvements to result decoding by caching reflection calls.
- Finished implementing the `Marshaler`/`Unmarshaler` interfaces
- Connection pool overhauled. There were a couple of issues with connections in the previous releases so this release replaces the `fatih/pool` package with a connection pool based on the `database/sql` connection pool.
- Another change is the removal of the prefetching mechanism as the connection+cursor logic was becoming quite complex and causing bugs, hopefully this will be added back in the near future but for now I am focusing my efforts on ensuring the driver is as stable as possible #130 #137
- Due to the above change the API for connecting has changed slightly (The API is now closer to the `database/sql` API. `ConnectOpts` changes:
  - `MaxActive` renamed to `MaxOpen`
  - `IdleTimeout` renamed to `Timeout`
- `Cursor`s are now only closed automatically when calling either `All` or `One`
- `Exec` now takes `ExecOpts` instead of `RunOpts`. The only difference is that `Exec` has the `NoReply` field

With that out the way here are the v1.16 changes:

- Added `Range` which generates all numbers from a given range
- Added an optional squash argument to the changes command, which lets the server combine multiple changes to the same document (defaults to true)
- Added new admin functions (`Config`, `Rebalance`, `Reconfigure`, `Status`, `Wait`)
- Added support for `SUCCESS_ATOM_FEED`
- Added `MinIndex` + `MaxInde`x functions
- Added `ToJSON` function
- Updated `WriteResponse` type

Since this release has a lot of changes and although I have tested these changes sometimes things fall through the gaps. If you discover any bugs please let me know and I will try to fix them as soon as possible.

## v.0.5.1 - 2014-12-14

- Fixed empty slices being returned as `[]T(nil)` not `[]T{}` #138

## v0.5.0 - 2014-10-06

- Added geospatial terms (`Circle`, `Distance`, `Fill`, `Geojson`, `ToGeojson`, `GetIntersecting`, `GetNearest`, `Includes`, `Intersects`, `Line`, `Point`, `Polygon`, `PolygonSub`)
- Added `UUID` term for generating unique IDs
- Added `AtIndex` term, combines `Nth` and `GetField`
- Added the `Geometry` type, see the types package
- Updated the `BatchConf` field in `RunOpts`, now uses the `BatchOpts` type
- Removed support for the `FieldMapper` interface

Internal Changes

- Fixed encoding performance issues, greatly improves writes/second
- Updated `Next` to zero the destination value every time it is called.

## v0.4.2 - 2014-09-06

- Fixed issue causing `Close` to start an infinite loop
- Tidied up connection closing logic

## v0.4.1 - 2014-09-05

- Fixed bug causing Pseudotypes to not be decoded properly (#117)
- Updated github.com/fatih/pool to v2 (#118)

## v0.4.0 - 2014-08-13

- Updated the driver to support RethinkDB v1.14 (#116)
- Added the Binary data type
- Added the Binary command which takes a `[]byte` or `bytes.Buffer{}` as an argument.
- Added the `BinaryFormat` optional argument to `RunOpts` 
- Added the `GroupFormat` optional argument to `RunOpts` 
- Added the `ArrayLimit` optional argument to `RunOpts` 
- Renamed the `ReturnVals` optional argument to `ReturnChanges` 
- Renamed the `Upsert` optional argument to `Conflict` 
- Added the `IndexRename` command
- Updated `Distinct` to now take the `Index` optional argument (using `DistinctOpts`)

Internal Changes

- Updated to use the new JSON protocol
- Switched the connection pool code to use github.com/fatih/pool
- Added some benchmarks

## v0.3.2 - 2014-08-17

- Fixed issue causing connections not to be closed correctly (#109)
- Fixed issue causing terms in optional arguments to be encoded incorrectly (#114)

## v0.3.1 - 2014-06-14

- Fixed "Token ## not in stream cache" error (#103)
- Changed Exec to no longer use NoReply. It now waits for the server to respond.

## v0.3.0 - 2014-06-26

- Replaced `ResultRows`/`ResultRow` with `Cursor`, `Cursor` has the `Next`, `All` and `One` methods which stores the relevant value in the value pointed at by result. For more information check the examples.
- Changed the time constants (Days and Months) to package globals instead of functions
- Added the `Args` term and changed the arguments for many terms to `args ...interface{}` to allow argument splicing
- Added the `Changes` term and support for the feed response type
- Added the `Random` term
- Added the `Http` term
- The second argument for `Slice` is now optional
- `EqJoin` now accepts a function as its first argument
- `Nth` now returns a selection

## v0.2.0 - 2014-04-13

* Changed `Connect` to use `ConnectOpts` instead of `map[string]interface{}`
* Migrated to new `Group`/`Ungroup` functions, these replace `GroupedMapReduce` and `GroupBy`
* Added new aggregators
* Removed base parameter for `Reduce`
* Added `Object` function
* Added `Upcase`, `Downcase` and `Split` string functions
* Added `GROUPED_DATA` pseudotype
* Fixed query printing

## v0.1.0 - 2013-11-27

* Added noreply writes
* Added the new terms `index_status`, `index_wait` and `sync`
* Added the profile flag to the run functions
* Optional arguments are now structs instead of key, pair strings. Almost all of the struct fields are of type interface{} as they can have terms inside them. For example: `r.TableCreateOpts{ PrimaryKey: r.Expr("index") }`
* Returned arrays are now properly loaded into ResultRows. In the past when running `r.Expr([]interface{}{1,2,3})` would require you to use `RunRow` followed by `Scan`. You can now use `Run` followed by `ScanAll`
