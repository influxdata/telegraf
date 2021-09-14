#!/bin/bash

os=$1 # windows or linux
extracted_path="/build/extracted" # Path will contain telegraf binary

telegraf="telegraf"
config_name="telegraf.conf"

if [ "$os" = "windows" ]; then
    zip=$(/bin/find ./build/dist -maxdepth 1 -name "*windows_amd64.zip" -print)
    extracted_path="./build/extracted"
    unzip "$zip" -d $extracted_path
    telegraf="telegraf.exe"
    config_name="telegraf_windows.conf"
    extracted_path="$extracted_path/telegraf*"
else
    tar=$(find /build/dist -maxdepth 1 -name "*linux_amd64.tar.gz" -print)
    tar -xf "$tar" -c "$extracted_path"
    extracted_path="$extracted_path./telegraf*/usr/bin"
fi

cd "$extracted_path" || exit
ls
./exe_path/$telegraf config > $config_name

mkdir ./new-config
mv $config_name ./new-config
