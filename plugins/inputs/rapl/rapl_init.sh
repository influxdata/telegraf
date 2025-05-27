#!/usr/bin/env bash

find /sys/devices/virtual/powercap/intel-rapl -type f -name energy_uj -exec chmod 444 {} \;
