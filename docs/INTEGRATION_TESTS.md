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

The `servicePort` is the port the service is running on and Telegraf will
connect to. When the port is specified as a single value (e.g. `11211`) then
testcontainers will generate a random port for the service to start on. This
way multiple tests can be run and prevent ports from conflicting.

The `test.Container` type requires at least the following three items:

1. An image name from [DockerHub][2], which can include a specific tag or not
2. An array of port(s) to expose to the test to connect to
3. A wait stanza. This lays out what testcontainers will wait for to determine
  that the container has started and is ready for use by the test. It is best
  to provide not only a port, but also a log message. Ports can come up very
  early in the container, and the service may not be ready.

There are other optional parameters like `Env` to pass environmental variables
to the container or `BindMounts` to pass test data into the container as well.

User's should start the container and then defer termination of the container.

[1]: <https://golang.testcontainers.org/> "testcontainers-go"
[2]: <https://hub.docker.com/> "DockerHub"

## Contributing

When adding integrations tests please do the following:

- Add integration to the end of the test name
- Use testcontainers when an external service is required
- Use the testutil.Container to setup and configure testcontainers
- Ensure the testcontainer wait stanza is well-tested
