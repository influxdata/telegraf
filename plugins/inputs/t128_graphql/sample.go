package t128_graphql

var sampleConfig = `
## Collect data using graphQL
[[inputs.t128_graphql]]
## Required. The telegraf collector name
# collector_name = "arp-state"

## Required. GraphQL ports vary across 128T versions
# base_url = "http://localhost:31517/api/v1/graphql/"

## A socket to use for retrieving metrics - unused by default
# unix_socket = "/var/run/128technology/web-server.sock"

## The starting point in the graphQL tree for all configured tags and fields
# entry_point = "allRouters(name:'ComboEast')/nodes/nodes(name:'combo-east')/nodes/arp/nodes"

## Amount of time allowed to complete a single HTTP request
# timeout = "5s"

## Required. The fields to collect with the desired name as the key (left) and graphQL 
## key as the value (right)
# [inputs.t128_graphql.extract_fields]
#   state = "state"

## Required. The tags for filtering data with the desired name as the key (left) and 
## graphQL key as the value (right)
# [inputs.t128_graphql.extract_tags]
#   network-interface = "networkInterface"
#   device-interface = "deviceInterface"
#   vlan = "vlan"
#   ip-address = "ipAddress"
#   destination-mac = "destinationMac"
`
