# Appdynamics Output Plugin

This plugin writes to [Appdynamics Machine Agent](http://localhost:8293)
via raw TCP.

## Configuration:

```toml
  ## controller information to connect and retrieve tier-id value
  controllerTierURL = "https://foo.saas.appdynamics.com/controller/rest/applications/bar/tiers/baz?output=JSON"
  controllerUserName = "apiuser"
  controllerPassword = "apipass"
  ## Machine agent custom metrics listener url format string
  ## |Component:%d| gets transformed into |Component:id| during initialization - where 'id' is a tier-id for
  ## this controller application/tier combination
  agentURL = "http://localhost:8293/machineagent/metrics?name=Server|Component:%d|Custom+Metrics|"
```
