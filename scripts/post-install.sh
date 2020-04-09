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
fi

# Distribution-specific logic
if [[ -f /etc/redhat-release ]] || [[ -f /etc/SuSE-release ]]; then
    # RHEL-variant logic
    if [[ "$(readlink /proc/1/exe)" == */systemd ]]; then
        install_systemd /usr/lib/systemd/system/telegraf.service
    else
        # Assuming SysVinit
        install_init
        # Run update-rc.d or fallback to chkconfig if not available
        if which update-rc.d &>/dev/null; then
            install_update_rcd
        else
            install_chkconfig
        fi
    fi
elif [[ -f /etc/debian_version ]]; then
    # Debian/Ubuntu logic

    # Ownership for RH-based platforms is set in build.py via the `rmp-attr` option.
    # We perform ownership change only for Debian-based systems.
    # Moving these lines out of this if statement would make `rmp -V` fail after installation.
    test -d $LOG_DIR || mkdir -p $LOG_DIR
    chown -R -L telegraf:telegraf $LOG_DIR
    chmod 755 $LOG_DIR

    if [[ "$(readlink /proc/1/exe)" == */systemd ]]; then
        install_systemd /lib/systemd/system/telegraf.service
        deb-systemd-invoke restart telegraf.service || echo "WARNING: systemd not running."
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
elif [[ -f /etc/os-release ]]; then
    source /etc/os-release
    if [[ "$NAME" = "Amazon Linux" ]]; then
        # Amazon Linux 2+ logic
        install_systemd /usr/lib/systemd/system/telegraf.service
    elif [[ "$NAME" = "Amazon Linux AMI" ]]; then
        # Amazon Linux logic
        install_init
        # Run update-rc.d or fallback to chkconfig if not available
        if which update-rc.d &>/dev/null; then
            install_update_rcd
        else
            install_chkconfig
        fi
    fi
fi
