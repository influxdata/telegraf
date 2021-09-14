#!/bin/bash

token=$1

config_path="/new-config"

if [ ! -f "$config_path/telegraf.conf" ]; then
    echo "$config_path/telegraf.conf does not exist"
fi
if [ ! -f "$config_path/telegraf_windows.conf" ]; then
    echo "$config_path/telegraf.conf does not exist"
fi

# curl -H "Authorization: Bearer $token" -X POST "https://182c7jdgog.execute-api.us-east-1.amazonaws.com/prod/updateconfig"
