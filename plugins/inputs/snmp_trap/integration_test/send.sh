#!/bin/sh
echo "in send.sh"
snmptrap -v 2c -c community udp:telegraf_test:12399 1234567 .1.3.6.1.6.3.1.1.5.3.0 0 s "This is a test linkDown trap from v2c"
echo $?
