#!/bin/sh

# confirm correct permissions on chrony run directory
if [ -d /run/chrony ]; then
  chown -R chrony:chrony /run/chrony
  chmod o-rx /run/chrony
  # remove previous pid file if it exist
  rm -f /var/run/chrony/chronyd.pid
fi

# confirm correct permissions on chrony variable state directory
if [ -d /var/lib/chrony ]; then
  chown -R chrony:chrony /var/lib/chrony
fi

## startup chronyd in the foreground
exec /usr/sbin/chronyd -u chrony -d -x -f /etc/telegraf-chrony.conf