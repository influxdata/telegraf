#!/bin/bash

os=$1 # windows or linux
exe_path="/build/extracted" # Path will contain telegraf binary

telegraf="telegraf"
config_name="telegraf.conf"

if [ "$os" = "windows" ]; then
    zip=$(/bin/find ./build/dist -maxdepth 1 -name "*windows_amd64.zip" -print)
    exe_path="$PWD/build/extracted"
    unzip "$zip" -d "$exe_path"
    telegraf="telegraf.exe"
    config_name="telegraf_windows.conf"
    exe_path="$exe_path/telegraf*"
else
    tar=$(find /build/dist -maxdepth 1 -name "*linux_amd64.tar.gz" -print)
    tar -xfc "$tar" "$exe_path"
    exe_path="$exe_path/./telegraf*/usr/bin"
fi

shopt -s extglob
cd "$exe_path" || exit
ls
./exe_path/$telegraf config > $config_name

mkdir ./new-config
mv $config_name ./new-config
