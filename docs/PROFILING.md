# Telegraf profiling

Telegraf uses the standard package `net/http/pprof`. This package serves via its HTTP server runtime profiling data in the format expected by the pprof visualization tool.

By default, the profiling is turned off.

To enable profiling you need to specify address to config parameter `pprof-addr`, for example:

```shell
telegraf --config telegraf.conf --pprof-addr localhost:6060
```

There are several paths to get different profiling information:

To look at the heap profile:

`go tool pprof http://localhost:6060/debug/pprof/heap`

or to look at a 30-second CPU profile:

`go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30`

To view all available profiles, open `http://localhost:6060/debug/pprof/` in your browser.
