Telegraf docker image

This image uses $HOST_ETC/hostname to populate $NODE_HOSTNAME, which can then be used in telegraf.conf. This is intended for deployment in Swarm or other environments with centralized configuration.
