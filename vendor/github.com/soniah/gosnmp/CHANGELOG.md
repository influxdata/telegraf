## v1.22.0

* travis now failing build when goimports needs running
* gometalinter
* shell script for running local tests
* SNMPv3 - avoid crash when missing SecurityParameters
* add support for Walk and Get over TCP - RFC 3430
* SNMPv3 - allow input of private key instead of passphrase

## v1.21.0

* add netsnmp functionality "not check returned OIDs are increasing"

## v1.20.0

* convert all tags to correct semantic versioning, and remove old tags
* SNMPv1 trap IDs should be marshalInt32() not single byte
* use packetSecParams not sp secretKey in v3 isAuthentic()
* fix IPAddress marshalling in Set()

## v1.19.0

* bug fix: handle uninitialized v3 SecurityParameters in SnmpDecodePacket()
* SNMPError, Asn1BER - stringers; types on constants

## v1.18.0

* bug fix: use format flags - logPrintf() not logPrint()
* bug fix: parseObjectIdentifier() now returns []byte{0} rather than error
  when it receive zero length input
* use gomock
* start using go modules
* start a changelog
