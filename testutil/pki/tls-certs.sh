#!/bin/sh

mkdir certs certs_by_serial private &&
chmod 700 private &&
echo 01 > ./serial &&
touch ./index.txt &&
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
openssl req -x509 -config ./openssl.conf -days 3650 -newkey rsa:1024 -out ./certs/cacert.pem -keyout ./private/cakey.pem -subj "/CN=Telegraf Test CA/" -nodes &&

# Create server keypair
openssl genrsa -out ./private/serverkey.pem 1024 &&
openssl req -new -key ./private/serverkey.pem -out ./certs/servercsr.pem -outform PEM -subj "/CN=server.localdomain/O=server/" &&
openssl ca -config ./openssl.conf -in ./certs/servercsr.pem -out ./certs/servercert.pem -notext -batch -extensions server_ca_extensions &&

# Create client keypair
openssl genrsa -out ./private/clientkey.pem 1024 &&
openssl req -new -key ./private/clientkey.pem -out ./certs/clientcsr.pem -outform PEM -subj "/CN=client.localdomain/O=client/" &&
openssl ca -config ./openssl.conf -in ./certs/clientcsr.pem -out ./certs/clientcert.pem -notext -batch -extensions client_ca_extensions
