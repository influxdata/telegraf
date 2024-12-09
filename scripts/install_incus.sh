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
    ipv4.address: 192.168.100.1/24
    ipv4.nat: true
    ipv6.address: none
storage_pools:
- name: default
  driver: dir
profiles:
- name: default
  devices:
    eth0:
      name: eth0
      network: incusbr0
      type: nic
    root:
      path: /
      pool: default
      type: disk
EOF

# Disable the firewall to prevent issues with the default configuration.
sudo ufw disable

sudo usermod -a -G incus-admin "$(whoami)"
