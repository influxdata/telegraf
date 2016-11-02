#!/usr/bin/bash

# Issue:
# Telegraf uses hostname for every point writtern to InfluxDB
# If Telegraf is started before hostname is setup via DHCP and the 
# network is UP then all points get written with host tag = 'localhost'
# Telegraf isn't restarted when the hostname gets setup after DHCP interaction completes.
# This issue happens only on systemd enabled systems.

# On systems with dhclient, telegraf is disabled at boot up and started 
#   only after dhclient has configured the hostname, IP etc using dhclient hooks
#   NOTE: a4 workspaces also use dhclient so we can't do simple process check
#   so we use lack of systemd-networkd to assume that dhclient is being used to manage DHCP
#   properties
# On systems with networkd, telegraf uses network-online systemd target

TRUE="0"
SYSTEMD_INUSE=`systemctl -a --no-pager > /dev/null;  echo $?`
SYSTEMD_NETWORKD_INUSE=`systemctl status systemd-networkd | grep 'Active' | grep running  > /dev/null; echo $?`
TELEGRAF_IN_DHCLIENT_HOOKS=`grep telegraf /etc/dhcp/dhclient-up-hooks > /dev/null 2>&1; echo $?`
TELEGRAF_SERVICE=/usr/lib/systemd/system/telegraf.service

if [[ "$SYSTEMD_INUSE" != "$TRUE" ]]; then
	exit 0
fi

if [[ "$SYSTEMD_NETWORKD_INUSE" != "$TRUE" ]]; then

  if [[ -f $TELEGRAF_SERVICE ]]; then
  	systemctl disable telegraf
  fi

  cp /usr/lib/systemd/system/telegraf-dhclient.service /usr/lib/systemd/system/telegraf.service
  systemctl daemon-reload
  if [[ "$TELEGRAF_IN_DHCLIENT_HOOKS" != "$TRUE"  ]]
  then
  cat >> /etc/dhcp/dhclient-up-hooks <<EOF
   echo "Starting telegraf from dhclient hook"; systemctl start telegraf
EOF
  chmod a+x /etc/dhcp/dhclient-up-hooks
  fi
else
  cp /usr/lib/systemd/system/telegraf-networkd.service /usr/lib/systemd/system/telegraf.service
  systemctl daemon-reload
  systemctl enable telegraf
fi

# restart telegraf only if hostname is not localhost
MYHOST=`hostname -s`
if [[ "$MYHOST" != "localhost" ]]
then
  systemctl restart telegraf
fi
