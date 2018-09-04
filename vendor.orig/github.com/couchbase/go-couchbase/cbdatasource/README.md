cbdatasource
============

golang library to stream data from a Couchbase cluster

The cbdatasource package is implemented using Couchbase DCP protocol
and has auto-reconnecting and auto-restarting goroutines underneath
the hood to provide a simple, high-level cluster-wide abstraction.  By
using cbdatasource, your application does not need to worry about
connections or reconnections to individual server nodes or cluster
topology changes, rebalance & failovers.  The API starting point is
NewBucketDataSource().

LICENSE: Apache 2.0

### Status & Links

[![GoDoc](https://godoc.org/github.com/couchbase/go-couchbase/cbdatasource?status.svg)](https://godoc.org/github.com/steveyen/cbdatasource)
