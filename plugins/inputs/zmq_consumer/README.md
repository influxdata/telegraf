# ZMQ SUB Consumer Input plugin

This input plugin implements a ZeroMQ SUB socket.

This plugin is currently used in an enterprise setting with >70k ingest per second,
but is provided as-is and should probably be considered beta until more users report on it.

## Known Issues

### libzmq library version

A common problem that occurs is when a libzmq version mismatch between the build environment
and the deployement environment. Even though `telegraf` is monolithic, the zmq go module makes
use of a .so import.

The problem manifests with a log error like follows:

```bash
2022-07-06T05:53:47Z E! [inputs.zmq_consumer::zmq] Error connecting to socket: zmq4 was compiled with ZeroMQ version 4.3.5, but the runtime links with version 4.3.1
```

### TODO: unit-tests are inadequate and will likely fail

The unit tests should make use of `ipc:///tmp/...` sockets that generate guaranteed unique (read:
guaranteed not to be already in use) endpoints that will subsequently be cleaned up on exit.

This requires setting up a test wide tempdir or per-test creation of socket endpoint.
