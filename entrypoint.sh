#!/bin/bash
set -e

if [ -f $HOST_ETC/hostname ] ; then
  export NODE_HOSTNAME=`cat $HOST_ETC/hostname`
fi

if [ "${1:0:1}" = '-' ]; then
    set -- telegraf "$@"
fi

exec "$@"
