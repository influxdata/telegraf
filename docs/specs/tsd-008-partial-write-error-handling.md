# Partial write error handling

## Objective

Provide a way to pass information about partial metric write errors from an
output to the output model.

## Keywords

output plugins, write, error, output model, metric, buffer

## Overview

When output plugins serialize metric and/or send them to the service endpoint,
single metrics might cause errors e.g. by not being serializable or by being
rejected by the service on the server side.
Currently, an output is only able to accept or reject the complete batch of
metrics it receives from the output model. This causes issues if only a subset
of metrics in the batch fails as the output has no way of telling the model
(and in turn the output buffer) which metrics failed but can only accept or
reject the whole batch. As a consequence, outputs need to "accept" the batch
to avoid a requeueing of the batch for the next flush interval. This distorts
statistics of accepted metrics and causes misleading log messages.
Even worse, for outputs ending-up with partial writes, e.g. only the first half
of the metrics can be written to the service, there is no way of only accepting
the written metrics so they need to internally buffer the remaining ones.

This specification aims at defining the handling of partially successful writes
and introduces the concept of a special _write error_ type. That error type
must reflect partial writes and partial serialization to overcome the
aforementioned issues.

To do so, the error must contain a list of successfully
written metrics, which must be marked as __accepted__ and must be removed from
the buffer. The error must contain a list of metrics fatally failed to be
written or serialized and cannot be retried, which must be marked as
__rejected__ and must be removed from the buffer.

The error may contain a list of metrics not-yet written to be __kept__ for the
next write cylce. Those metrics must not be marked and must be kept in the
buffer. If the error does not contain the list, the list must be inferred using
the accept and reject lists and the metrics in the batch.

All metric lists should be communicated as indices into the batch to be able
to handle tracking metrics correctly.

For backward compatibility and simplicity output plugins can return a `nil`
error to indicate that __all__ metrics of the batch are __accepted__. Similarly,
returing an error _not_ being a _write error_ indicates that __all__ metrics of
the batch should be __kept__ in the buffer for the next write cycle.

## Related Issues

- [issue #11942](https://github.com/influxdata/telegraf/issues/11942) for
  contradicting log messages
- [issue #14802](https://github.com/influxdata/telegraf/issues/14802) for
  rate-limiting multiple batch sends
- [issue #15908](https://github.com/influxdata/telegraf/issues/15908) for
  infinite loop if single metrics cannot be written
