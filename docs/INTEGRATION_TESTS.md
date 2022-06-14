# Integration Tests

## Running

To run all named integration tests:

```shell
make test-integration
```

To run all tests, including unit and integration tests:

```shell
go test -count 1 -race ./...
```

## Developing

To run integration tests against a service the project uses
[testcontainers][1]. The makes it very easy to create and cleanup
container-based tests.

The `testutil/container.go` has a `Container` type that wraps this project to
easily create containers for testing in Telegraf. A typical test looks like
the following:

```go
servicePort := "5432"

container := testutil.Container{
    Image:        "postgres:alpine",
    ExposedPorts: []string{servicePort},
    Env: map[string]string{
        "POSTGRES_HOST_AUTH_METHOD": "trust",
    },
    WaitingFor: wait.ForAll(
        wait.ForLog("database system is ready to accept connections"),
        wait.ForListeningPort(nat.Port(servicePort)),
    ),
}

err := container.Start()
require.NoError(t, err, "failed to start container")

defer func() {
    require.NoError(t, container.Terminate(), "terminating container failed")
}()
```

User's should start the container and then defer termination of the container.

The `test.Container` type requires at least an image, ports to expose, and a
wait stanza. See the following to learn more:

### Images

Images are pulled from [DockerHub][2] by default. When looking for and
selecting an image from DockerHub, please use the following priority order:

1. [Official Images][3]: these images are generally produced by the publisher
  themselves and are fully supported with great documentation. These images are
  easy to spot as they do not have an author in the name (e.g. "mysql")
2. Publisher produced: not all software has an entry in the above Official
  Images. This may be due to the project being smaller or moving faster. In
  this case, pull directly from the publisher's DockerHub whenever possible.
3. [Bitnami][4]: If neither of the above images exist, look at the images
  produced and maintained by Bitnami. They go to great efforts to create images
  for the most popular software, produce great documentation, and ensure that
  images are maintained.
4. Other images: If, and only if, none of the above images will work for a
  particular use-case, then another image can be used. Be prepared to justify,
  the use of these types of images.

### Ports

When the port is specified as a single value (e.g. `11211`) then testcontainers
will generate a random port for the service to start on. This way multiple
tests can be run and prevent ports from conflicting.

The test container will expect an array of ports to expose for testing. For
most tests only a single port is used, but a user can specify more than one
to allow for testing if another port is open for example.

On each container's DockerHub page, the README will usually specify what ports
are used by the container by default. For many containers this port can be
changed or specified with an environment variable.

If no ports are specified, a user can view the image tag and view the various
image layers. Find an image layer with the `EXPOSE` keyword to determine what
ports are used by the container.

### Wait Stanza

The wait stanza lays out what test containers will wait for to determine that
the container has started and is ready for use by the test. It is best to
provide not only a port, but also a log message. Ports can come up very early
in the container, and the service may not be ready.

To find a good log message, it is suggested to launch the container manually
and see what the final message is printed. Usually this is something to the
effect of "ready for connections" or "setup complete". Also ensure that this
message only shows up once, or the use of the

### Other Parameters

There are other optional parameters that user can make use of for additional
configuration of the test containers:

- `BindMounts`: used to mount local test data into the container. The order is
  location in the container as the key and the local file as the value.
- `Entrypoint`: if a user wishes to override the entrypoint with a custom
  command
- `Env`: to pass environmental variables to the container similar to Docker
  CLI's `--env` option
- `Name`: if a container needs a hostname set or expects a certain name use
  this option to set the containers hostname
- `Networks`: if the user creates a custom network

[1]: <https://golang.testcontainers.org/> "testcontainers-go"
[2]: <https://hub.docker.com/> "DockerHub"
[3]: <https://hub.docker.com/search?q=&type=image&image_filter=official> "DockerHub Official Images"
[4]: <https://hub.docker.com/u/bitnami> "Bitnami Images"

## Network

By default the containers will use the bridge network where other containers
cannot talk to each other.

If a custom network is required for running tests, for example if containers
do need to communicate, then users can set that up with the following code:

```go
networkName := "test-network"
net, err := testcontainers.GenericNetwork(ctx, testcontainers.GenericNetworkRequest{
    NetworkRequest: testcontainers.NetworkRequest{
        Name:           networkName,
        Attachable:     true,
        CheckDuplicate: true,
    },
})
require.NoError(t, err)
defer func() {
    require.NoError(t, net.Remove(ctx), "terminating network failed")
}()
```

Then specify the network name in the container startup:

```go
zookeeper := testutil.Container{
    Image:        "wurstmeister/zookeeper",
    ExposedPorts: []string{"2181:2181"},
    Networks:     []string{networkName},
    WaitingFor:   wait.ForLog("binding to port"),
    Name:         "telegraf-test-zookeeper",
}
```

## Contributing

When adding integrations tests please do the following:

- Add integration to the end of the test name
- Use testcontainers when an external service is required
- Use the testutil.Container to setup and configure testcontainers
- Ensure the testcontainer wait stanza is well-tested
