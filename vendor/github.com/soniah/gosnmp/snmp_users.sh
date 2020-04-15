#!/bin/bash

cat << EOF >> /etc/snmp/snmpd.conf
createUser noAuthNoPrivUser
createUser authMD5OnlyUser  MD5 testingpass0123456789
createUser authSHAOnlyUser  SHA testingpass9876543210
createUser authMD5PrivDESUser MD5 testingpass9876543210 DES
createUser authSHAPrivDESUser SHA testingpassabc6543210 DES
createUser authMD5PrivAESUser MD5 AEStestingpass9876543210 AES
createUser authSHAPrivAESUser SHA AEStestingpassabc6543210 AES

rouser   noAuthNoPrivUser noauth
rouser   authMD5OnlyUser auth
rouser   authSHAOnlyUser auth
rouser   authMD5PrivDESUser authPriv
rouser   authSHAPrivDESUser authPriv
rouser   authMD5PrivAESUser authPriv
rouser   authSHAPrivAESUser authPriv
EOF

# enable ipv6 TODO restart fails - need to enable ipv6 on interface; spin up a Linux instance to check this
# sed -i -e '/agentAddress/ s/^/#/' -e '/agentAddress/ s/^##//' /etc/snmp/snmpd.conf
