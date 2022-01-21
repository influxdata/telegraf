# Aiven Procstat Input Plugin

Was copied from the procstat input. Divergences: 

* add 'systemd_units' configuration parameter. A list that specifies the units to fetch the pids from
* to that end it parses the output from `systemctl status` in one go instead of invoking `systemctl status [...]` for every unit
* it is not possible to use the globbing feature of the original `procstat` input for several reasons, one being that the tags are not expanded with the glob, the other is that the units we are targeting are not named glob friendly

