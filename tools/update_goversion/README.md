# Update Go Version

The version doesn't require a leading "v" and minor versions don't need
a trailing ".0". The tool will still will work correctly if they are provided.

`go run tools/update_goversion/main.go 1.19.2`
`go run tools/update_goversion/main.go 1.19`

This tool is meant to be used to create a pull request that will update the
Telegraf project to use the latest version of Go.
The Dockerfile `quay.io/influxdb/telegraf-ci` used by the CI will have to be
pushed to the quay repository by a maintainer with `make ci`.
