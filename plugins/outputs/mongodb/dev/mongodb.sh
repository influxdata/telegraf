#!/bin/bash
# no auth
mongod --dbpath data/noauth --fork --logpath /var/log/mongodb_noauth/mongod.log --bind_ip 0.0.0.0 --port 27017

# scram auth
mongod --dbpath data/scram --fork --logpath /var/log/mongodb_scram/mongod.log --bind_ip 0.0.0.0 --port 27018
mongo localhost:27018/admin --eval "db.createUser({user:\"root\", pwd:\"changeme\", roles:[{role:\"root\",db:\"admin\"}]})"
mongo localhost:27018/admin --eval "db.shutdownServer()"
mongod --dbpath data/scram --fork --logpath /var/log/mongodb_scram/mongod.log --auth --setParameter authenticationMechanisms=SCRAM-SHA-256 --bind_ip 0.0.0.0 --port 27018

# get client certificate subject for creating x509 authenticating user
dn=$(openssl x509 -in ./private/client.pem -noout -subject -nameopt RFC2253 | sed 's/subject=//g')

# x509 auth
mongod --dbpath data/x509 --fork --logpath /var/log/mongodb_x509/mongod.log --bind_ip 0.0.0.0 --port 27019
mongo localhost:27019/admin --eval "db.getSiblingDB(\"\$external\").runCommand({createUser:\"$dn\",roles:[{role:\"root\",db:\"admin\"}]})"
mongo localhost:27019/admin --eval "db.shutdownServer()"
mongod --dbpath data/x509 --fork --logpath /var/log/mongodb_x509/mongod.log --auth --setParameter authenticationMechanisms=MONGODB-X509 --tlsMode preferTLS --tlsCAFile certs/cacert.pem --tlsCertificateKeyFile private/server.pem --bind_ip 0.0.0.0 --port 27019

# x509 auth short expirey
# mongodb will not start with an expired certificate. service must be started before certificate expires. tests should be run after certificate expiry
mongod --dbpath data/x509_expire --fork --logpath /var/log/mongodb_x509_expire/mongod.log --bind_ip 0.0.0.0 --port 27020
mongo localhost:27020/admin --eval "db.getSiblingDB(\"\$external\").runCommand({createUser:\"$dn\",roles:[{role:\"root\",db:\"admin\"}]})"
mongo localhost:27020/admin --eval "db.shutdownServer()"
mongod --dbpath data/x509_expire --fork --logpath /var/log/mongodb_x509_expire/mongod.log --auth --setParameter authenticationMechanisms=MONGODB-X509 --tlsMode preferTLS --tlsCAFile certs/cacert.pem --tlsCertificateKeyFile private/serverexp.pem --bind_ip 0.0.0.0 --port 27020

# note about key size and mongodb
# x509 must be 2048 bytes or stronger in order for mongodb to start. otherwise you will receive similar error below
# {"keyFile":"/opt/private/server.pem","error":"error:140AB18F:SSL routines:SSL_CTX_use_certificate:ee key too small"}

# copy key files to /opt/export. docker volume should point /opt/export to outputs/mongodb/dev in order to run non short x509 tests
cp /opt/certs/cacert.pem /opt/private/client.pem /opt/private/clientenc.pem /opt/export

while true; do sleep 1; done # leave container running.
