# Frequently Asked Questions

### Q: Why do I get a "no such host" error resolving hostnames that other
programs can resolve?

Go uses a pure Go resolver by default for [name resolution](https://golang.org/pkg/net/#hdr-Name_Resolution).
This resolver behaves differently than the C library functions but is more
efficient when used with the Go runtime.

If you encounter problems or want to use more advanced name resolution methods
that are unsupported by the pure Go resolver, you can switch to the cgo
resolver.

If running manually set:
```
export GODEBUG=netdns=cgo
```

If running as a service add the environment variable to `/etc/default/telegraf`:
```
GODEBUG=netdns=cgo
```

### Q: When will the next version be released?

The latest release date estimate can be viewed on the
[milestones](https://github.com/influxdata/telegraf/milestones) page.
