#!/bin/bash

token=$1

config_path="/new-config"

if [ ! -f "$config_path/telegraf.conf" ]; then
    echo "$config_path/telegraf.conf does not exist"
    exit
fi
if [ ! -f "$config_path/telegraf_windows.conf" ]; then
    echo "$config_path/telegraf_windows.conf does not exist"
    exit
fi

if cmp -s "$config_path/telegraf.conf" "etc/telegraf.conf" && cmp -s "$config_path/telegraf_windows.conf" "etc/telegraf_windows.conf"; then
    echo "Both telegraf.conf and telegraf_windows.conf haven't changed"
fi

# curl -H "Authorization: Bearer $token" -X POST "https://182c7jdgog.execute-api.us-east-1.amazonaws.com/prod/updateconfig"
