# Integration Tests

To run our current integration test suite: 

Running the integration tests requires several docker containers to be
running.  You can start the containers with:
```
docker-compose up
```

To run only the integration tests use:

```
make test-integration
```

Use `make docker-kill` to stop the containers.

Contributing integration tests: 

- Add Integration to the end of the test name so it will be run with the above command.
- Writes tests where no library is being used in the plugin
- There is poor code coverage
- It has dynamic code that only gets run at runtime eg: SQL

Current areas we have integration tests: 

| Area                               | Info                                      |
|------------------------------------|-------------------------------------------|
| Inputs: Aerospike                  |                                           |
| Inputs: Disque                     |                                           |
| Inputs: Dovecot                    |                                           |                         
| Inputs: Mcrouter                   |                                           |                         
| Inputs: Memcached                  |                                           |                         
| Inputs: Mysql                      | Limited scope                             |                         
| Inputs: Opcua                      | Currently in a broken state               |                         
| Inputs: Openldap                   | Some are currently broken                 |                          
| Inputs: Pgbouncer                  | Currently in a broken state / only test   |                         
| Inputs: Postgresql                 |                                           |                         
| Inputs: Postgresql extensible      |                                           |                          
| Inputs: Procstat / Native windows  |                                           |                           
| Inputs: Prometheus                 |                                           |                          
| Inputs: Redis                      |                                           |                          
| Inputs: Sqlserver                  | Currently in a broken state               |                         
| Inputs: Win perf counters          |                                           |                          
| Inputs: Win services               |                                           |                          
| Inputs: Zookeeper                  |                                           |                          
| Processors: Ifname / SNMP          | Currently in a broken state               |                          
| Outputs: Cratedb / Postgres        | Currently in a broken state               |                          
| Outputs: Elasticsearch             |                                           |                          
| Outputs: Kafka                     |                                           |                          
| Outputs: MQTT                      | Is the only test                          |                          
| Outputs: Nats                      | Is the only test                          |                          
| Outputs: NSQ                       | Is the only test                          |                          
| Outputs: Opentsdb                  | Currently in a broken state               |                          
| Logger: Event logger               |                                           |                          

Areas we would benefit most from new integration tests:

| Area                               |
|------------------------------------|
|  SNMP                              |  
|  MYSQL                             |  
|  SQLSERVER                         |  
