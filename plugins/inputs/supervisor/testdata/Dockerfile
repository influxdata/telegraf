FROM debian:stretch
RUN apt update && apt install -y supervisor && apt clean && rm -rf /var/lib/apt/lists/*
CMD /usr/bin/supervisord -c /etc/supervisor/supervisord.conf