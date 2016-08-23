# Ping input plugin

This input plugin will measures the round-trip

## Windows:
### Configration:
```
	## urls to ping
	urls = ["www.google.com"] # required
	
	## number of pings to send per collection (ping -n <COUNT>)
	count = 4 # required
	
	## Ping timeout, in seconds. 0 means default timeout (ping -w <TIMEOUT>)
	Timeout = 0
```
### Measurements & Fields:
- packets_transmitted ( from ping output )
- reply_received ( increasing only on valid metric from echo replay, eg. 'Destination net unreachable' reply will increment packets_received but not reply_received )
- packets_received ( from ping output )
- percent_reply_loss ( compute from packets_transmitted and reply_received )
- percent_packets_loss ( compute from packets_transmitted and packets_received )
- errors ( when host can not be found or wrong prameters is passed to application )
- response time
    - average_response_ms ( compute from minimum_response_ms and maximum_response_ms )
    - minimum_response_ms ( from ping output )
    - maximum_response_ms ( from ping output )
	
### Tags:
- server

### Example Output:
```
* Plugin: ping, Collection 1
ping,host=WIN-PBAPLP511R7,url=www.google.com average_response_ms=7i,maximum_response_ms=9i,minimum_response_ms=7i,packets_received=4i,packets_transmitted=4i,percent_packet_loss=0,percent_reply_loss=0,reply_received=4i 1469879119000000000
```