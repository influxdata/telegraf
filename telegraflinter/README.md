# Private linter for Telegraf

The purpose of this linter is to enforce the review criteria for the Telegraf project, outlined here: https://github.com/influxdata/telegraf/wiki/Review. This is currently not compatible with the linter running in the CI and can only be ran locally.

## Running it locally

To use the Telegraf linter, you need a binary of golangci-lint that was compiled with CGO enabled. Currently no release is provided with it enabled, therefore you will need to clone the source code and compile it yourself. You can run the following commands to achieve this:

1. `git clone https://github.com/sspaink/golangci-lint.git`
2. `cd golangci-lint`
3. `git checkout tags/v1.39.0 -b 1390`
4. `CGO_ENABLED=true go build -o golangci-lint-cgo ./cmd/golangci-lint`

You will now have the binary you need to run the Telegraf linter. The Telegraf linter will now need to be compiled as a plugin to get a *.so file. [Currently plugins are only supported on Linux, FreeBSD, and macOS](https://golang.org/pkg/plugin/). From the root of the Telegraf project, you can run the following commands to compile the linter and run it:

1. `CGO_ENABLED=true go build -buildmode=plugin telegraflinter/telegraflinter.go`
2. In the .golanci-lint file:
    * uncomment the `custom` section under the `linters-settings` section
    * uncomment `telegraflinter` under the `enable` section 
3. `golanci-lint-cgo run`

*Note:* If you made a change to the telegraf linter and want to run it again, be sure to clear the [cache directory](https://golang.org/pkg/os/#UserCacheDir). On unix systems you can run `rm -rf ~/.cache/golangci-lint` otherwise it will seem like nothing changed.

## Requirement

This linter lives in the Telegraf repository and is compiled to become a Go plugin, any packages used in the linter *MUST* match the version in the golanci-lint otherwise there will be issues. For example the import `golang.org/x/tools v0.1.0` needs to match what golangci-lint is using.

## Useful references

* https://golangci-lint.run/contributing/new-linters/#how-to-add-a-private-linter-to-golangci-lint
* https://github.com/golangci/example-plugin-linter
