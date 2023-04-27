#!/bin/bash

KNOWN_DISTRIBUTION="(Debian|Ubuntu|RedHat|CentOS|openSUSE|Amazon|Arista|SUSE|Rocky|AlmaLinux)"
DISTRIBUTION=$(lsb_release -d 2>/dev/null | grep -Eo $KNOWN_DISTRIBUTION  || grep -Eo $KNOWN_DISTRIBUTION /etc/issue 2>/dev/null || grep -Eo $KNOWN_DISTRIBUTION /etc/Eos-release 2>/dev/null || grep -m1 -Eo $KNOWN_DISTRIBUTION /etc/os-release 2>/dev/null || uname -s)

if [ "$DISTRIBUTION" = "Darwin" ]; then
    VENDORID=
    MODELNAME=$(sysctl hw.model | sed -n 's/hw.model: [ \t]*//p') 
    CPUS=$(sysctl hw.ncpu | sed -n 's/hw.ncpu: [ \t]*//p') 
    MHZ=
    FAMILY=
    MODEL=
    STEPPING=
elif [ "$DISTRIBUTION" = "FreeBSD" ]; then
    VENDORID=
    MODELNAME=$(sysctl hw.model | sed -n 's/hw.model: [ \t]*//p') 
    CPUS=$(sysctl hw.ncpu | sed -n 's/hw.ncpu: [ \t]*//p') 
    MHZ=
    FAMILY=
    MODEL=
    STEPPING=
else
    VENDORID=$(lscpu | sed -n 's/Vendor ID:[ \t]*//p')
    MODELNAME=$(lscpu | sed -n 's/Model name:[ \t]*//p')
    THREADS=$(lscpu | sed -n 's/Thread(s) per core:[ \t]*//p')
    CORES=$(lscpu | sed -n 's/Core(s) per socket:[ \t]*//p')
    SOCKET=$(lscpu | sed -n 's/Socket(s):[ \t]*//p')
    CPUS=$(expr $THREADS '*' $CORES '*' $SOCKET) # LOGICAL CPUS
    MHZ=$(lscpu | sed -n 's/CPU MHz:[ \t]*//p')
    FAMILY=$(lscpu | sed -n 's/CPU family:[ \t]*//p')
    MODEL=$(lscpu | sed -n 's/Model:[ \t]*//p')
    STEPPING=$(lscpu | sed -n 's/Stepping:[ \t]*//p')
fi

echo "server_information,type=cpu vendor_id=\"${VENDORID}\",model_name=\"${MODELNAME}\",cpus=${CPUS},MHz=\"${MHZ}\",family=\"${FAMILY}\",model=\"${MODEL}\",stepping=\"${STEPPING}\""
