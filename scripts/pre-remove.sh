#!/bin/bash

BIN_DIR=/usr/bin

# Distribution-specific logic
if [[ -f /etc/debian_version ]]; then
    # Debian/Ubuntu logic
    if [[ "$(readlink /proc/1/exe)" == */systemd ]]; then
        deb-systemd-invoke stop telegraf.service
    else
        # Assuming sysv
        invoke-rc.d telegraf stop
    fi
fi
