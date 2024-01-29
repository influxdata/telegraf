#!/bin/bash
set -eux

sudo mkdir -p /etc/apt/keyrings
sudo curl -fsSL https://pkgs.zabbly.com/key.asc -o /etc/apt/keyrings/zabbly.asc
sudo sh -c 'cat <<EOF > /etc/apt/sources.list.d/zabbly-incus-stable.sources
Enabled: yes
Types: deb
URIs: https://pkgs.zabbly.com/incus/stable
Suites: $(. /etc/os-release && echo ${VERSION_CODENAME})
Components: main
Architectures: $(dpkg --print-architecture)
Signed-By: /etc/apt/keyrings/zabbly.asc

EOF'

sudo apt-get update && sudo apt-get install --yes incus

# On CircleCI instances do not have IPv6 enabled, force IPv4 usage only.
cat <<EOF | sudo incus admin init --preseed
networks:
- name: incusbr0
  type: bridge
  config:
    ipv4.address: auto
    ipv6.address: none
EOF

sudo usermod -a -G incus-admin "$(whoami)"
