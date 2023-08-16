#!/bin/bash

SCRIPT_DIR=/usr/lib/telegraf/scripts

function install_init {
    cp -f $SCRIPT_DIR/init.sh /etc/init.d/telegraf
    chmod +x /etc/init.d/telegraf
}

function install_systemd {
    #shellcheck disable=SC2086
    cp -f $SCRIPT_DIR/telegraf.service $1
    systemctl enable telegraf || true
    systemctl daemon-reload || true
}

function install_update_rcd {
    update-rc.d telegraf defaults
}

function install_chkconfig {
    chkconfig --add telegraf
}

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

if [ -d /run/systemd/system ]; then
    install_systemd /lib/systemd/system/telegraf.service
    # if and only if the service was already running then restart
    deb-systemd-invoke try-restart telegraf.service >/dev/null || true
else
	# Assuming SysVinit
	install_init
	# Run update-rc.d or fallback to chkconfig if not available
	if which update-rc.d &>/dev/null; then
		install_update_rcd
	else
		install_chkconfig
	fi
	invoke-rc.d telegraf restart
fi
