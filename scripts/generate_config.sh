#!/bin/bash

os=$1 # windows or linux
artifact_dir="./build/dist" # Path is from .circleci/config.yml
extracted_path=$artifact_dir+"/extracted" # Path will contain telegraf binary

telegraf="telegraf"
config_name="telegraf.conf"

if [ "$os" = "windows" ]; then
    zip=$(find $artifact_dir -maxdepth 1 -name "*windows_amd64.zip" -print)
    unzip "$zip" -d $extracted_path
    telegraf="telegraf.exe"
    config_name="telegraf_windows.conf"
else
    tar=$(find $artifact_dir -maxdepth 1 -name "*linux_amd64.tar.gz" -print)
    tar -xf "$tar" -c $extracted_path || exit
fi

cd $extracted_path || exit
./$telegraf config > $config_name

mkdir ./new-config
mv $config_name ./new-config
