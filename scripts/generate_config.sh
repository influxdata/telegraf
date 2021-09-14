#!/bin/bash

os=$1 # windows or linux
exe_path="/build/extracted" # Path will contain telegraf binary
config_name="telegraf.conf"

if [ "$os" = "windows" ]; then
    zip=$(/bin/find ./build/dist -maxdepth 1 -name "*windows_amd64.zip" -print)
    exe_path="$PWD/build/extracted"
    unzip "$zip" -d "$exe_path"
    config_name="telegraf_windows.conf"
    exe_path=$(/bin/find "$exe_path" -name telegraf.exe -type f -print)
else
    tar_path=$(find /build/dist -maxdepth 1 -name "*linux_amd64.tar.gz" -print | grep -v ".*static.*")
    mkdir "$exe_path"
    tar --extract --file="$tar_path" --directory "$exe_path"
    exe_path=.$(find "$exe_path" -name telegraf -type f -print | grep ".*usr/bin/.*")
fi

"$exe_path" config > $config_name

mkdir ./new-config
mv $config_name ./new-config
