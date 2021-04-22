# Profiling
This article describes how to collect performance traces and memory profiles
from Telegraf. If you are submitting this for an issue, please include the
version.txt generated below.

Use the `--pprof-addr` option to enable the profiler, the easiest way to do
this may be to add this line to `/etc/default/telegraf`:
```
TELEGRAF_OPTS="--pprof-addr localhost:6060"
```

Restart Telegraf to activate the profile address.

#### Trace Profile
Collect a trace during the time where the performance issue is occurring.  This
example collects a 10 second trace and runs for 10 seconds:
```
curl 'http://localhost:6060/debug/pprof/trace?seconds=10' > trace.bin
telegraf --version > version.txt
go env GOOS GOARCH >> version.txt
```

The `trace.bin` and `version.txt` files can be sent in for analysis or, if desired, you can
analyze the trace with:
```
go tool trace trace.bin
```

#### Memory Profile
Collect a heap memory profile:
```
curl 'http://localhost:6060/debug/pprof/heap' > mem.prof
telegraf --version > version.txt
go env GOOS GOARCH >> version.txt
```

Analyze:
```
$ go tool pprof mem.prof
(pprof) top5
```

#### CPU Profile
Collect a 30s CPU profile:
```
curl 'http://localhost:6060/debug/pprof/profile' > cpu.prof
telegraf --version > version.txt
go env GOOS GOARCH >> version.txt
```

Analyze:
```
go tool pprof cpu.prof
(pprof) top5
```
