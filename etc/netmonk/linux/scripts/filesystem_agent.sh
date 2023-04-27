#!/bin/bash

KNOWN_DISTRIBUTION="(Debian|Ubuntu|RedHat|CentOS|openSUSE|Amazon|Arista|SUSE|Rocky|AlmaLinux)"
DISTRIBUTION=$(lsb_release -d 2>/dev/null | grep -Eo $KNOWN_DISTRIBUTION  || grep -Eo $KNOWN_DISTRIBUTION /etc/issue 2>/dev/null || grep -Eo $KNOWN_DISTRIBUTION /etc/Eos-release 2>/dev/null || grep -m1 -Eo $KNOWN_DISTRIBUTION /etc/os-release 2>/dev/null || uname -s)

if [ "$DISTRIBUTION" = "Darwin" ]; then
    df -h | awk '{
        if(NR!=1){
            if(NR==2){
                SOURCE=$1
                SIZE=$2
                TARGET=$9
            }else{
                SOURCE=SOURCE","$1
                SIZE=SIZE","$2
                TARGET=TARGET","$9
            }
        }
    }
    END {printf "server_information,type=filesystem key=\"%s\",size=\"%s\",mounted_on=\"%s\"\n", SOURCE, SIZE, TARGET}
    '
elif [ "$DISTRIBUTION" = "FreeBSD" ]; then
    df -h | awk '{
        if(NR!=1){
            if(NR==2){
                SOURCE=$1
                SIZE=$2
                TARGET=$6
            }else{
                SOURCE=SOURCE","$1
                SIZE=SIZE","$2
                TARGET=TARGET","$6
            }
        }
    }
    END {printf "server_information,type=filesystem key=\"%s\",size=\"%s\",mounted_on=\"%s\"\n", SOURCE, SIZE, TARGET}
    '
else
    df -h | awk '{
        if(NR!=1){
            if(NR==2){
                SOURCE=$1
                SIZE=$2
                TARGET=$6
            }else{
                SOURCE=SOURCE","$1
                SIZE=SIZE","$2
                TARGET=TARGET","$6
            }
        }
    }
    END {printf "server_information,type=filesystem key=\"%s\",size=\"%s\",mounted_on=\"%s\"\n", SOURCE, SIZE, TARGET}
    '
fi
