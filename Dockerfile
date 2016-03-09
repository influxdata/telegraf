FROM alpine

MAINTAINER ops@yotpo.com

RUN mkdir -p /etc/telegraf/telegraf.d/

COPY telegraf /usr/bin/
COPY telegraf.conf /etc/telegraf/

CMD [ "/usr/bin/telegraf", "-config", "/etc/telegraf/telegraf.conf", "-config-directory", "/etc/telegraf/telegraf.d"]
