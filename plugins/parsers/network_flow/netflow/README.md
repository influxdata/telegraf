# SFlow Parser

## Current Scope

Currently this Netflow Packet parser will  parse Netflow Version 10 (IPFIX) and as a by-product of the backwards compatability Netflow Version 9 as well.

## Schema

The field decoders are generated based on the IANA IPFIX element definitions: https://www.iana.org/assignments/ipfix/ipfix.xhtml.
The general priciple is that everything should be turned into a field appart from the elements listed in the following table. It is possible
to use Telegraf processor () to mask out certain default tags and fields and transfer particular elements between tags and fields as required

| ID | Name | 
|---|---|
|4|protocolIdentifier|  
|5|ipClassOfService|
|6|tcpControlBits|
|7|sourceTransportPort|
|8|sourceIPv4Address|
|9|sourceIPv4PrefixLength|
|10|ingressInterface|
|11|destinationTransportPort|
|12|destinationIPv4Address|
|13|destinationIPv4PrefixLength|
|14|egressInterface|
|16|bgpSourceAsNumber|
|17|bgpDestinationAsNumber|
|18|bgpNextHopIPv4Address|
|27|sourceIPv6Address|
|28|destinationIPv6Address|
|48|samplerId|
|61|flowDirection|
|70|mplsTopLabelStackSection|
|89|forwardingStatus|
|234|ingressVRFID|
|235|egressVRFID|

Not at elemeents are current decoded into fields or tags. At the moment anything but the following field/types are just ignored:
* unsigned8 -> int64 
* unsigned16 -> int64 
* unsigned32 -> int64 
* unsigned64 -> int64 
* ipv4Address -> IPV4 string 
* ipv6Address -> IPV6 string
* macAddress -> string
* octetArray -> hex string

