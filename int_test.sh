#!/bin/bash

#exit script if any command/pipeline fails
set -e

#unset variables are errors
set -u

#return value of a pipeline is the last non-zero return
set -o pipefail

#print commands before they're executed
#set -x


file=telegraf.conf.int_test
port=7551

function log()
{
    echo ------- $*
}

log writing temp config file
tee $file <<EOF
[agent]
  interval = "2s"
  round_interval = true
  metric_batch_size = 1000
  metric_buffer_limit = 10000
  collection_jitter = "0s"
  flush_interval = "2s"
  flush_jitter = "0s"
  precision = ""
  hostname = ""
  omit_hostname = false
[[config.api]] 
  service_address = ":$port" 
[config.api.storage.internal] 
  file = "config_state.db"
EOF

log building telegraf
make telegraf

log running telegraf
./telegraf -config $file &
pid=$!

function cleanup()
{
    #log cleanup
    log send SIGKILL
    kill -9 $pid
    rm $file
    log failure
}

trap cleanup EXIT 

until ss -ltn "( sport = :$port )" | grep ":$port" > /dev/null; do
    log waiting for listening port
    sleep 2
done

#it doesn't accept connections for a while after it starts listening
#sleep 10

log rest create outputs.file
curl \
    -d '{"name": "outputs.file", "config": {"files": ["stdout"]} }' \
    -H "Content-Type: application/json" \
    -X POST \
    localhost:$port/plugins/create
echo

log rest create inputs.cpu
curl \
    -d '{"name": "inputs.cpu" }' \
    -H "Content-Type: application/json" \
    -X POST \
    localhost:$port/plugins/create
echo

function running()
{
    ps $pid > /dev/null
    local ret=$?
    #echo $ret
    return $ret
}

log make sure it''s still running
running

log send SIGINT
kill $pid

log wait for it to exit
sleep 2

log make sure it has exited
[ ! running ]

#unhook the cleanup trap
trap EXIT

#remove the temp config file
rm $file

log success
