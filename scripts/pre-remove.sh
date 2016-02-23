#!/bin/bash

BIN_DIR=/usr/bin

# Distribution-specific logic
if [[ -f /etc/debian_version ]]; then
    # Debian/Ubuntu logic
    which systemctl &>/dev/null
    if [[ $? -eq 0 ]]; then
	deb-systemd-invoke stop telegraf.service
    else
	# Assuming sysv
	invoke-rc.d telegraf stop
    fi
fi
