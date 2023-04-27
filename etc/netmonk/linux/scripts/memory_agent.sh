#!/bin/bash

KNOWN_DISTRIBUTION="(Debian|Ubuntu|RedHat|CentOS|openSUSE|Amazon|Arista|SUSE|Rocky|AlmaLinux)"
DISTRIBUTION=$(lsb_release -d 2>/dev/null | grep -Eo $KNOWN_DISTRIBUTION  || grep -Eo $KNOWN_DISTRIBUTION /etc/issue 2>/dev/null || grep -Eo $KNOWN_DISTRIBUTION /etc/Eos-release 2>/dev/null || grep -m1 -Eo $KNOWN_DISTRIBUTION /etc/os-release 2>/dev/null || uname -s)

if [ "$DISTRIBUTION" = "Darwin" ]; then
    MEMTOTAL=$(sysctl -n hw.physmem) # in Kb
    SWAPTOTAL=0
elif [ "$DISTRIBUTION" = "FreeBSD" ]; then
    MEMTOTAL=$(sysctl -n hw.physmem) # in Kb
    SWAPTOTAL=$(swapinfo -k | awk '{if(NR==2) print $4}') # in Kb
else     
    MEMTOTAL=$(awk '/MemTotal/ { print $2 }' /proc/meminfo) # in Kb
    SWAPTOTAL=$(awk '/SwapTotal/ { print $2 }' /proc/meminfo) # in Kb
fi

echo "server_information,type=memory mem_total=${MEMTOTAL},swap_total=${SWAPTOTAL}"
