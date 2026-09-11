# Context-Aware Output Connect/Write

## Objective

Give output plugins an opt-in way to receive a `context.Context` on
`Connect()` and `Write()` so a stuck call can be bounded by a configurable
`write_timeout` instead of hanging forever, and define the process for
migrating existing output plugins onto it.

## Keywords

outputs, agent, context, cancellation, write-timeout, shutdown

## Overview

`telegraf.Output` currently exposes `Connect() error` and
`Write(metrics []Metric) error`. Neither method takes a `context.Context`,
so once the agent calls into a plugin it has no way to unblock that call.
If the underlying client library performs a blocking network operation that
never returns — because a broker stopped acknowledging traffic, a TCP
connection is half-open, or a DNS/dial call hangs — the plugin's goroutine
is stuck indefinitely. The metric buffer fills and drops metrics, `Close()`
cannot run because `Write()` never returns, and the only recovery is killing
the process.

This is not hypothetical: issue [#19446][issue_19446] documents
`outputs.kafka`'s `SyncProducer.SendMessages` stuck for roughly 15 days in
production (Telegraf 1.39.2, sarama v1.60.0), with the agent unable to shut
down because `Agent.Run` was waiting on the flush worker.

Many client libraries expose their own timeout knobs, and those should be
set where available — they are the first line of defense, they are cheaper
than cancellation, and they let the client fail the operation cleanly rather
than having Telegraf abandon it. PR [#19447][pr_19447] does exactly that for
`outputs.kafka` by exposing sarama's dial/read/write and acknowledgement
timeouts.

That approach alone is not sufficient as a general answer:

* Coverage is inconsistent. Not every client library exposes a full set of
  timeouts, and some plugins wrap libraries that expose none.
* A client-level timeout often bounds one operation rather than the call
  Telegraf actually made. A produce/publish/commit call that retries
  internally can outlive every individual timeout it is composed of.
* A per-operation timeout is not a cancellation mechanism. It cannot be
  triggered by agent shutdown, so a plugin can still hold shutdown for the
  remainder of a long, legitimate operation.
* Users have no uniform, plugin-agnostic way to say "never let a write to
  this output run longer than N seconds".

A context-based mechanism gives Telegraf that uniform backstop regardless of
what the underlying client supports, and gives shutdown a way to reach into
a plugin that is mid-write. The two mechanisms are complementary: the client
timeout is what should normally fire; the context is what bounds the case
the client library did not anticipate.

## Design

### New optional interfaces

```go
// OutputWithContext is an output whose write operation can return on context cancellation.
type OutputWithContext interface {
    Output

    // WriteContext takes in group of points to be written to the Output.
    // It must return when the context is cancelled. The implementation is
    // responsible for any cleanup of an underlying operation that cannot be cancelled.
    WriteContext(ctx context.Context, metrics []Metric) error
}

// OutputWithConnectContext is an output whose connection attempt can be cancelled.
type OutputWithConnectContext interface {
    Output

    // ConnectContext performs any connection setup required for writing, like
    // Connect. It must return when the context is cancelled.
    ConnectContext(ctx context.Context) error
}
```

Both are separate, optional interfaces a plugin may implement in addition
to the existing `Connect()`/`Write()`. Plugins that implement neither
continue to work exactly as before — `RunningOutput` type-asserts for the
context-aware interfaces and falls back to the plain methods when absent.
This keeps the change fully backward compatible and lets plugins be
migrated one at a time.

A plugin implementing `WriteContext` does not need to also implement
`ConnectContext`, and vice versa, though most plugins that block on the
network in `Write` will have the same risk in `Connect`.

### Contract

A plugin's `WriteContext`/`ConnectContext` implementation must return
promptly once `ctx.Done()` fires. If the underlying client call cannot
itself be cancelled (no context support, no deadline option, blocking
syscall), the plugin must race it in a goroutine and return on
`ctx.Done()` without waiting for that goroutine, e.g.:

```go
done := make(chan error, 1)
go func() { done <- underlyingBlockingCall() }()
select {
case err := <-done:
    return err
case <-ctx.Done():
    return ctx.Err()
}
```

Abandoning the goroutine leaks it until the underlying call eventually
returns (or forever, in the pathological case). The plugin is responsible
for making that abandonment safe and bounded:

* Whatever the goroutine was using — connection, producer, session — must
  be discarded so a later `Connect`/`Write` cannot reuse a poisoned
  resource, and a replacement must be created instead.
* Cleanup of a discarded resource must be tied to that specific resource
  (a generation counter, or the pointer itself), so cleanup of an
  abandoned one cannot close its replacement.
* The number of abandoned resources in flight must be capped. Repeated
  timeouts must not grow goroutines or connections without bound; once the
  cap is reached, writes fail fast until cleanup completes.
* Reaching that cap must be logged, so a string of repeated timeouts is
  visible rather than silent.

`Close()` must remain safe to call after a cancelled `Write`/`Connect`,
including while an abandoned goroutine from the pattern above is still
running.

This narrows the existing `Output.Close` contract, which promises that
"Close will not be called until all writes have finished ... so locking is
not necessary". That promise still holds for plain `Output`
implementations. It cannot hold for an `OutputWithContext` implementation
that detaches an uncancellable operation, because the whole point of the
pattern is that `WriteContext` returns while the operation is still in
flight. Such implementations own the synchronization between `Close` and
their own abandoned goroutines, and the `Output` interface documentation in
`output.go` should be amended to say so, so that plugin authors do not rely
on a guarantee their own cancellation strategy has given up.

For a raw `net.Conn`, which has no context-aware read/write API, the way to
unblock an in-flight operation is to force its deadline to expire. Plugins
must not hand-roll that watcher; a shared helper in `internal` should
provide it, because three details are easy to get wrong and impossible to
notice in testing:

* It must operate on the exact connection handed to it, so it cannot race
  an error path that already closed the connection and installed a
  replacement.
* Its stop function must wait for the watcher goroutine to exit before
  returning, so no deadline can land after the caller has moved on.
* It must report whether the deadline was actually forced, including when
  cancellation raced a *successful* operation. In that case the connection
  is live but permanently poisoned, and the caller must discard it rather
  than hand it to the next flush.

### Delivery semantics

Cancelling a write does not cancel delivery. The receiving system may
already have accepted some or all of the batch, and in general the plugin
cannot know which.

A cancelled write is therefore reported as an error, and `RunningOutput`
keeps the batch for retry under the normal buffer behavior. This makes
`write_timeout` an at-least-once trade: it bounds how long a flush can
wedge the output loop, at the cost of possible duplicates when a slow but
healthy endpoint exceeds the timeout. That trade is why the option is off
by default; it must be documented as such in `docs/CONFIGURATION.md` and in
the README of every converted plugin.

Plugins must not attempt to hide this by silently dropping the batch on
cancellation.

### Interaction with `startup_error_behavior`

A `ConnectContext` that fails because the supplied context was cancelled —
agent shutdown, or an operator-initiated stop — is not a startup failure
and must not trigger [`startup_error_behavior`][tsd_006] handling. Treating
it as one would let a shutdown during startup flip a plugin into `ignore`
or `retry` handling for a condition that says nothing about the endpoint.

A deadline the agent itself imposed is different. If `write_timeout`
expired while the supplied parent context was still healthy, the endpoint
really did fail to connect in the configured time and the configured
`startup_error_behavior` applies as usual.

### Configuration: `write_timeout`

A new per-output option, `write_timeout` (duration, default: disabled), to
be added to `models.OutputConfig`. When set, the agent wraps each write
attempt in `context.WithTimeout(ctx, write_timeout)` before calling
`WriteContext`/`ConnectContext`.

What the deadline covers, precisely:

* **One flush, not one plugin call.** The timeout is created around
  `RunningOutput.WriteContext`, which may in turn hand several metric
  batches to the plugin's `WriteContext`. So `write_timeout = 30s` means
  "30 seconds for this whole flush", not "30 seconds per underlying plugin
  write". This is deliberate: the value users care about bounding is how
  long a flush can wedge the output loop.
* **Connect is bounded only for `OutputWithConnectContext`.** Each connect
  attempt gets its own fresh deadline, including the initial startup
  connection, so a plugin that hangs on first connect cannot block startup
  regardless of the retry backoff. An output that implements
  `OutputWithContext` but *not* `OutputWithConnectContext` gets bounded
  writes and unbounded connects; the two interfaces are independent and so
  is the protection they buy.
* **It does not replace client-native timeouts.** `write_timeout` is a
  backstop for the case the client library did not handle, layered on top
  of whatever timeouts that library exposes. A converted plugin should
  expose and set those as well.

`write_timeout` only has an effect on plugins implementing
`OutputWithContext`/`OutputWithConnectContext`. Setting it on a plugin that
has not been converted would otherwise be a silent no-op, leaving users
believing they are protected when they are not. `NewRunningOutput` should
therefore warn at config-load time when `write_timeout` is set on an output
implementing neither interface.

The warning deliberately fires only for "neither". A plugin implementing
just one of the two is a valid, documented state — the option really does
bound what that interface covers — so warning there would be noise.

### Final flush on shutdown

On agent shutdown the normal run context is already cancelled, which would
make any bounded final-flush attempt immediately fail for a context-aware
output — the exact plugins that gained the ability to shut down cleanly
would be the ones losing their last flush. The agent therefore
special-cases it: if the output is context-aware, it gets one real final
attempt on a fresh context, bounded either by the configured
`write_timeout` or by a default final write timeout (15s) when none is set.
Non-context-aware outputs keep receiving the already-cancelled shutdown
context, i.e. unchanged behavior.

## Plugin Requirements

Converting an existing output plugin to be context-aware means:

1. Implement `WriteContext(ctx, metrics) error`, threading `ctx` down to
   the underlying client call. Keep `Write(metrics) error` as
   `return p.WriteContext(context.Background(), metrics)` for callers that
   still use the plain interface (tests, embedding plugins, etc.).
2. If the client library's calls accept a `context.Context` natively
   (many modern Go clients do — HTTP-based outputs via
   `http.NewRequestWithContext`, gRPC, database drivers, cloud SDKs), pass
   it straight through. Prefer this over the goroutine-race pattern
   wherever available, since it avoids abandoning anything.
3. If the client library has no context support, use the goroutine-race
   pattern from the Contract section above, including the bounded,
   generation-checked abandonment it requires.
4. Implement `ConnectContext` the same way if `Connect()` can block on the
   network (dialing, auth handshake, metadata fetch, etc.).
5. Also expose and set any client-library-native timeout the plugin can
   (dial timeout, request timeout, acknowledgement timeout) — context
   cancellation is a backstop, not a replacement for those.
6. Add tests that assert `WriteContext`/`ConnectContext` return once the
   context is cancelled, without relying on wall-clock sleeps racing
   against a hang: inject a hook or fake client that blocks until
   signaled, cancel the context, and assert the call returns and the error
   wraps `ctx.Err()`.
7. Document `write_timeout` support, and the duplicate-on-cancellation
   trade, in the plugin's `README.md`.

### Plugins that should not be converted

Conversion is not automatically correct for every output. A plugin should
stay on the plain interface, with a comment in the plugin recording why,
when either:

* `Write`/`Connect` makes no outbound call that can hang — it writes to a
  local file or an in-memory structure, or it is a passive listener that
  serves data rather than sending it. There is nothing for `write_timeout`
  to usefully bound, and implementing the interface would advertise a
  protection that is not real.
* Its underlying client cannot safely be abandoned mid-call — for example
  a session that is not safe to close out from under an in-flight
  operation, or a non-thread-safe cgo binding. Racing it would trade a
  hang for memory corruption.

### Rollout process

* Each plugin conversion is its own PR against this spec — no bundled,
  repo-wide PR. This mirrors how `startup_error_behavior`
  ([TSD-006][tsd_006]) was rolled out incrementally.
* The core plumbing (the interfaces, `write_timeout`, the agent and
  `RunningOutput` changes, and the first converted plugin as a worked
  example) lands first, as its own PR.
* Prioritize plugins with a history of hang/stuck-write reports over a
  blanket pass through all output plugins; candidates for the next passes
  should be pulled from open issues describing a stuck or hung write.
* A plugin PR should be small: the `WriteContext`/`ConnectContext`
  addition, the goroutine-race wrapper only if the client needs it, tests,
  and a README note. It should not bundle unrelated refactors.

## Is/Is-not

**Is:**

* An opt-in pair of interfaces (`OutputWithContext`,
  `OutputWithConnectContext`) that plugins implement incrementally.
* A `write_timeout` config option that bounds a write/connect attempt for
  plugins that implement the interfaces.
* A per-plugin migration checklist and rollout process for converting the
  remaining output plugins over time, each as its own PR.

**Is-not:**

* Not a guarantee that every plugin can be interrupted immediately —
  plugins with non-cancellable underlying clients must use the
  goroutine-race-and-abandon pattern, which leaves a goroutine/resource
  outstanding until the underlying call returns. That is bounded and
  logged, not eliminated.
* Not exactly-once delivery: a cancelled write may already have been
  delivered and is retried, so `write_timeout` can produce duplicates.
* Not a replacement for a client library's own timeout configuration;
  plugins should set both where available.
* Not a change to `inputs`, `processors`, or `aggregators`. `Gather()` has
  the same class of risk, and the same approach should work there, but it
  is a separate spec — the plugin count, the existing interval semantics
  and the "what happens to a partial gather" question are different enough
  to deserve their own discussion rather than being folded in here.
* Not a single repo-wide conversion PR — plugins are converted
  incrementally, each independently reviewable.

## Prior art

* Issue [#19446][issue_19446] — `outputs.kafka` `SendMessages` stuck for
  ~15 days in production; sanitized incident evidence.
* PR [#19447][pr_19447] — the client-library half of the same problem:
  exposes sarama's dial, read, write and acknowledgement timeouts for
  `outputs.kafka`.
* [TSD-006][tsd_006] (`startup_error_behavior`) — precedent for defining a
  shared behavior/config option once and rolling it out per-plugin rather
  than in one PR.
* [TSD-008][tsd_008] (partial write error handling) — precedent for
  defining how outputs report a write that partially succeeded, which is
  the same class of question as a cancelled write's ambiguous outcome.
* Previously closed issues describing the same class of problem in
  `outputs.kafka`: [#11427](https://github.com/influxdata/telegraf/issues/11427),
  [#8349](https://github.com/influxdata/telegraf/issues/8349).
* A reference implementation of the interfaces, the agent plumbing and a
  first pass over the applicable output plugins exists at
  <https://github.com/FireBurn/telegraf/tree/output-context-conversion>,
  to be split into the per-PR rollout described above once this spec is
  agreed.

[issue_19446]: https://github.com/influxdata/telegraf/issues/19446
[pr_19447]: https://github.com/influxdata/telegraf/pull/19447
[tsd_006]: /docs/specs/tsd-006-startup-error-behavior.md
[tsd_008]: /docs/specs/tsd-008-partial-write-error-handling.md

## Open questions

* Should `write_timeout` set on an output implementing neither context
  interface be a warning or a startup error? The proposal above is a
  warning: conversions land over time, and a shared config template
  applied across many outputs should not hard-fail because one of them has
  not been converted yet. An error is the stronger guarantee that nobody
  believes they are protected when they are not. Needs maintainer
  sign-off.
* Is there an agreed list or priority order of output plugins to convert
  first, or does it stay purely issue-driven?
