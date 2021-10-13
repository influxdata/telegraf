#!/bin/bash
# This script is responsible for triggering the Tiger Bot endpoint that will create the pull request with the newly generated configs.
# This script is meant to be only ran in within the Circle CI pipeline.

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

curl -H "Authorization: Bearer $token" -X POST "https://182c7jdgog.execute-api.us-east-1.amazonaws.com/prod/updateConfig"
