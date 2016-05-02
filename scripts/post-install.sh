#!/bin/bash

BIN_DIR=/usr/bin
LOG_DIR=/var/log/telegraf
SCRIPT_DIR=/usr/lib/telegraf/scripts
LOGROTATE_DIR=/etc/logrotate.d

function install_init {
    cp -f $SCRIPT_DIR/init.sh /etc/init.d/telegraf
    chmod +x /etc/init.d/telegraf
}

function install_systemd {
    cp -f $SCRIPT_DIR/telegraf.service /lib/systemd/system/telegraf.service
    systemctl enable telegraf || true
    systemctl daemon-reload || true
}

function install_update_rcd {
    update-rc.d telegraf defaults
}

function install_chkconfig {
    chkconfig --add telegraf
}

id telegraf &>/dev/null
if [[ $? -ne 0 ]]; then
    useradd -r -K USERGROUPS_ENAB=yes -M telegraf -s /bin/false -d /etc/telegraf
fi

test -d $LOG_DIR || mkdir -p $LOG_DIR
chown -R -L telegraf:telegraf $LOG_DIR
chmod 755 $LOG_DIR

# Remove legacy symlink, if it exists
if [[ -L /etc/init.d/telegraf ]]; then
    rm -f /etc/init.d/telegraf
fi

# Add defaults file, if it doesn't exist
if [[ ! -f /etc/default/telegraf ]]; then
    touch /etc/default/telegraf
fi

# Add .d configuration directory
if [[ ! -d /etc/telegraf/telegraf.d ]]; then
    mkdir -p /etc/telegraf/telegraf.d
fi

# Distribution-specific logic
if [[ -f /etc/redhat-release ]]; then
    # RHEL-variant logic
    which systemctl &>/dev/null
    if [[ $? -eq 0 ]]; then
	    install_systemd
    else
	    # Assuming sysv
	    install_init
	    install_chkconfig
    fi
elif [[ -f /etc/debian_version ]]; then
    # Debian/Ubuntu logic
    which systemctl &>/dev/null
    if [[ $? -eq 0 ]]; then
	    install_systemd
	    systemctl restart telegraf || echo "WARNING: systemd not running."
    else
	    # Assuming sysv
	    install_init
	    install_update_rcd
	    invoke-rc.d telegraf restart
    fi
elif [[ -f /etc/os-release ]]; then
    source /etc/os-release
    if [[ $ID = "amzn" ]]; then
	    # Amazon Linux logic
	    install_init
	    install_chkconfig
    fi
fi
