#!/bin/bash

if [ -d /run/systemd/system ]; then
	if [ "$1" = remove ]; then
		deb-systemd-invoke stop telegraf.service
	fi
else
	# Assuming sysv
	invoke-rc.d telegraf stop
fi
