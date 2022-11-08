#!/bin/bash
# This script is responsible for checking if the Telegraf config has been updated
# This script is meant to be only ran in within the Circle CI pipeline so that the Tiger Bot can update them automatically.
# It supports Windows and Linux because the configs are different depending on the OS.


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
    exe_path=$(find "$exe_path" -name telegraf -type f -print | grep ".*usr/bin/.*")
fi

$exe_path config > $config_name

DIFF=$(diff -q $config_name etc/$config_name)

if [ "$DIFF" ]
then
    mkdir ./new-config
    mv $config_name ./new-config
    echo "The sample configuration has been updated due to this pull request."
    echo "You can get the new configs here: https://output.circle-artifacts.com/output/job/${CIRCLE_WORKFLOW_JOB_ID}/artifacts/${CIRCLE_NODE_INDEX}/new-config/${config_name}"
    echo "Update the sample configurations found under 'etc/' with these new ones in your pull request for this step to pass."
    exit 1
fi
