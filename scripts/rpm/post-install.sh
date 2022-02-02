#!/bin/bash

BIN_DIR=/usr/bin
LOG_DIR=/var/log/telegraf
SCRIPT_DIR=/usr/lib/telegraf/scripts
LOGROTATE_DIR=/etc/logrotate.d

# Remove legacy symlink, if it exists
if [[ -L /etc/init.d/telegraf ]]; then
    rm -f /etc/init.d/telegraf
fi
# Remove legacy symlink, if it exists
if [[ -L /etc/systemd/system/telegraf.service ]]; then
    rm -f /etc/systemd/system/telegraf.service
fi

# Add defaults file, if it doesn't exist
if [[ ! -f /etc/default/telegraf ]]; then
    touch /etc/default/telegraf
fi

# Add .d configuration directory
if [[ ! -d /etc/telegraf/telegraf.d ]]; then
    mkdir -p /etc/telegraf/telegraf.d
fi

# If 'telegraf.conf' is not present use package's sample (fresh install)
if [[ ! -f /etc/telegraf/telegraf.conf ]] && [[ -f /etc/telegraf/telegraf.conf.sample ]]; then
   cp /etc/telegraf/telegraf.conf.sample /etc/telegraf/telegraf.conf
fi

# Set up log directories
test -d $LOG_DIR || mkdir -p $LOG_DIR
chown -R -L telegraf:telegraf $LOG_DIR
chmod 755 $LOG_DIR

# Set up systemd service
if [[ -d /run/systemd/system ]]; then
    cp -f $SCRIPT_DIR/telegraf.service /usr/lib/systemd/system/telegraf.service
    systemctl enable telegraf || true
    systemctl daemon-reload || true
fi
