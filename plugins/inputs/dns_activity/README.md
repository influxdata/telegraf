# DNS Activity Input Plugin

Uses a PCAP listener on a configured port to listen for any valid DNS answers.
 
 N.B. the telegraf binary should have permission to capture packets and (if required) enumerate the network devices on the host.
 
 If running telegraf as a non-privileged user, packet capture on Linux can be enabled using (as root) `setcap`:
 
     # setcap CAP_NET_RAW=ep /usr/bin/telegraf

## Plugin arguments:
- **port** int: The TCP and UDP port to capture packets on and inspect for DNS answers.
- **device** string: The name of the network device to capture on. Leave blank for all available devices (requires device enumeration permissions).

## Measurements:

Returns:
 - the number and cumulative size of DNS answer types broken down by DNS query type. Sizes are the on-the-wire size of the answer _contents_ - a typical A answer has a size of 4 bytes (1 IPv4 address). A single DNS message can contain multiple answers, this will return a count for each answer.
 - The number of DNS response messages broken down by error code. Each message has a single error status.
