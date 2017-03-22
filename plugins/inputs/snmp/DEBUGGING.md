# Debugging & Testing SNMP Issues

### Install net-snmp on your system:

Mac:

```
brew install net-snmp
```

### Run an SNMP simulator docker image to get a full MIB on port 161:

```
docker run -d -p 161:161/udp xeemetric/snmp-simulator
```

### snmpget:

snmpget corresponds to the inputs.snmp.field configuration.

```bash
$ # get an snmp field with fully-qualified MIB name.
$ snmpget -v2c -c public localhost:161 system.sysUpTime.0
DISMAN-EVENT-MIB::sysUpTimeInstance = Timeticks: (1643) 0:00:16.43

$ # get an snmp field, outputting the numeric OID.
$ snmpget -On -v2c -c public localhost:161 system.sysUpTime.0
.1.3.6.1.2.1.1.3.0 = Timeticks: (1638) 0:00:16.38
```

### snmptranslate:

snmptranslate can be used to translate an OID to a MIB name:

```bash
$ snmptranslate .1.3.6.1.2.1.1.3.0
DISMAN-EVENT-MIB::sysUpTimeInstance
```

And to convert a partial MIB name to a fully qualified one:

```bash
$ snmptranslate -IR sysUpTime.0
DISMAN-EVENT-MIB::sysUpTimeInstance
```

And to convert a MIB name to an OID:

```bash
$ snmptranslate -On -IR system.sysUpTime.0
.1.3.6.1.2.1.1.3.0
```

