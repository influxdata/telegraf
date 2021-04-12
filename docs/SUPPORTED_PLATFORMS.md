Supported Platforms
===================
Telegraf is a cross-platform application. This doc helps define which operating systems, distributions, and releases Telegraf supports.

Telegraf is supported on Linux, FreeBSD, Windows, and macOS. It is written in Go which supports these operating systems and more. Telegraf may work on Go's other operating systems and users are welcome to build their own binaries for them. Bug reports should be submitted only for supported platforms.

https://golang.org/doc/install/source

https://github.com/golang/go/wiki/MinimumRequirements#operating-systems

Linux
-----
Telegraf intent: *Support latest versions of the most popular distributions*

https://distrowatch.com/dwres.php?resource=major

Telegraf supports RHEL, Fedora, Debian, and Ubuntu. InfluxData provides package repositories for these distributions. Instructions for using the package repositories can be found on [docs.influxdata.com](https://docs.influxdata.com/telegraf/v1.16/introduction/installation/) Bug reports shuld be submitted only for supported distributions and releases.

Telegraf releases include .tar.gz packages for use with other distributions, for building containers, or for installation without a package manager. As part of telegraf's release process we publish official images to Docker Hub.

https://hub.docker.com/_/telegraf

### RHEL
Red Hat makes a major release every four to five years and supports each release in production for ten years. Extended support is available for three or more years.

Telegraf intent: *Support releases in RHEL production, but not in extended support.*

https://en.wikipedia.org/wiki/Red_Hat_Enterprise_Linux#Version_history_and_timeline

As of April 2021, 7 and 8 are production releases.

### Ubuntu
Ubuntu makes two releases a year. Every two years one of the releases is an LTS (long-term support) release. Non-LTS releases are supported for nine months. LTS releases are in maintenance for five years, then in extended security maintenance for up to three more years.

Telegraf intent: *Support LTS releases in Ubuntu maintenance, but not in extended security maintenance.*

https://ubuntu.com/about/release-cycle

https://en.wikipedia.org/wiki/Ubuntu_version_history

As of April 2021, Ubuntu 16.04 LTS, 20.04 LTS, and 21.04 are in maintenance.

### Debian
Debian generally makes major releases every two years and provides security support for each release for three years.

Telegraf intent: *Support releases under Debian security support*

https://en.wikipedia.org/wiki/Debian_version_history#Release_table

As of April 2021, Debian 10 is in security support.

### Fedora
Fedora makes two releases a year and supports each release for a year.

Telegraf intent: *Support releases supported by Fedora*

https://en.wikipedia.org/wiki/Fedora_version_history#Version_history

FreeBSD
-------
FreeBSD makes major releases about every two years. Releases reach end of life after five years.

Telegraf intent: *Support releases under FreeBSD security support*

https://en.wikipedia.org/wiki/FreeBSD#Version_history

https://www.freebsd.org/security/#sup

As of April 2021, releases 11 and 12 are under security support.

Windows
-------
Windows 10 and Windows Server.

### Windows 10
Windows 10 releases are in mainstream servicing timeline for 18 or 30 months.

Telegraf intent: *Support versions in mainstream servicing timeline*

https://docs.microsoft.com/en-us/lifecycle/faq/windows

https://en.wikipedia.org/wiki/Windows_10#Feature_updates

### Windows Server

Windows server has much longer support periods: 5 years mainstream + 5 additional years extended.

https://en.wikipedia.org/wiki/Windows_Server#Long_Term_Servicing_Channel_(LTSC)

Telegraf intent: *Support releases under mainstream support*

As of April 2021, Server 2016 and Server 2019 are under mainstream support.

macOS
-----
MacOS makes one major release a year and provides support for each release for three years.

Telegraf intent: *Support releases supported by Apple*

https://en.wikipedia.org/wiki/MacOS#Release_history

As of April 2021, 10.14, 10.15, and 11 are supported.
