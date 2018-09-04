## collectd plugins in Go

## About

This is _experimental_ code to write _collectd_ plugins in Go. It requires Go
1.5 or later and a recent version of the collectd sources to build.

## Build

To set up your build environment, set the `CGO_CPPFLAGS` environment variable
so that _cgo_ can find the required header files:

    export COLLECTD_SRC="/path/to/collectd"
    export CGO_CPPFLAGS="-I${COLLECTD_SRC}/src/daemon -I${COLLECTD_SRC}/src"

You can then compile your Go plugins with:

    go build -buildmode=c-shared -o example.so

More information is available in the documentation of the `collectd.org/plugin`
package.

    godoc collectd.org/plugin

## Future

Only *read* and *write* callbacks are currently supported. Based on these
implementations it should be fairly straightforward to implement the remaining
callbacks. The *init*, *shutdown*, *log*, *flush* and *missing* callbacks are
all likely low-hanging fruit. The *notification* callback is a bit trickier
because it requires implementing notifications in the `collectd.org/api` package
and the (un)marshaling of `notification_t`. The (complex) *config* callback is
arguably the most important but, unfortunately, also the most complex to
implemented.

If you're willing to give any of this a shot, please ping @octo to avoid
duplicate work.
