# Change History

## April 25 2017: v1.27.0

  Feature, Performance improvements and bug fix release.

  * **New Features**

    - Added `BatchGetObjects` method.
    - Added Exponential Backoff by introducing `BasePolicy.SleepMultiplier`. Only Values > 1.0 are effective. PR #192, thanks to [Venil Noronha](https://github.com/venilnoronha)

  * **Improvements**

    - Packer tries to see if it can use generic data types before using reflection.
    - Operations, including CDTs do not allocate a buffer anymore, unless reused.

  * **Incompatible changes**:
    - `BinName` and `BinValue` are not exported in `Operation` anymore. These fields shouldn't have been used anyway since `Operation`s used to cache their internal command.

  * **Fixes**

    - Documentation Fixes. Thanks to [Nassor Paulino da Silva](https://github.com/nassor) and [HArmen](https://github.com/alicebob)


## April 5 2017: v1.26.0

  Feature, Performance improvements and bug fix release.

  * **New Features**

    - Predicate API is supported (for server v3.12+)
    - Added `Truncate` method to quickly remove all data from namespaces or sets (for server v3.12+).
    - Support `ScanPolicy.ServerSocketTimeout` (for server v3.12+).
    - Support `ClientPolicy.IgnoreOtherSubnetAliases` to ignore hosts from other subnets. PR #182, thanks to [wedi-dev](https://github.com/wedi-dev)

  * **Improvements**

    - Added a lot of predefined generic slice and map types in `NewValue` method to avoid hitting reflection as much as possible.
    - Fix `go vet` complaints.

  * **Fixes**

    - Allow streaming commands (scan/query/aggregation) to retry unless the error occurs during parsing of the results. Fixes issue #187
    - Use `net.JoinHostPort` to concatinate host and port values instead of doing it directly. Fixes some issues in IPv6 connection strings.
    - Improved initial Tend run.
    - Fixes `cluster-name` checking bug.

## March 8 2017: v1.25.1

  Hot fix release. Updating the client is recommended.

  * **Fixes**

    - Fixed an issue where errors in Scan/Query unmarshalling would be duplicated and could cause a deadlock.

## February 28 2017: v1.25.0

  Performance improvements and fix release.

  * **Improvements**

    - Check tend duration and compare it to tend interval, and warn the user if tend takes longer than tend interval.
    - Seed the cluster concurrently, and return as soon as any of the seeds is validated.
    - Tend the cluster concurrently. Allows use of very big clusters with no delay.
    - Partitions the connection queue to avoid contention.
    - Cluster partition map is merged from all node fragments and updated only once per tend to reduce contention to absolute minimum.

  * **Fixes**

    - Fixed an issue where a valid but unreachable seed could timeout and stall connecting and tending the cluster..
    - Fix result code comments.

## January 11 2017: v1.24.0

  Minor feature and fix release.

  * **New Features**

    - TLS/SSL connections are now officially supported.
    - Added Role/Privilege API.

  * **Improvements**

    - Return a client-side error when no ops are passed to the operate command.
    - Export error attribute in `NodeError`
    - Do not attempt to refresh peers if it is not supported by the nodes.

  * **Fixes**

    - Use namespace default-ttl for tests instead of assuming 30d
    - Always drain scan connections after parsing the records.
    - Fix panic in GetObject() if all bins in result is nil. PR #172, thanks to [Hamper](https://github.com/hamper)
    - Fix WritePolicy usage with UDF. PR #174, thanks to [Bertrand Paquet](https://github.com/bpaquet)
    - Close connection right when it has an io error and don't wait for the caller.

## December 20 2016 : v1.23.0

  Minor feature and fix release.

  * **New Features**

    - Exposes the internal `client.Cluster` object to the users.
    - Added New API for high-performance complex data type packing, and removed the old API.

  * **Improvements**

    - Only update the partition map if the partition generatio has changed.
    - Use tend connection for user management commands.
    - Marks LargeList as deprecated. Use CDT methods instead.
    - Always validate the message header to avoid reading the remainder of other command buffers.
    - Removes GeoJson from key helper.
    - Improves tend algorthm to allow complete disconnection from the cluster if none of the clusters are accessible.
    - `PutObject` method will now accept objects as well. PR #156, thanks to [Sarath S Pillai](https://github.com/sarathsp06)

  * **Fixes**

    - Do not attemp to add a node which were unaccessible to avoid panic.
    - Fix invalid connectionCount. PR #168, thanks to [Jun Kimura](https://github.com/bluele)
    - Fixes minor bug that didn't return the error on reading from the connection during scans.

## November 29 2016 : v1.22.0

  Hot fix release. Please upgrade if you have been using other aerospike clients with your database parallel to Go.

  * **Fixes**

    - Fixes an issue where short strings in Lists and Maps wouldn't unpack correctly. Resolves #161.

## November 16 2016 : v1.21.0

  Minor fix release.

  * **New Features**

    - Added new constants for expiration in `WritePolicy`: `TTLServerDefault`, `TTLDontExpire`, `TTLDontUpdate`

  * **Improvements**

    - Corrects typos in the code. PR #142, thanks to [Muyiwa Olurin ](https://github.com/muyiwaolurin)
    - Use the tend connection for `RequestInfo` commands.

  * **Fixes**

    - Fixes an issue where TTL values were calcualted wrongly when they were set not to expire.
    - Fixes an issue where `PutObjects` would marshal `[]byte` to `List` in database. PR #152, thanks to [blide](https://github.com/blide)
    - Fixes an issue where `Recordset` could leak goroutines. PR #153, thanks to [Deepak Prabhakara](https://github.com/deepakprabhakara)

## October 25 2016 : v1.20.0

  Major improvements release. There has been major changes in the library. Please test rigorously before upgrading to the new version.

  * **New Features**

    - Let user define the desired tag for bin names in structs using `SetAerospikeTag` function.
    - Added `as_performance` build tag to avoid including the slow convenience API which uses reflections in the client code.
      To use this feature, you should include -tags="as_performance" when building your project.

      *NOTICE*: Keep in mind that your code may not compile using this flag. That is by design.

  * **Improvements**

    - Added special packer for map[string]interface{} in `NewValue` method.
    - Avoid allocating memory for Map and List values.
    - Allocate commands on the stack to avoid heap allcations.
    - Avoid allocating memory for `packer`.
    - Avoid Allocating memory in computeHash for keys.
    - Avoid allocating memory in Ripe160MD digest.
    - Removed BufferPool and moved buffers to `Connection` objects to remove lock contention.
    - Added `ListIter` and `MapIter` interfaces to support passing Maps and Lists to the client without using reflection.

## October 14 2016 : v1.19.0

  Major feature and improvement release.

  * **New Features**

    * Support TLS secured connections. (Feature will be supported in coming server releases.)

    * Support IPv6 protocol. Supported by Aerospike Server 3.10+.

    * Support `cluster-name` verification. Supported by Aerospike Server 3.10+.

    * Support new peers info protocol. Supported by Aerospike Server 3.10+.

  * **Improvements**

    * Will retry the operation even when reading from the buffer. Set `Policy.MaxRetries = 0` to avoid this behavior. PR #143, thanks to [Hector Jusforgues](https://github.com/hectorj)

    * Much improved cluster management algorithm. Will now handle the case where multiple nodes go down simultaneously, still protecting against split brain rogue nodes.

  * **Fixes**

    * Try all alias IPs in node validator. Resolves #144.

    * Updated job status check for execute tasks.

## August 19 2016 : v1.18.0

  Minor improvements release.

  * **New Features**

    * Support 'Durable Deletes' for the next version of Aerospike Server Enterprise.

  * **Improvements**

    * Don't run tests for features that are not supported by the server.

    * Added new server error codes.


## July 27 2016 : v1.17.1

  Minor improvements release.

  * **Improvements**

    * Add `TaskId()` method for `Recordset`.

    * Cleanup indexes after test cases.

    * Keep connections on recoverable server errors.

    * Return the error on unexpected keys in `BatchCommandGet/Header`.

    * Use the same client object in tests and support using replicas on travis.

## July 19 2016 : v1.17.0

  Major feature and improvement release.

  * **New Features**

    * Client now supports distributing reads from Replicas using `ClientPolicy.RequestProleReplicas` and `Policy.ReplicaPolicy`

  * **Improvements**

    * `Cluster.GetConnection` will now retry to acquire a connection until timeout.

    * `Client.DropIndex` method now blocks until all nodes report the index is dropped.

    * Async tasks like `CreateIndex` will retry a few times before deciding a non-existing job means it has finished.

    * Don't use math.MaxInt64, it breaks 32-bit builds. PR #139, thanks to [Cameron Sparr](https://github.com/sparrc)

  * **Fixes**

    * Maps with 0 elements will automatically shortcut to unordered empty maps.

    * Return the error in BatchCommandGet on parse error.

## June 28 2016 : v1.16.3

  Major bugfix release. Update recommended.

  * **Improvements**

    * Skip LDT tests if LDT is not enabled.

    * Returns last error after all retry attempts to run a command are exhausted.

    * Reserves a connection for tend operation to avoid dropping a node when high load prevents acquiring a proper connection.

    * Added Finalizers to `Client` and `Recordset`. Both will be automatically closed by the GC.

  * **Fixes**

    * Fixes an issue where `services-alternate` wasn't used in `Node.addFriends()` when instructed so in the policy.

    * Fixes an issue where object metadata wasn't cached if `QueryObjects` was called before `PutObject`.

    * Fixes an issue where idle connections were not dropped.

    * Fixes an issue where requested buffer sizes were not guarded against negative numbers.

## June 7 2016 : v1.16.2

  Minor bugfix release.

  * **Fixes**

    * Fixes an issue where empty unordered maps were confused with CDT maps.

## June 6 2016 : v1.16.1

  Minor bugfix release.

  * **Fixes**

    * Fixes an issue where complex maps and lists weren't unmarshalled correctly in `GetObject` method.

## June 2 2016 : v1.16

  Major feature and improvements release.

  > NOTICE: Due to the relatively extensive code overhaul, upgrade with caution.

  * **New Features**

    * Added CDT Ordered Map API. (Requires server v3.8.3+)

  * **Improvements**

    * Removed mutexes from `Cluster` and `Node` code.

    * Improved code quality using various linters.

## May 27 2016 : v1.15

  Minor fixes and improvements release.

  * **Fixes**

    * Fixed an issue where unmarshalling embedded structs and pointers didn't work properly if they were tagged.

## May 16 2016 : v1.14

  Minor fixes and improvements release.

  * **Fixes**

    * Fixed an issue in which go-routines were leaked in `Results()` method of `Recordset` on cancellation. Based on PR #128, thanks to [Noel Cower](https://github.com/nilium)

    * Fixed issues regarding leaked goroutines in `Cluster.WaitTillStablized()`, `Cluster.MigrationInProgress()`, and `Cluster.WaitUntillMigrationIsFinished()` methods. PR #126, thanks to [Anton](https://github.com/yiiton)

  * **Improvements**

    * Improved cluster `tend()` logic.

    * Added `Recordset.Read()` method.

    * Minor fixes in docs and code formatting. Thanks to [Andrew Murray](https://github.com/radarhere) and [Erik Dubbelboer](https://github.com/erikdubbelboer)

## April 1 2016 : v1.13

  Minor features and improvements release.

  * **New Features**

    * Added `NewGeoWithinRegionForCollectionFilter`, `NewGeoRegionsContainingPointForCollectionFilter`, `NewGeoWithinRadiusForCollectionFilter` for queries on collection bins.

  * **Fixes**

    * Fixed an issue in which bounded byte arrays were silently being dropped as map keys.

  * **Improvements**

    * Removed and fixed unused assignments and variables.

    * Fixed typos in the comments.

    * Minor changes and formatting. PR #124, thanks to [Harmen](https://github.com/alicebob)

## March 8 2016 : v1.12

  Minor features and improvements release.

  * **New Features**

    * Support Metadata in struct tags to fetch TTL and Generation via `GetObject`.
    Notice: Metadata attributes in an struct are considered transient, and won't be persisted.

    Example:
    ```go
    type SomeStruct struct {
      TTL  uint32         `asm:"ttl"` // record time-to-live in seconds
      Gen  uint32         `asm:"gen"` // record generation
      A    int
      Self *SomeStruct
    }

    key, _ := as.NewKey("ns", "set", value)
    err := client.PutObject(nil, key, obj)
    // handle error here

    rObj := &OtherStruct{}
    err = client.GetObject(nil, key, rObj)
    ```

    * GeoJSON support in Lists and Maps

  * **Improvements**

    * Use `ClientPolicy.timeout` for connection timeout when refreshing nodes

    * Added new server error codes

    * Protect RNG pool against low-precision clocks during init

    * Better error message distingushing between timeout because of reaching deadline and exceeding maximum retries

  * **Fixes**

    * Fixed object mapping cache for anonymous structs. PR #115, thanks to [Moshe Revah](https://github.com/zippoxer)

    * Fixed an issue where `Execute()` method wasn't observing the `SendKey` flag in Policy.

## February 9 2016 : v1.11

  Minor features and improvements release.

  * **New Features**

      * Can now use `services-alternate` for cluster tend.

      * New CDT List API: `ListGetRangeFromOp`, `ListRemoveRangeFromOp`, `ListPopRangeFromOp`

  * **Improvements**

      * Improves marshalling of data types into and out of the Lua library and avoids marshalling values before they are needed.

      * Returns error for having more than one Filter on client-side to avoid confusion.

      * Increases default `ClientPolicy.Timeout` and return a meaningful error message when the client is not fully connected to the cluster after `waitTillStabilized` call

## January 13 2016 : v1.10

  Major release. Adds Aggregation.

  * **New Features**

    * Added `client.QueryAggregate` method.

      * For examples regarding how to use this feature, look at the examples directory.

      * You can find more documentation regarding the [Aggregation Feature on Aerospike Website](http://www.aerospike.com/docs/guide/aggregation.html)

  * **Improvements**

    * Improve Query/Scan performance by reading from the socket in bigger chunks

## December 14 2015 : v1.9

  Major release. Adds new features.

  * **New Features**

    * Added CDT List operations.

    * Added `NewGeoWithinRadiusFilter` filter for queries.

  * **Changes**

    * Renamed `NewGeoPointsWithinRegionFilter` to `NewGeoWithinRegionFilter`

## December 1 2015 : v1.8

  Major release. Adds new features and fixes important bugs.

  * **New Features**

    * Added `ScanAllObjects`, `ScanNodeObjects`, `QueryObjects` and `QueryNodeObjects` to the client, to facilitate automatic unmarshalling of data similar to `GetObject`.

      * NOTICE: This feature and its API are experimental, and may change in the future. Please test your code throughly, and provide feedback via Github.

    * Added `ScanPolicy.IncludeLDT` option (Usable with yet to be released server v 3.7.0)

    * Added `LargeList.Exist` method.

  * **Improvements**

    * Makes Generation and Expiration values consistent for WritePolicy and Record.

      * NOTICE! BREAKING CHANGE: Types of `Record.Generation` and `Record.Expiration`, and also `WritePolicy.Generation` and `WritePolicy.Expiration` have changed, and may require casting in older code.

    * Refactor tools/asinfo to be more idiomatic Go. PR #86, thanks to [Tyler Gibbons](https://github.com/Kavec)

    * Many documentation fixes thanks to [Charl Matthee](https://github.com/charl) and [Tyler Gibbons](https://github.com/Kavec)

  * **Fixes**

    * Changed the `KeepConnection` logic from black-list to white-list, to drop all

    * Fix RemoveNodesCopy logic error.

    * Add missing send on recordset Error channel. PR #99, thanks to [Geert-Johan Riemer](https://github.com/GeertJohan)

    * Fix skipping of errors/records in (*recordset).Results() select after cancellation. PR #99, thanks to [Geert-Johan Riemer](https://github.com/GeertJohan)

## October 16 2015 : v1.7

  Major release. Adds new features and fixes important bugs.

  * **New Features**

    * Added support for Geo spatial queries.

    * Added support for creating indexes on List and Map bins, and querying them.

    * Added support for native floating point values.

    * Added `ClientPolicy.IpMap` to use IP translation for alias recognition. PR #81, Thanks to [Christopher Guiney](https://github.com/chrisguiney)

  * **Improvements**

    * Cosmetic change to improve code consistency for `PackLong` in `packer.go`. PR #78, Thanks to [Erik Dubbelboer](https://github.com/ErikDubbelboer)

  * **Fixes**

    * Fixes an issue when the info->services string was malformed and caused the client to panic.

    * Fixes an issue with unmarshalling maps of type map[ANY]struct{} into embedded structs.

    * Fixes issue with unmarshalling maps of type map[ANY]struct{} into embedded structs.

    * Fixes an issue with bound checking. PR #85, Thanks to [Tait Clarridge](https://github.com/oldmantaiter)

    * Fixes aa few typos in the docs. PR #76, Thanks to [Charl Matthee](https://github.com/charl)

## August 2015 : v1.6.5

  Minor maintenance release.

  * **Improvements**

    * Export `MaxBufferSize` to allow tweaking of maximum buffer size allowed to read a record. If a record is bigger than this size (e.g: A lot of LDT elements in scan), this setting wil allow to tweak the buffer size.

## July 16 2015 : v1.6.4

  Hot fix release.

  * **Fixes**

    * Fix panic when a scan/query fails and the connection is not dropped.

## July 9 2015 : v1.6.3

  Minor fix release.

  * **Improvements**

    * Improved documentation. PR #64 and #68. Thanks to [Geert-Johan Riemer](https://github.com/GeertJohan)

  * **Fixes**

    * Fix a bunch of golint notices. PR #69, Thanks to [Geert-Johan Riemer](https://github.com/GeertJohan)

    * Connection.Read() total bytes count on error. PR #71, Thanks to [Geert-Johan Riemer](https://github.com/GeertJohan)

    * Fixed a race condition on objectMappings map. PR #72, Thanks to [Geert-Johan Riemer](https://github.com/GeertJohan)

    * Fixed a few uint -> int convertions.

## June 11 2015 : v1.6.2

  Minor fix release.

  * **Improvements**

    * Improved documentation. Replaced all old API references regarding Recordset/Query/Scan to newer, more elegant API.

  * **Fixes**

    * Fixed an issue where erroring out on Scan would result a panic.

    * Fixed an issue where `Statement.TaskId` would be negative. converted `Statement.TaskId` to `uint64`

## June 9 2015 : v1.6.1

  Minor fix release.

  * **Fixes**

    * Fixed an issue where marshaller wouldn't marshal some embedded structs.

    * Fixed an issue where querying/scanning empty sets wouldn't drain the socket before return.

## May 30 2015 : v1.6.0

  There's an important performance regression bug fix in this release. We recommend everyone to upgrade.

  * **New Features**

    * Added New LargeList API.

      * NOTICE! BREAKING CHANGE: New LargeList API on the Go Client uses the New API defined on newer server versions. As Such, it has changed some signatures in LargeList.

  * **Fixes**

    * Fixed an issue where connections where not put back to the pool on some non-critical errors.

    * Fixed an issue where Object Unmarshaller wouldn't extend a slice.

    * Decode RegisterUDF() error message from base64

    * Fixed invalid connection handling on node connections (thanks to @rndive)

## May 15 2015 : v1.5.2

  Hotfix release.

  * **Fixes**

    * Fixed a branch-merge mistake regarding error handling during connection authentication.

## May 15 2015 : v1.5.1

  Major maintenance release.

  NOTICE: All LDTs on server other than LLIST have been deprecated, and will be removed in the future. As Such, all API regarding those features are considered deprecated and will be removed in tandem.

  * **Improvements**

    * Introduces `ClientPolicy.IdleTimeout` to close stale connections to the server. Thanks to Mário Freitas (@imkira). PR #57

    * Use type alias instead of struct for NullValue.

    * Removed workaround regarding filtering bin names on the client for `BatchGet`. Issue #60

  * **Fixes**

    * Fixed a few race conditions.

    * Fixed #58 regarding race condition accessing `Cluster.password`.

    * Fixed minor bugs regarding handling of nulls in structs for `GetObj()` and `PutObj()`.

    * Fixed a bug regarding setting TaskIds on the client.

  * ** Other Changes **

    * Removed deprecated `ReplaceRoles()` method.

    * Removed deprecated `SetCapacity()` and `GetCapacity()` methods for LDTs.

## April 13 2015 : v1.5.0

  This release includes potential BREAKING CHANGES.

  * **New Features**

    * Introduces `ClientPolicy.LimitConnectionsToQueueSize`. If set to true, the client won't attemp to create new connections to the node if the total number of pooled connections to the node is equal or more than the pool size. The client will retry to poll a connection from the queue until a timeout occurs. If no timeout is set, it will only retry for ten times.

  * **Improvements**

    * BREAKING CHANGE: |
                        Uses type aliases instead of structs in several XXXValue methods. This removes a memory allocation per `Value` usage.
                        Since every `Put` operation uses at list one value object, this has the potential to improve application performance.
                        Since the signature of several `NewXXXValue` methods have changed, this might break some existing code if you have used the value objects directly.

    * Improved `Logger` so that it will accept a generalized `Logger` interface. Any Logger with a `Printf(format string, values ...interface{})` method can be used. Examples include Logrus.

    * Improved `Client.BatchGet()` performance.

  * **Fixes**

    * Bin names were ignored in BatchCommands.

    * `BatchCommandGet.parseRecord()` returned wrong values when `BinNames` was empty but not nil.

## March 31 2015 : v1.4.2

  Maintenance release.

  * **Improvements**

    * Replace channel-based queue system with a lock-based algorithm.
    * Marshaller now supports arrays of arbitrary types.
    * `Client.GetObject()` now returns an error when the object is not found.
    * Partition calculation uses a trick that is twice as fast.

  * **Improvements**

    * Unpacking BLOBs resulted in returning references to pooled buffers. Now copies are returned.

## March 12 2015 : v1.4.1

  This is a minor release to help improve the compatibility of the client on Mac OS, and to make cross compilation easier.

  * **Improvements**

    * Node validator won't call net.HostLookup if an IP is passed as a seed to it.

## Feb 17 2015 : v1.4.0

  This is a major release, and makes using the client much easier to develop applications.

  * **New Features**

    * Added Marshalling Support for Put and Get operations. Refer to [Marshalling Test](client_object_test.go) to see how to take advantage.
    Same functionality for other APIs will follow soon.
    Example:
    ```go
    type SomeStruct struct {
      A    int            `as:"a"`  // alias the field to a
      Self *SomeStruct    `as:"-"`  // will not persist the field
    }

    type OtherStruct struct {
      i interface{}
      OtherObject *OtherStruct
    }

    obj := &OtherStruct {
      i: 15,
      OtherObject: OtherStruct {A: 18},
    }

    key, _ := as.NewKey("ns", "set", value)
    err := client.PutObject(nil, key, obj)
    // handle error here

    rObj := &OtherStruct{}
    err = client.GetObject(nil, key, rObj)
    ```

    * Added `Recordset.Results()`. Consumers of a recordset do not have to implement a select anymore. Instead of:
    ```go
    recordset, err := client.ScanAll(...)
    L:
    for {
      select {
      case r := <-recordset.Record:
        if r == nil {
          break L
        }
        // process record here
      case e := <-recordset.Errors:
        // handle error here
      }
    }
    ```

    one should only range on `recordset.Results()`:

    ```go
    recordset, err := client.ScanAll(...)
    for res := range recordset.Results() {
      if res.Err != nil {
        // handle error here
      } else {
        // process record here
        fmt.Println(res.Record.Bins)
      }
    }
    ```

    Use of the old pattern is discouraged and deprecated, and direct access to recordset.Records and recordset.Errors will be removed in a future release.

  * **Improvements**

    * Custom Types are now allowed as bin values.

## Jan 26 2015 : v1.3.1

  * **Improvements**

    * Removed dependency on `unsafe` package.

## Jan 20 2015 : v1.3.0

  * **Breaking Changes**

    * Removed `Record.Duplicates` and `GenerationPolicy/DUPLICATE`

  * **New Features**

    * Added Security Features: Please consult [Security Docs](https://www.aerospike.com/docs/guide/security.html) on Aerospike website.

      * `ClientPolicy.User`, `ClientPolicy.Password`
      * `Client.CreateUser()`, `Client.DropUser()`, `Client.ChangePassword()`
      * `Client.GrantRoles()`, `Client.RevokeRoles()`, `Client.ReplaceRoles()`
      * `Client.QueryUser()`, `Client.QueryUsers`

    * Added `Client.QueryNode()`

    * Added `ClientPolicy.TendInterval`

  * **Improvements**

    * Cleaned up Scan/Query/Recordset concurrent code

  * **Fixes**

      * Fixed a bug in `tools/cli/cli.go`.

      * Fixed a bug when `GetHeaderOp()` would always translate into `GetOp()`

## Dec 29 2014: v1.2.0

  * **New Features**

    * Added `NewKeyWithDigest()` method. You can now create keys with custom digests, or only using digests without
      knowing the original value. (Useful when you are getting back results with Query and Scan)

## Dec 22 2014

  * **New Features**

    * Added `ConsistencyLevel` to `BasePolicy`.

    * Added `CommitLevel` to `WritePolicy`.

    * Added `LargeList.Range` and `LargeList.RangeThenFilter` methods.

    * Added `LargeMap.Exists` method.

  * **Improvements**

    * We use a pooled XORShift RNG to produce random numbers in the client. It is FAST.

## Dec 19 2014

  * **Fixes**

    * `Record.Expiration` wasn't converted to TTL values on `Client.BatchGet`, `Client.Scan` and `Client.Query`.

## Dec 10 2014

  * **Fixes**:

    * Fixed issue when the size of key field would not be estimated correctly when WritePolicy.SendKey was set.

## Nov 27 2014

  Major Performance Enhancements. Minor new features and fixes.

  * **Improvements**

    * Go client is much faster and more memory efficient now.
      In some workloads, it competes and wins against C and Java clients.

    * Complex objects are now de/serialized much faster.

  * **New Features**

    * Added Default Policies for Client object.
      Instead of creating a new policy when the passed policy is nil, default policies will be used.

## Nov 24 2014

  * **Fixes**:

    * Fixed issue when WritePolicy.SendKey = true was not respected in Touch() and Operate()

## Nov 22 2014

  Hotfix in unpacker. Update strongly recommended for everyone using Complex objects, LDTs and UDFs.

  * **Fixes**:

    * When Blob, ByteArray or String size has a bit sign set, unpacker reads it wrong.
        Note: This bug only affects unpacking of these objects. Packing was unaffected, and data in the database is valid.

## Nov 2 2014

  Minor, but very impoortant fix.

  * **Fixes**:

    * Node selection in partition map was flawed on first refresh.

  * **Incompatible changes**:

    * `Expiration` and `Generation` in `WritePolicy` are now `int32`
    * `TaskId` in `Statement` is now always set in the client, and is `int64`

  * **New Features**:

    * float32, float64 and bool are now supported in map and array types

## Oct 15 2014 (Beta 2)

  * **Hot fix**:

    * Fixed pack/unpack for uint64

## Aug 20 2014 (Beta 1)

  Major changes and improvements.

  * **New Features**:

    * Added client.Query()
    * Added client.ScanNode()/All()
    * Added client.Operate()
    * Added client.CreateIndex()
    * Added client.DropIndex()
    * Added client.RegisterUDF()
    * Added client.RegisterUDFFromFile()
    * Added client.Execute()
    * Added client.ExecuteUDF()
    * Added client.BatchGet()
    * Added client.BatchGetHeader()
    * Added client.BatchExists()
    * Added LDT implementation
    * Added `Node` and `Key` references to the Record

  * **Changes**:

    * Many minor and major bug fixes
    * Potentially breaking change: Reduced Undocumented API surface
    * Fixed a few places where error results were not checked
    * Breaking Change: Convert Key.namespace & Key.setName from pointer to string; affects Key API
    * Renamed all `this` receivers to appropriate names
    * Major performance improvements (~2X improvements in speed and memory consumption):
      * better memory management for commands; won't allocate if capacity is big enough
      * better hash management in key; avoids two redundant memory allocs
      * use a buffer pool to reduce GC load
      * fine-grained, customizable and deterministic buffer pool implementation for command

    * Optimizations for Key & Digest
      * changed digest implementation, removed an allocation
      * Added RIPEMD160 hash files from crypto to lib
      * pool hash objects

    * Various Benchmark tool improvements
      * now profileable using localhost:6060
      * minor bug fixes

## Jul 26 2014 (Alpha)

  * Initial Release.
