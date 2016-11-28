#!/bin/bash

function disable_systemd {
    systemctl disable telegraf
    rm -f /lib/systemd/system/telegraf.service
}

function disable_update_rcd {
    update-rc.d -f telegraf remove
    rm -f /etc/init.d/telegraf
}

function disable_chkconfig {
    chkconfig --del telegraf
    rm -f /etc/init.d/telegraf
}

if [[ "$1" == "0" ]]; then
    # RHEL and any distribution that follow RHEL, Amazon Linux covered
    # telegraf is no longer installed, remove from init system
    rm -f /etc/default/telegraf

    which systemctl &>/dev/null
    if [[ $? -eq 0 ]]; then
        disable_systemd
    else
        # Assuming sysv
        disable_chkconfig
    fi
elif [ "$1" == "remove" -o "$1" == "purge" ]; then
    # Debian/Ubuntu logic
    # Remove/purge
    rm -f /etc/default/telegraf

    which systemctl &>/dev/null
    if [[ $? -eq 0 ]]; then
        disable_systemd
    else
        # Assuming sysv
        disable_update_rcd
    fi
fi
