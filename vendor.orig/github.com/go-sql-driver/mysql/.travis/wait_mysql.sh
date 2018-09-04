#!/bin/sh
while :
do
    if mysql -e 'select version()' 2>&1 | grep 'version()\|ERROR 2059 (HY000):'; then
        break
    fi
    sleep 3
done
