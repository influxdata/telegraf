# Supported Platforms

This doc helps define the platform support for Telegraf. See the
[install guide][] for specific options for installing Telegraf.

Bug reports should be submitted only for supported platforms that are under
general support, not extended or paid support. In general, Telegraf supports
Linux, macOS, Microsoft Windows, and FreeBSD.

Telegraf is written in Go, which supports many operating systems. Golang.org
has a [table][go-table] of valid OS and architecture combinations and the Go
Wiki has more specific [minimum requirements][go-reqs] for Go itself. Telegraf
may work and produce builds for other operating systems and users are welcome to
build their own binaries for them. Again, bug reports must be made on a
supported platform.

[install guide]: /docs/INSTALL_GUIDE.md
[go-table]: https://golang.org/doc/install/source#environment
[go-reqs]: https://github.com/golang/go/wiki/MinimumRequirements#operating-systems

## FreeBSD

Telegraf supports releases under FreeBSD security support. See the
[FreeBSD security page][] for specific versions.

[FreeBSD security page]: https://www.freebsd.org/security/#sup

## Linux

Telegraf will support the latest generally supported versions of major linux
distributions. This does not include extended supported releases where customers
can pay for additional support.

Below are some of the major distributions and the intent to support:

* [Debian][]: Releases supported by security and release teams
* [Fedora][]: Releases currently supported by Fedora team
* [Red Hat Enterprise Linux][]: Releases under full support
* [Ubuntu][]: Releases, interim and LTS, releases in standard support

[Debian]: https://wiki.debian.org/LTS
[Fedora]: https://fedoraproject.org/wiki/Releases
[Red Hat Enterprise Linux]: https://access.redhat.com/support/policy/updates/errata#Life_Cycle_Dates
[Ubuntu]: https://ubuntu.com/about/release-cycle

## macOS

Telegraf supports releases supported by Apple. Release history is available from
[wikipedia][wp-macos].

[wp-macos]: https://endoflife.date/macos

## Microsoft Windows

Telegraf intends to support current versions of [Windows][] and
[Windows Server][]. The release must be under mainstream or generally supported
and not under any paid or extended security support.

[Windows]: https://learn.microsoft.com/en-us/lifecycle/faq/windows
[Windows Server]: https://learn.microsoft.com/en-us/windows-server/get-started/windows-server-release-info
