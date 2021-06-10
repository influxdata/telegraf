# Supported Platforms

Telegraf is a cross-platform application. This doc helps define which
operating systems, distributions, and releases Telegraf supports.

Telegraf is supported on Linux, FreeBSD, Windows, and macOS. It is
written in Go which supports these operating systems and
more. Telegraf may work on Go's other operating systems and users are
welcome to build their own binaries for them. Bug reports should be
submitted only for supported platforms.

Golang.org has a [table][go-table] of valid OS and architecture
combinations and the golang wiki has more specific [minimum
requirements][go-reqs] for Go itself.

[go-table]: https://golang.org/doc/install/source#environment
[go-reqs]: https://github.com/golang/go/wiki/MinimumRequirements#operating-systems

## Linux

Telegraf intent: *Support latest versions of major linux
distributions*

Telegraf supports RHEL, Fedora, Debian, and Ubuntu. InfluxData
provides package repositories for these distributions. Instructions
for using the package repositories can be found on
[docs.influxdata.com][repo-docs]. Bug reports should be submitted only
for supported distributions and releases.

Telegraf's Debian or Ubuntu packages are likely to work on other
Debian-based distributions although these are not
supported. Similarly, Telegraf's Fedora and RHEL packages are likely
to work on other Redhat-based distributions although again these are
not supported.

Telegraf releases include .tar.gz packages for use with other
distributions, for building container images, or for installation
without a package manager. As part of telegraf's release process we
publish [official images][docker-hub] to Docker Hub.

Distrowatch lists [major distributions][dw-major] and tracks
[popularity][dw-pop] of distributions. Wikipedia lists [linux
distributions][wp-distro] by the major distribution they're based on.

[repo-docs]: https://docs.influxdata.com/telegraf/latest/introduction/installation/
[docker-hub]: https://hub.docker.com/_/telegraf
[dw-major]: https://distrowatch.com/dwres.php?resource=major
[dw-pop]: https://distrowatch.com/dwres.php?resource=popularity
[wp-distro]: https://en.wikipedia.org/wiki/List_of_Linux_distributions

### RHEL

Red Hat makes a major release every four to five years and supports
each release in production for ten years. Extended support is
available for three or more years.

Telegraf intent: *Support releases in RHEL production, but not in
extended support.*

Redhat publishes [release history][rh-history] and wikipedia has a
[summary timeline][wp-rhel].

As of April 2021, 7 and 8 are production releases.

[rh-history]: https://access.redhat.com/articles/3078
[wp-rhel]: https://en.wikipedia.org/wiki/Red_Hat_Enterprise_Linux#Version_history_and_timeline

### Ubuntu

Ubuntu makes two releases a year. Every two years one of the releases
is an LTS (long-term support) release. Interim (non-LTS) releases are
in standard support for nine months. LTS releases are in maintenance
for five years, then in extended security maintenance for up to three
more years.

Telegraf intent: *Support interim releases and LTS releases in Ubuntu
maintenance, but not in extended security maintenance.*

Ubuntu publishes [release history][ub-history] and wikipedia has a
[table][wp-ub] of all releases and support status.

As of April 2021, Ubuntu 20.10 is in standard support. Ubuntu 18.04
LTS and 20.04 LTS are in maintenance.

[ub-history]: https://ubuntu.com/about/release-cycle
[wp-ub]: https://en.wikipedia.org/wiki/Ubuntu_version_history#Table_of_versions

### Debian

Debian generally makes major releases every two years and provides
security support for each release for three years. After security
support expires the release enters long term support (LTS) until at
least five years after release.

Telegraf intent: *Support releases under Debian security support*

Debian publishes [releases and support status][deb-history] and
wikipedia has a [summary table][wp-deb].

As of April 2021, Debian 10 is in security support.

[deb-history]: https://www.debian.org/releases/
[wp-deb]: https://en.wikipedia.org/wiki/Debian_version_history#Release_table

### Fedora

Fedora makes two releases a year and supports each release for a year.

Telegraf intent: *Support releases supported by Fedora*

Fedora publishes [release history][fed-history] and wikipedia has a
[summary table][wp-fed].

[fed-history]: https://fedoraproject.org/wiki/Releases
[wp-fed]: https://en.wikipedia.org/wiki/Fedora_version_history#Version_history

## FreeBSD

FreeBSD makes major releases about every two years. Releases reach end
of life after five years.

Telegraf intent: *Support releases under FreeBSD security support*

FreeBSD publishes [release history][freebsd-history] and wikipedia has
a [summary table][wp-freebsd].

As of April 2021, releases 11 and 12 are under security support.

[freebsd-history]: https://www.freebsd.org/security/#sup
[wp-freebsd]: https://en.wikipedia.org/wiki/FreeBSD#Version_history

## Windows

Telegraf intent: *Support current versions of Windows and Windows
Server*

Microsoft has two release channels, the semi-annual channel (SAC) and
the long-term servicing channel (LTSC). The semi-annual channel is for
mainstream feature releases.

Microsoft publishes [lifecycle policy by release][ms-lifecycle] and a
[product lifecycle faq][ms-lifecycle-faq].

[ms-lifecycle]: https://docs.microsoft.com/en-us/lifecycle/products/?terms=windows
[ms-lifecycle-faq]: https://docs.microsoft.com/en-us/lifecycle/faq/windows

### Windows 10

Windows 10 makes SAC releases twice a year and supports those releases
for [18 or 30 months][w10-timeline]. They also make LTSC releases
which are supported for 10 years but are intended only for medical or
industrial devices that require a static feature set.

Telegraf intent: *Support semi-annual channel releases supported by
Microsoft*

Microsoft publishes Windows 10 [release information][w10-history], and
[servicing channels][w10-channels]. Wikipedia has a [summary
table][wp-w10] of support status.

As of April 2021, versions 19H2, 20H1, and 20H2 are supported.

[w10-timeline]: https://docs.microsoft.com/en-us/lifecycle/faq/windows#what-is-the-servicing-timeline-for-a-version-feature-update-of-windows-10
[w10-history]: https://docs.microsoft.com/en-us/windows/release-health/release-information
[w10-channels]: https://docs.microsoft.com/en-us/windows/deployment/update/get-started-updates-channels-tools
[wp-w10]: https://en.wikipedia.org/wiki/Windows_10_version_history#Channels

### Windows Server

Windows Server makes SAC releases for that are supported for 18 months
and LTSC releases that are supported for five years under mainstream
support and five more years under extended support.

Telegraf intent: *Support current semi-annual channel releases
supported by Microsoft and long-term releases under mainstream
support*

Microsoft publishes Windows Server [release information][ws-history]
and [servicing channels][ws-channels].

As of April 2021, Server 2016 (version 1607) and Server 2019 (version
1809) are LTSC releases under mainstream support and versions 1909,
2004, and 20H2 are supported SAC releases.

[ws-history]: https://docs.microsoft.com/en-us/windows-server/get-started/windows-server-release-info
[ws-channels]: https://docs.microsoft.com/en-us/windows-server/get-started-19/servicing-channels-19

## macOS

MacOS makes one major release a year and provides support for each
release for three years.

Telegraf intent: *Support releases supported by Apple*

Release history is available from [wikipedia][wp-macos].

As of April 2021, 10.14, 10.15, and 11 are supported.

[wp-macos]: https://en.wikipedia.org/wiki/MacOS#Release_history
