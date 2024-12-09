# Profiling

Telegraf uses the standard package `net/http/pprof`. This package serves via
its HTTP server runtime profiling data in the format expected by the pprof
visualization tool.

## Enable profiling

By default, the profiling is turned off. To enable profiling users need to
specify the pprof address config parameter `pprof-addr`. For example:

```shell
telegraf --config telegraf.conf --pprof-addr localhost:6060
```

## Profiles

To view all available profiles, open the URL specified in a browser. For
example, open `http://localhost:6060/debug/pprof/` in your browser.

To look at the heap profile:

```shell
go tool pprof http://localhost:6060/debug/pprof/heap
```

To look at a 30-second CPU profile:

```shell
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30
```

## Generate heap image

It is very helpful to generate an image to visualize what heap memory is used.
It is best to capture an image a few moments after Telegraf starts and then at
additional periods (e.g. 1min, 5min, etc.).

A user can capture the image with Go via:

```shell
go tool pprof -png http://localhost:6060/debug/pprof/heap > heap.png
```

The resulting image can be uploaded to a bug report.

## References

For additional information on pprof see the following:

* [net/http/pprof][]
* [Julia Evans: Profiling Go programs with pprof][]
* [Debugging Go Code][]

[net/http/pprof]: https://pkg.go.dev/net/http/pprof
[julia evans: profiling go programs with pprof]: https://jvns.ca/blog/2017/09/24/profiling-go-with-pprof/
[Debugging Go Code]: https://www.infoq.com/articles/debugging-go-programs-pprof-trace/
