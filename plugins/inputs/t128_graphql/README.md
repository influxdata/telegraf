# 128T GraphQL Input Plugin

The graphql input plugin collects data from a 128T instance via graphQL.

### Configuration

```toml
# Collect data from a 128T instance using graphQL.
[[inputs.t128_graphql]]
## Required. A name for the collector which will be used as the measurement name of the produced data.
# collector_name = "peer-paths"

## Required. The base url for data collection.
## graphQL ports vary across 128T versions.
# base_url = "http://localhost:31517/api/v1/graphql/"

## A socket to use for retrieving data - unused by default
# unix_socket = "/var/run/128technology/web-server.sock"

## Required. The path to a point in the graphQL tree from which extract_fields and extract_tags will
## be specified. This path may contain (<key>:<value>) graphQL arguments such as
## (name:'RTR_EAST_COMBO').
# entry_point = "allRouters(name:'RTR_EAST_COMBO')/nodes/peers/nodes"

## Amount of time allowed to complete a single HTTP request
# timeout = "5s"

## Required. The fields to collect with the desired name as the key (left) and the graphQL 
## query path as the value (right). The path can be relative to the entry point or an absolute
## path that does not diverge from the entry-point and does not contain graphQL arguments such
## as (name:'RTR_EAST_COMBO').
# [inputs.t128_graphql.extract_fields]
#   is-active = "paths/isActive"
#   status = "paths/status"
#   other = "allRouters/nodes/other-field"  # absolute path

## The tags for filtering data with the desired name as the key (left) and the graphQL 
## query path as the value (right). The path can be relative to the entry point or an absolute
## path that does not diverge from the entry-point and does not contain graphQL arguments such
## as (name:'RTR_EAST_COMBO').
# [inputs.t128_graphql.extract_tags]
#   peer-name = "name"
#   device-interface = "paths/deviceInterface"
#   router-name = "allRouters/nodes/name"  # absolute path

```

### Example GraphQL Query
For the configuration above, the plugin will build the following graphQL query:

```
query {
  allRouters(name: "RTR_EAST_COMBO") {
    nodes {
      name
      other-field
      peers {
        nodes {
          name
          paths {
            isActive
            status
            deviceInterface
          }
        }
      }
    }
  }
}
```

### Example GraphQL Response
For the query above, an example graphQL response is:

```
{
  "data": {
    "allRouters": {
      "nodes": [
        {
          "name": "RTR_EAST_COMBO",
          "other-field": "foo",
          "peers": {
            "nodes": [
              {
                "paths": [
                  {
                    "isActive": true,
                    "status": "DOWN",
                    "deviceInterface": "10"
                  },
                  {
                    "isActive": true,
                    "status": "UP",
                    "deviceInterface": "11"
                  }
                ],
                "name": "fake"
              }
            ]
          }
        }
      ]
    }
  }
}
```

### Example Output
For the response above, the collector outputs:

```
peer-paths,router-name=RTR_EAST_COMBO,device-interface=10,peer-name=fake other="foo",is-active=true,status="DOWN" 1617285085000000000
peer-paths,router-name=RTR_EAST_COMBO,device-interface=11,peer-name=fake other="foo",is-active=true,status="UP" 1617285085000000000
```
