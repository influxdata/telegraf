# Example shows how to load custom plugins into starlark scripts.
# Plugins should been build with `go build -buildmode=plugin`.
# The plugin should contain a `InitModule` function which expects a
# `telegraf.Logger` and returns a `*starlarkstruct.Module`
# See `../example_custom_module/custom.go` for an example.
#
# Only supported for linux, freebsd and macos. See https://pkg.go.dev/plugin
#
# Example Input:
# custom message="invalid" 1515581000000000000
#
# Example Output:
# custom message="Hallo from custom module" 1515581000000000000
# 

load("example_custom_module/custom.star", "custom")

def apply(metric):
  message = custom.test()
  metric.fields["message"] = message

  return metric