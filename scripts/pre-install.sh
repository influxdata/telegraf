#!/bin/bash

if [[ -f /etc/opt/telegraf/telegraf.conf ]]; then
    # Legacy configuration found
    if [[ ! -d /etc/telegraf ]]; then
	# New configuration does not exist, move legacy configuration to new location
	echo -e "Please note, Telegraf's configuration is now located at '/etc/telegraf' (previously '/etc/opt/telegraf')."
	mv /etc/opt/telegraf /etc/telegraf

	backup_name="telegraf.conf.$(date +%s).backup"
	echo "A backup of your current configuration can be found at: /etc/telegraf/$backup_name"
	cp -a /etc/telegraf/telegraf.conf /etc/telegraf/$backup_name
    fi
fi
