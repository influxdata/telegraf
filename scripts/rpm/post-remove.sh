#!/bin/bash

# Telegraf is no longer installed, remove from systemd
if [[ "$1" = "0" ]]; then
    rm -f /etc/default/telegraf

    if [[ -d /run/systemd/system ]]; then
        systemctl disable telegraf
        rm -f /usr/lib/systemd/system/telegraf.service
    fi
fi

# Telegraf upgrade, restart service
if [[ $1 -ge 1 ]]; then
    if [[ -d /run/systemd/system ]]; then
        systemctl try-restart telegraf.service >/dev/null 2>&1 || :
    fi
fi
