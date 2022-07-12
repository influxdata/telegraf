package t128_graphql

var sampleConfig = `
## Collect data from a 128T instance using graphQL.
[[inputs.t128_graphql]]
## Required. A name for the collector which will be used as the measurement name of the produced data.
# collector_name = "peer-paths"

## Required. The base url for data collection.
## graphQL ports vary across 128T versions.
# base_url = "http://localhost:31517/api/v1/graphql/"

## A socket to use for retrieving data - unused by default
# unix_socket = "/var/run/128technology/web-server.sock"

## Required. The path to a point in the graphQL tree from which extract_fields and extract_tags will
## be specified. This path may contain (<key>:<value>) graphQL arguments such as (name:'ComboEast').
# entry_point = "allRouters(name:'ComboEast')/nodes/peers/nodes"

## Amount of time allowed before the client cancels the HTTP request
# timeout = "5s"

## If false, the collector will continue querying when the graphQL server responds with 404 Not Found
# retry_if_not_found = false

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
`
