#!/usr/bin/env bash

TRUE="0"
SYSTEMD_INUSE=`systemctl -a --no-pager > /dev/null;  echo $?`
DHCLIENT_HOOKS_FILE=/etc/dhcp/dhclient-up-hooks

if [[ "$SYSTEMD_INUSE" != "$TRUE" ]]; then
	exit 0
fi

# Clean up only when rpm is uninstalled not during upgrade
if [[ "$1" == "0" ]]; then 
    systemctl stop telegraf
    systemctl disable telegraf
    [[ -f ${DHCLIENT_HOOKS_FILE} ]] && sed -i '/start telegraf/d' ${DHCLIENT_HOOKS_FILE}
fi
