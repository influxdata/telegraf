#!/bin/bash -e

scripts_dir=$(cd $(dirname $0) && pwd)
assets_dir=${1}

echo "Generating certs into ${assets_dir}"

test ! `which certstrap` && go get -u -v github.com/square/certstrap

rm -f ${assets_dir}/*
# CA to distribute to loggregator certs
certstrap --depot-path ${assets_dir} init --passphrase '' --common-name loggregatorCA --expires "25 years"
certstrap --depot-path ${assets_dir} request-cert --passphrase '' --common-name telegraf
certstrap --depot-path ${assets_dir} sign telegraf --CA loggregatorCA --expires "25 years"
