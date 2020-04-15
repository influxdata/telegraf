# setup for working on traps

```
$ sudo aptitude -y install snmp-mibs-downloader snmp snmpd snmp-mibs-downloader
```

In the file `/etc/snmp/snmp.conf`
```
mibs +ALL
```

In the file `/etc/snmp/snmpd.conf`

```
comment out:
    agentAddress  udp:127.0.0.1:161

uncomment:
    agentAddress udp:161,udp6:[::1]:161

comment out:
    rocommunity public  default    -V systemonly

uncomment:
    rocommunity public 10.0.0.0/16

comment out:
    trapsink     localhost public

uncomment:
    trap2sink    localhost public
```

Create the file `~/.snmp/snmp.conf` with the contents:

```
# ~ expansion fails
persistentDir /home/sonia/.snmp_persist
```

```
$ sudo /etc/init.d/snmpd restart
```

# test

```
snmptrap -v 2c -c public 192.168.1.10 '' SNMPv2-MIB::system SNMPv2-MIB::sysDescr.0 s "red laptop" SNMPv2-MIB::sysServices.0 i "5" SNMPv2-MIB::sysObjectID o "1.3.6.1.4.1.2.3.4.5"
```

# tshark, wireshark

```
sudo aptitude -y install wireshark tshark
sudo dpkg-reconfigure wireshark-common # allow captures
sudo usermod -a -G wireshark sonia
sudo setcap cap_net_raw,cap_net_admin=eip /usr/bin/dumpcap
sudo getcap /usr/bin/dumpcap
# still 'Couldn't run /usr/bin/dumpcap in child process', so nuke it
sudo chmod 777 /usr/bin/dumpcap
```
Logout, login to apply wireshark and tshark permissions

In a second terminal, run:

```
tshark -i eth0 -f "port 161" -w trap.pcap
```

# snmptrap and MIBs

```
The TYPE is a single character, one of:
       i  INTEGER                   INTEGER
       u  UNSIGNED
       c  COUNTER32
       s  STRING                    DisplayString
       x  HEX STRING
       d  DECIMAL STRING
       n  NULLOBJ
       o  OBJID                     OBJECT IDENTIFIER
       t  TIMETICKS
       a  IPADDRESS
       b  BITS
```

# finding MIBs

Look in the file `/usr/share/mibs/ietf/SNMPv2-MIB`. Here are some
example lines:

```
line:77     sysDescr
line:88     sysObjectID
line:146    sysServices
```

For a gui MIB browser:

https://l3net.wordpress.com/2013/05/12/installing-net-snmp-mibs-on-ubuntu-and-debian/
