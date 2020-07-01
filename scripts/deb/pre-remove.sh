#!/bin/bash

BIN_DIR=/usr/bin

if [[ "$(readlink /proc/1/exe)" == */systemd ]]; then
	deb-systemd-invoke stop telegraf.service
else
	# Assuming sysv
	invoke-rc.d telegraf stop
fi
