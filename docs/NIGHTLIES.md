
# Nightly Builds

These builds are generated from the master branch each night:

| DEB             | RPM             | TAR GZ                        | ZIP |
| --------------- | --------------- | ------------------------------| --- |
| [amd64.deb](https://dl.influxdata.com/telegraf/nightlies/telegraf_nightly_amd64.deb)   | [aarch64.rpm](https://dl.influxdata.com/telegraf/nightlies/telegraf-nightly.aarch64.rpm) | [darwin_amd64.tar.gz](https://dl.influxdata.com/telegraf/nightlies/telegraf-nightly_darwin_amd64.tar.gz)       |  [windows_amd64.zip](https://dl.influxdata.com/telegraf/nightlies/telegraf-nightly_windows_amd64.zip) |
| [arm64.deb](https://dl.influxdata.com/telegraf/nightlies/telegraf_nightly_arm64.deb)   | [armel.rpm](https://dl.influxdata.com/telegraf/nightlies/telegraf-nightly.armel.rpm)   | [darwin_arm64.tar.gz](https://dl.influxdata.com/telegraf/nightlies/telegraf-nightly_darwin_arm64.tar.gz)     |  [windows_i386.zip](https://dl.influxdata.com/telegraf/nightlies/telegraf-nightly_windows_i386.zip) |
| [armel.deb](https://dl.influxdata.com/telegraf/nightlies/telegraf_nightly_armel.deb)   | [armv6hl.rpm](https://dl.influxdata.com/telegraf/nightlies/telegraf-nightly.armv6hl.rpm) | [freebsd_amd64.tar.gz](https://dl.influxdata.com/telegraf/nightlies/telegraf-nightly_freebsd_amd64.tar.gz)      | |
| [armhf.deb](https://dl.influxdata.com/telegraf/nightlies/telegraf_nightly_armhf.deb)   | [i386.rpm](https://dl.influxdata.com/telegraf/nightlies/telegraf-nightly.i386.rpm)    | [freebsd_armv7.tar.gz](https://dl.influxdata.com/telegraf/nightlies/telegraf-nightly_freebsd_armv7.tar.gz)       | |
| [i386.deb](https://dl.influxdata.com/telegraf/nightlies/telegraf_nightly_i386.deb)    | [ppc64le.rpm](https://dl.influxdata.com/telegraf/nightlies/telegraf-nightly.ppc64le.rpm) | [freebsd_i386.tar.gz](https://dl.influxdata.com/telegraf/nightlies/telegraf-nightly_freebsd_i386.tar.gz)        | |
| [mips.deb](https://dl.influxdata.com/telegraf/nightlies/telegraf_nightly_mips.deb)    | [riscv64.rpm](https://dl.influxdata.com/telegraf/nightlies/telegraf-nightly.riscv64.rpm) | [linux_amd64.tar.gz](https://dl.influxdata.com/telegraf/nightlies/telegraf-nightly_linux_amd64.tar.gz)        | |
| [mipsel.deb](https://dl.influxdata.com/telegraf/nightlies/telegraf_nightly_mipsel.deb)  | [s390x.rpm](https://dl.influxdata.com/telegraf/nightlies/telegraf-nightly.s390x.rpm) | [linux_arm64.tar.gz](https://dl.influxdata.com/telegraf/nightlies/telegraf-nightly_linux_arm64.tar.gz)        | |
| [ppc64el.deb](https://dl.influxdata.com/telegraf/nightlies/telegraf_nightly_ppc64el.deb) | [x86_64.rpm](https://dl.influxdata.com/telegraf/nightlies/telegraf-nightly.x86_64.rpm) | [linux_armel.tar.gz](https://dl.influxdata.com/telegraf/nightlies/telegraf-nightly_linux_armel.tar.gz)        | |
| [riscv64.deb](https://dl.influxdata.com/telegraf/nightlies/telegraf_nightly_riscv64.deb) |    |  [linux_armhf.tar.gz](https://dl.influxdata.com/telegraf/nightlies/telegraf-nightly_linux_armhf.tar.gz)         | |
| [s390x.deb](https://dl.influxdata.com/telegraf/nightlies/telegraf_nightly_s390x.deb) |    | [linux_i386.tar.gz](https://dl.influxdata.com/telegraf/nightlies/telegraf-nightly_linux_i386.tar.gz)         | |
|                 |                 | [linux_mips.tar.gz](https://dl.influxdata.com/telegraf/nightlies/telegraf-nightly_linux_mips.tar.gz)         | |
|                 |                 | [linux_mipsel.tar.gz](https://dl.influxdata.com/telegraf/nightlies/telegraf-nightly_linux_mipsel.tar.gz)       | |
|                 |                 | [linux_ppc64le.tar.gz](https://dl.influxdata.com/telegraf/nightlies/telegraf-nightly_linux_ppc64le.tar.gz)      | |
|                 |                 | [linux_riscv64.tar.gz](https://dl.influxdata.com/telegraf/nightlies/telegraf-nightly_linux_riscv64.tar.gz)      | |
|                 |                 | [linux_s390x.tar.gz](https://dl.influxdata.com/telegraf/nightlies/telegraf-nightly_linux_s390x.tar.gz)        | |
|                 |                 | [static_linux_amd64.tar.gz](https://dl.influxdata.com/telegraf/nightlies/telegraf-nightly_static_linux_amd64.tar.gz) | |

Nightly docker images are available on [quay.io](https://quay.io/repository/influxdb/telegraf-nightly?tab=tags):

```shell
# Debian-based image
docker pull quay.io/influxdb/telegraf-nightly:latest
# Alpine-based image
docker pull quay.io/influxdb/telegraf-nightly:alpine
```
