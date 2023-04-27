#!/bin/bash

KNOWN_DISTRIBUTION="(Debian|Ubuntu|RedHat|CentOS|openSUSE|Amazon|Arista|SUSE|Rocky|AlmaLinux)"
DISTRIBUTION=$(lsb_release -d 2>/dev/null | grep -Eo $KNOWN_DISTRIBUTION  || grep -Eo $KNOWN_DISTRIBUTION /etc/issue 2>/dev/null || grep -Eo $KNOWN_DISTRIBUTION /etc/Eos-release 2>/dev/null || grep -m1 -Eo $KNOWN_DISTRIBUTION /etc/os-release 2>/dev/null || uname -s)

if [ "$DISTRIBUTION" = "Darwin" ]; then
    CHASSIS=
    HOST=$(uname -n)
    OS=$(uname -s)
    DISTRO=$(uname -s)
    PROCESSOR=$(uname -p)
    KERNELRELEASE=$(uname -r)
    KERNELVERSION=$(uname -v)
    MACHINE=$(uname -m)
    HARDWAREPLATFORM=
    VIRTUALIZATION=
elif [ "$DISTRIBUTION" = "FreeBSD" ]; then
    CHASSIS=
    HOST=$(uname -n)
    OS=$(uname -o)
    DISTRO=$(uname -s)
    KERNELNAME=$(uname -s)
    PROCESSOR=$(uname -p)
    KERNELRELEASE=$(uname -r)
    KERNELVERSION=$(uname -v)
    MACHINE=$(uname -m)
    HARDWAREPLATFORM=$(uname -i)
    VIRTUALIZATION=
else
    CHASSIS=$(hostnamectl | sed -n 's/Chassis:[ \t]*//p')
    CHASSIS=$(sed "s/^[ \t]*//" <<< $CHASSIS)
    HOST=$(uname -n)
    OS=$(uname -o)
    DISTRO=$(hostnamectl | sed -n 's/Operating System:[ \t]*//p')
    DISTRO=$(sed "s/^[ \t]*//" <<< $DISTRO)
    KERNELNAME=$(uname -s)
    PROCESSOR=$(uname -p)
    KERNELRELEASE=$(uname -r)
    KERNELVERSION=$(uname -v)
    MACHINE=$(uname -m)
    HARDWAREPLATFORM=$(uname -i)
    VIRTUALIZATION=$(hostnamectl | sed -n 's/Virtualization:[ \t]*//p')
    VIRTUALIZATION=$(sed "s/^[ \t]*//" <<< $VIRTUALIZATION)
fi

# echo $DISTRO
echo "server_information,type=platform chassis=\"${CHASSIS}\",hostname=\"${HOST}\",os=\"${OS}-${DISTRO}\",kernel_name=\"${KERNELNAME}\",processor=\"${PROCESSOR}\",kernel_release=\"${KERNELRELEASE}\",kernel_version=\"${KERNELVERSION}\",machine=\"${MACHINE}\",hardware_platform=\"${HARDWAREPLATFORM}\",virtualization=\"${VIRTUALIZATION}\""
