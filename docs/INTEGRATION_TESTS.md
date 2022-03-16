# Integration Tests

To run our current integration test suite:

Running the integration tests requires several docker containers to be
running.  You can start the containers with:

```shell
docker-compose up
```

To run only the integration tests use:

```shell
make test-integration
```

Use `make docker-kill` to stop the containers.

Contributing integration tests:

- Add Integration to the end of the test name so it will be run with the above command.
- Writes tests where no library is being used in the plugin
- There is poor code coverage
- It has dynamic code that only gets run at runtime eg: SQL

Current areas we have integration tests:

| Area                               | What it does                              |
|------------------------------------|-------------------------------------------|
| Inputs: Aerospike                  |                                           |
| Inputs: Disque                     |                                           |
| Inputs: Dovecot                    |                                           |
| Inputs: Mcrouter                   |                                           |
| Inputs: Memcached                  |                                           |
| Inputs: Mysql                      |                                           |
| Inputs: Opcua                      |                                           |
| Inputs: Openldap                   |                                           |
| Inputs: Pgbouncer                  |                                           |
| Inputs: Postgresql                 |                                           |
| Inputs: Postgresql extensible      |                                           |
| Inputs: Procstat / Native windows  |                                           |
| Inputs: Prometheus                 |                                           |
| Inputs: Redis                      |                                           |
| Inputs: Sqlserver                  |                                           |
| Inputs: Win perf counters          |                                           |
| Inputs: Win services               |                                           |
| Inputs: Zookeeper                  |                                           |
| Outputs: Cratedb / Postgres        |                                           |
| Outputs: Elasticsearch             |                                           |
| Outputs: Kafka                     |                                           |
| Outputs: MQTT                      |                                           |
| Outputs: Nats                      |                                           |
| Outputs: NSQ                       |                                           |

Areas we would benefit most from new integration tests:

| Area                               |
|------------------------------------|
|  SNMP                              |
|  MYSQL                             |
|  SQLSERVER                         |
