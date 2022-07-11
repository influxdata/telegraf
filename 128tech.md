# Contributions

This README will describe how to modify and build Telegraf for telegraf-128tech.

## Pulling in Upstream Changes

When updating an existing version, the tasks are straight forward.

1. Merge the upstream branch
2. Cherry-pick desired changes that exist on some other upstream branch (master for example)

## Moving to a New Upstream Version

When moving to a new upstream version, things are a little more complicated. It requires identification of what has been added to our custom telegraf version which must be pulled into the new release branch. This can be done as described in [this little article](https://til.hashrocket.com/posts/18139f4f20-list-different-commits-between-two-branches).

First, pull down the new upstream branch. Then, determine what's been added locally and needs to be included in the new custom build. Do this by finding the commits that were added in the custom branch. This example uses release 1.14, but that will change as time passes.

```
git log --no-merges --left-right --graph --cherry-pick --oneline release-1.14..release-128tech-1.14 | tail -r
```

That should provide a limited number of commits that will need to be cherry-picked from the original custom branch to the new one. It is possible these would already exist in the new upstream branch if they were back ported to the custom branch.

## Building a New RPM

Building a new RPM should be straight forward. The necessary building environments exist in the CI docker containers. There is a script `./scripts/docker-env` that wraps docker commands for easy use. To build an RPM from the current source code (example versioning used), simply run:

```
./scripts/docker-env build --version 1.13.1 --release 2
```

This will produce new RPMs and place them into the `build` directory.

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
go get -v -t -d ./...
go test -short ./...
```

or to run a single plugin's tests, run

```
go get -v -t -d ./...
go test ./plugins/outputs/http/
```
