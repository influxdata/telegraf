#!/usr/bin/bash

# restart when only the telegraf binaries are updated
[[ -f /etc/telegraf/telegraf.con ]] && systemctl restart telegraf
