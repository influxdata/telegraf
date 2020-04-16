workflow "Test" {
  on = "push"
  resolves = ["go test"]
}

action "go test" {
  uses = "docker://golang:1.12-alpine"
  secrets = ["TEST_IOTHUB_SERVICE_CONNECTION_STRING", "TEST_EVENTHUB_CONNECTION_STRING"]
  env = {
    CGO_ENABLED = "0"
  }
  args = ["/bin/sh", "-c", "apk add --update git; go test ./..."]
}
