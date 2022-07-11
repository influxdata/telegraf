# Contributions

This README will describe how to modify and build Telegraf for telegraf-128tech.

## Fetching Dependencies

At some point, all of the Telegraf dependencies will need to be pulled down. This will be done during build automatically unless you've already done so and use the flag to skip that step (see "Building a New RPM"). _It's nice to do this manually because there's poor visibility into the build step_.

If you try to run commands without having the dependencies downloaded, you will see errors of the following form.

```
internal/internal.go:24:2: cannot find package "github.com/alecthomas/units" in any of:
        /usr/local/go/src/github.com/alecthomas/units (from $GOROOT)
        /go/src/github.com/alecthomas/units (from $GOPATH)
```

To fetch dependencies directly, you can do it simply from the shell. See "Using the Shell" for how to get into it. From the shell's default directory, simply run:

```
dep ensure --vendor-only -v
```

The above command provides the best visibility. The technically sanctioned fetch step is:

```
make deps
```

It does take some time to complete. After that, the dependencies exist in the `vendor` folder and don't need to be fetched again.

## Building a New RPM

Building a new RPM should be straight forward. The necessary building environments exist in the CI docker containers. There is a script `./scripts/docker-env` that wraps docker commands for easy use. To build an RPM from the current source code (example versioning used), simply run:

```
./scripts/docker-env build --version 1.13.1 --release 2
```

This will produce new RPMs and place them into the `build` directory.

Fetching can be skipped by using the `--no-fetch` flag:

```
./scripts/docker-env build --version 1.13.1 --release 3 --no-fetch
```

## Using the Shell

While not a comprehensive guide, this will get you started. You can drop into the docker environment by running:

```
./scripts/docker-env shell
```

From there, you can use `go` and the Telegraf `make` commands as desired. For a few examples, to run all the tests, simply run:

```
make test
```

or to run a single plugin's tests, run

```
go test ./plugins/outputs/http/
```
