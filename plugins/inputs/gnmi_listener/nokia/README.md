# Nokia dial-out telemetry

This package implements the GRPC server for Nokia devices to support dial-out
telemetry available in Nokia SR OS (and potentially other) devices.

The proto-buffer file in this directory is taken from the
[Nokia Github repository][nokia_repo]. The protocol-buffer file is subject to
the license below.

To update protocol-buffer file or to regenerate the Go code please run

```sh
go generate
```

to update the Go code.

[nokia_repo]: https://github.com/nokia/7x50_protobufs

## License

Copyright 2020 Nokia. All rights reserved. Reproduction of this document is
authorized on the condition that the foregoing copyright notice is included.

The protobufs embody Nokia's proprietary intellectual property. Nokia
retains all title and ownership in the specification, including any
revisions.

Nokia grants all interested parties a non-exclusive license to use and
distribute an unmodified copy of this specification in connection with
management of Nokia products, and without fee, provided this copyright
notice and license appear on all copies.

This specification is supplied 'as is', and Nokia makes no warranty, either
express or implied, as to the use, operation, condition, or performance
of the specification.
