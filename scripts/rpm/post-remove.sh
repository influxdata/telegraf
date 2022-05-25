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

if [[ -f /etc/redhat-release ]] || [[ -f /etc/SuSE-release ]]; then
    # RHEL-variant logic
    if [[ "$1" = "0" ]]; then
        # InfluxDB is no longer installed, remove from init system
        rm -f /etc/default/telegraf

        if [[ "$(readlink /proc/1/exe)" == */systemd ]]; then
            disable_systemd /usr/lib/systemd/system/telegraf.service
        else
            # Assuming sysv
            disable_chkconfig
        fi
    fi
    if [[ $1 -ge 1 ]]; then
        # Package upgrade, not uninstall

        if [[ "$(readlink /proc/1/exe)" == */systemd ]]; then
            systemctl try-restart telegraf.service >/dev/null 2>&1 || :
        fi
    fi
elif [[ -f /etc/os-release ]]; then
    source /etc/os-release
    if [[ "$ID" = "amzn" ]] && [[ "$1" = "0" ]]; then
        # InfluxDB is no longer installed, remove from init system
        rm -f /etc/default/telegraf

        if [[ "$NAME" = "Amazon Linux" ]]; then
            # Amazon Linux 2+ logic
            disable_systemd /usr/lib/systemd/system/telegraf.service
        elif [[ "$NAME" = "Amazon Linux AMI" ]]; then
            # Amazon Linux logic
            disable_chkconfig
        fi
    elif [[ "$NAME" = "Solus" ]]; then
        rm -f /etc/default/telegraf
        disable_systemd /usr/lib/systemd/system/telegraf.service
    fi
fi
