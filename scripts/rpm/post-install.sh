#!/bin/bash

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
   chmod 640 /etc/telegraf/telegraf.conf
   chmod 750 /etc/telegraf/telegraf.d
fi

# Set up log directories
LOG_DIR=/var/log/telegraf
test -d $LOG_DIR || {
    mkdir -p $LOG_DIR
    chown -R -L telegraf:telegraf $LOG_DIR
    chmod 755 $LOG_DIR
}

STATE_DIR=/var/lib/telegraf
test -d "$STATE_DIR" || {
    mkdir -p "$STATE_DIR"
    chmod 770 "$STATE_DIR"
    chown root:telegraf "$STATE_DIR"
}

STATE_FILE="$STATE_DIR/statefile"
test -f "$STATE_FILE" || {
    touch "$STATE_FILE"
    chown root:telegraf "$STATE_FILE"
    chmod 660 "$STATE_FILE"
}

# Set up systemd service - check if the systemd directory exists per:
# https://www.freedesktop.org/software/systemd/man/sd_booted.html
if [[ -d /run/systemd/system ]]; then
    cp -f /usr/lib/telegraf/scripts/telegraf.service /usr/lib/systemd/system/telegraf.service
    systemctl enable telegraf
    systemctl daemon-reload
fi
