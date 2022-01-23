#!/bin/bash

function disable_systemd {
    systemctl disable telegraf
    rm -f $1
}

function disable_update_rcd {
    update-rc.d -f telegraf remove
    rm -f /etc/init.d/telegraf
}

function disable_chkconfig {
    chkconfig --del telegraf
    rm -f /etc/init.d/telegraf
}

if [ "$1" == "remove" -o "$1" == "purge" ]; then
	# Remove/purge
	rm -f /etc/default/telegraf

	if [[ "$(readlink /proc/1/exe)" == */systemd ]]; then
		disable_systemd /lib/systemd/system/telegraf.service
	else
		# Assuming sysv
		# Run update-rc.d or fallback to chkconfig if not available
		if which update-rc.d &>/dev/null; then
			disable_update_rcd
		else
			disable_chkconfig
		fi
	fi
fi
