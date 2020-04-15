# Change Log

All notable changes to this project will be documented in this file.

## [Unreleased (`master`)][unreleased]

Major changes:

* Remove concept of `Client` it was confusing as it wasn't a direct standin for a Kinesis client.
* Rename `ScanError` to `ScanStatus` as it's not always an error.

Minor changes:

* Update tests to use Kinesis mock

## v0.2.0 - 2018-07-28

This is the last stable release from which there is a separate Client. It has caused confusion and will be removed going forward.

https://github.com/harlow/kinesis-consumer/releases/tag/v0.2.0

## v0.1.0 - 2017-11-20

This is the last stable release of the consumer which aggregated records in `batch` before calling the callback func.

https://github.com/harlow/kinesis-consumer/releases/tag/v0.1.0

[unreleased]: https://github.com/harlow/kinesis-consumer/compare/v0.2.0...HEAD
[options]: https://dave.cheney.net/2014/10/17/functional-options-for-friendly-apis
