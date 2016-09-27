#!/usr/bin/bash

# restart when only the telegraf binaries are updated
if [[ -f /etc/telegraf/telegraf.conf ]];
then
   systemctl restart telegraf
fi
