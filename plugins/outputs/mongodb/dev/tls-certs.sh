#!/bin/sh

mkdir certs certs_by_serial private &&
chmod 700 private &&
echo 01 > ./serial &&
touch ./index.txt &&
echo 'unique_subject = no' > index.txt.attr
cat >./openssl.conf <<EOF
[ ca ]
default_ca = telegraf_ca

[ telegraf_ca ]
certificate = ./certs/cacert.pem
database = ./index.txt
new_certs_dir = ./certs_by_serial
private_key = ./private/cakey.pem
serial = ./serial

default_crl_days = 3650
default_days = 3650
default_md = sha256

policy = telegraf_ca_policy
x509_extensions = certificate_extensions

[ telegraf_ca_policy ]
commonName = supplied

[ certificate_extensions ]
basicConstraints = CA:false

[ req ]
default_bits = 1024
default_keyfile = ./private/cakey.pem
default_md = sha256
prompt = yes
distinguished_name = root_ca_distinguished_name
x509_extensions = root_ca_extensions

[ root_ca_distinguished_name ]
commonName = hostname

[ root_ca_extensions ]
basicConstraints = CA:true
keyUsage = keyCertSign, cRLSign

[ client_ca_extensions ]
basicConstraints = CA:false
keyUsage = digitalSignature
subjectAltName = @client_alt_names
extendedKeyUsage = 1.3.6.1.5.5.7.3.2

[ client_alt_names ]
DNS.1 = localhost
IP.1 = 127.0.0.1

[ server_ca_extensions ]
basicConstraints = CA:false
subjectAltName = @server_alt_names
keyUsage = keyEncipherment, digitalSignature
extendedKeyUsage = 1.3.6.1.5.5.7.3.1

[ server_alt_names ]
DNS.1 = localhost
IP.1 = 127.0.0.1
EOF
openssl req -x509 -config ./openssl.conf -days 3650 -newkey rsa:2048 -out ./certs/cacert.pem -keyout ./private/cakey.pem -subj "/CN=Telegraf Test CA/" -nodes &&

# Create server and soon to expire keypair
openssl genrsa -out ./private/serverkey.pem 2048 &&
openssl req -new -key ./private/serverkey.pem -out ./certs/servercsr.pem -outform PEM -subj "/CN=$(cat /proc/sys/kernel/hostname)/O=server/" &&
openssl ca -config ./openssl.conf -in ./certs/servercsr.pem -out ./certs/servercert.pem -notext -batch -extensions server_ca_extensions &&
openssl ca -config ./openssl.conf -in ./certs/servercsr.pem -out ./certs/servercertexp.pem -startdate $(date +%y%m%d%H%M00 --date='-5 minutes')'Z' -enddate $(date +%y%m%d%H%M00 --date='5 minutes')'Z' -notext -batch -extensions server_ca_extensions &&

# Create client and client encrypted keypair
openssl genrsa -out ./private/clientkey.pem 2048 &&
openssl req -new -key ./private/clientkey.pem -out ./certs/clientcsr.pem -outform PEM -subj "/CN=$(cat /proc/sys/kernel/hostname)/O=client/" &&
openssl ca -config ./openssl.conf -in ./certs/clientcsr.pem -out ./certs/clientcert.pem -notext -batch -extensions client_ca_extensions &&
cp ./private/clientkey.pem ./private/clientkeyenc.pem &&
ssh-keygen -p -f ./private/clientkeyenc.pem -m PEM -N 'changeme'

# Combine crt and key to create pem formatted keyfile
cat ./certs/clientcert.pem ./private/clientkey.pem > ./private/client.pem &&
cat ./certs/clientcert.pem ./private/clientkeyenc.pem > ./private/clientenc.pem &&
cat ./certs/servercert.pem ./private/serverkey.pem > ./private/server.pem &&
cat ./certs/servercertexp.pem ./private/serverkey.pem > ./private/serverexp.pem
