# Update Go Version

`go run tools/update_goversion/main.go $LATEST_GO_VERSION`

This tool is meant to be used to create a pull request that will update the
Telegraf project to use the latest version of Go.
The Dockerfile `quay.io/influxdb/telegraf-ci` used by the CI will have to be
pushed to the quay repository by a maintainer.
