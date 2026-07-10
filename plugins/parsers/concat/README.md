# Concat Parser Plugin

The `concat` parser splits a binary byte stream into messages by locating and
validating a message header, then passes each message to an embedded parser
(such as `json`, `csv` or `binary`) to produce metrics.

Use it for framed binary protocols where a header marks the start of each
message. Validating the header lets the parser find message boundaries reliably
even when reception starts mid-message or a message is spread across several
packets (e.g. UDP datagrams).

For each chunk of input the parser buffers the bytes and tries to read any
complete messages. If a header does not validate, the parser advances one byte
and retries. If the header validates but the message is incomplete, the parser
returns no metrics and keeps the buffer until the next chunk arrives.

> [!NOTE]
> The parser is stateful, keeping a buffer and scan cursor across calls. Run one
> instance per stream and keep `max_parallel_parsers = 1` (the default);
> parallel parsing corrupts its state.
> The parser is susceptible to dropping / creating nonexistent metrics depending
> on the chosen header format. It is likely that this bogus data will be caught
> by the embedded parser, but precautions should still be made.

## Configuration

```toml
[[inputs.socket_listener]]
  service_address = "udp://:9000"

  ## Data format to consume.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  data_format = "concat"

  ## The parser is stateful and must not run in parallel.
  max_parallel_parsers = 1

  ## Header schema used to locate and validate message boundaries.
  [inputs.socket_listener.split]
    ## Bytes that are constant in every header, as offset = "hex" pairs
    ## (a "0x"/"x" prefix is optional). Used to spot the start of a message.
    const_bytes = {0 = "0x50"}

    ## Number of header bytes to strip before passing a message to the
    ## embedded parser. If zero, the header is kept.
    header_length = 5

    ## Field holding the total message length in bytes (header included).
    ## If omitted, a message runs until the start of the next valid header.
    [inputs.socket_listener.split.message_length_field]
      offset = 2
      bytes = 2
      endianness = "be"  # "be" or "le"
      max_length = 0     # reject a header whose length exceeds this; 0 = no limit

    ## Checksum verifying a candidate header. A mismatch triggers
    ## resynchronization. Offsets and range are relative to the header start.
    [inputs.socket_listener.split.checksum]
      strategy = "xor"  # "none", "crc32", "md5", "sha256" or "xor"
      range = [0, 3]    # half-open [start, end) covered by the checksum
      offset = 4        # where the checksum is stored
      bytes = 1         # crc32=4, md5=16, sha256=32; xor is the fold width

  ## Embedded parser applied to each extracted message, selected by its own
  ## data_format and configured like any other parser.
  [inputs.socket_listener.embedded_parser]
    data_format = "json"
```

### message_length_field

An unsigned integer (width 1, 2, 4 or 8 bytes) giving the **total** message
length, header included. The parser consumes that many bytes and strips
`header_length` of them before parsing the remainder.

Set `max_length` to bound the accepted length. A header reporting a length
greater than `max_length` is treated as invalid and the parser resynchronizes.

Omit this field to run in delimiter mode, where a message ends at the start of
the next valid header. Delimiter mode needs `const_bytes` and/or a `checksum` so
headers can be recognized. In delimiter mode a message is only emitted once the
next one starts.

### checksum

The checksum is computed over the half-open `range` `[start, end)` and compared
against the `bytes`-wide value stored at `offset`. A mismatch is treated like a
failed header match, so the parser resynchronizes.

## Example

The sample configuration above describes a 5-byte header followed by a
JSON payload:

```text
 offset    0      1      2      3      4      5 ...
        +------+------+------+------+------+-------------------+
        | 0x50 | seq  |   length    | xor  |  payload (JSON)   |
        +------+------+------+------+------+-------------------+
         const             uint16    check
```

- `const_bytes = {0 = "0x50"}` — byte 0 is always `0x50`, marking a header.
- `message_length_field` at offset 2, 2 bytes big-endian — the `length` field
  holds the total frame size, so a `length` of `20` means 15 bytes of payload.
- `checksum` `xor` over `range = [0, 3]` stored at `offset = 4` — the parser
  XORs bytes 0-2 and compares the result against byte 4 to confirm the header.
- `header_length = 5` — the five header bytes are stripped before the `length`-5
  payload bytes are handed to the embedded JSON parser.

Byte 1 (`seq`) is not referenced by any option, so it is neither validated nor
stripped separately; it is simply part of the discarded header.

## Behavior

- Bytes skipped during resynchronization are dropped.
- In length mode, if a valid header appears before the length-implied end, the
  message is treated as truncated and discarded up to that header.
