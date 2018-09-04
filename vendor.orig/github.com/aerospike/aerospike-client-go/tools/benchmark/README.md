# Benchmark Tool

Benchmark tool is intended to generate artificial yet customizable load on your database cluster to help you tweak your connection properties.


## Usage

To see available switches:

```$ ./benchmark -u```

## How it works

By default, load is generated on keys with values in key range (```-k``` switch). Bin data is static by default.

To generate random bin data, use ```-R``` switch. To specify the type of bin data, use ```-o``` switch. By default it is set to 64 bit integer values.

## Considerations

In our lab tests, we have observed that a concurrency level of 16 can easily saturate a database node. Increasing concurrency level beyond that doesn't increase server throughput.

The client is sensitive to timeouts, and they should be chosen carefully. Connection Timeouts are set using ClientPolicy object, while data operation timeouts are set in their respective policies. If a connection timeout occurs during the request, and the number of retries or the operation timeout is not exhausted, the client will retry the request.

## Profiling

Passing the debug switch(```-d```) will add some garbage collection stats to the reports.

By passing ```-profile``` switch, you con use Go's ```pprof``` tool to profile the benchmark.

Run the benchmark with ```-profile``` switch and then connect to it:

For 30 second CPU profile:

```$ go tool pprof http://localhost:6060/debug/pprof/profile```

For a heap profile:

```$ go tool pprof http://localhost:6060/debug/pprof/heap```

For goroutine contention and blocking profile:

```$ go tool pprof http://localhost:6060/debug/pprof/blocking```

Please refer to [/net/http/pprof](http://golang.org/pkg/net/http/pprof/) for more information.

To learn more about using pprof and profiling Go programs refer to [this canonical post on golang website](http://blog.golang.org/profiling-go-programs)

## Examples

To write 10,000,000 keys to the database (static bin data):

```$ ./benchmark -k 10000000```

To generate a load consisting 50% reads and 50% updates (static bin data):

```$ ./benchmark -k 10000000 -w RU,50```

To generate a load consisting 50% reads and 50% updates, using random bin data:

```$ ./benchmark -k 10000000 -w RU,50 -R```

To generate a load consisting 80% reads, using random bin data of strings 50 characters long:

```$ ./benchmark -k 10000000 -w RU,50 -R -o S:50```

To generate a load consisting 80% reads, using random bin data of strings 50 characters long, and set a timeout of 10ms:

```$ ./benchmark -k 10000000 -w RU,50 -R -o S:50 - T 50```
