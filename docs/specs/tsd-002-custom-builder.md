# Telegraf Custom-Builder

## Objective

Provide a tool to build a customized, smaller version of Telegraf with only
the required plugins included.

## Keywords

tool, binary size, customization

## Overview

The Telegraf binary continues to grow as new plugins and features are added
and dependencies are updated. Users running on resource constrained systems
such as embedded-systems or inside containers might suffer from the growth.

This document specifies a tool to build a smaller Telegraf binary tailored to
the plugins configured and actually used, removing unnecessary and unused
plugins. The implementation should be able to cope with configured parsers and
serializers including defaults for those plugin categories. Valid Telegraf
configuration files, including directories containing such files, are the input
to the customization process.

The customization tool might not be available for older versions of Telegraf.
Furthermore, the degree of customization and thus the effective size reduction
might vary across versions. The tool must create a single static Telegraf
binary. Distribution packages or containers are *not* targeted.

## Prior art

[PR #5809](https://github.com/influxdata/telegraf/pull/5809) and
[telegraf-lite-builder](https://github.com/influxdata/telegraf/tree/telegraf-lite-builder/cmd/telegraf-lite-builder):

- Uses docker
- Uses browser:
  - Generates a webpage to pick what options you want. User chooses plugins;
    does not take a config file
  - Build a binary, then minifies by stripping and compressing that binary
- Does some steps that belong in makefile, not builder
  - Special case for upx
  - Makes gzip, zip, tar.gz
- Uses gopkg.in?
- Can also work from the command line

[PR #8519](https://github.com/influxdata/telegraf/pull/8519)

- User chooses plugins OR provides a config file

[powers/telegraf-build](https://github.com/powersj/telegraf-build)

- User chooses plugins OR provides a config file
- Currently kept in separate repo
- Undoes changes to all.go files

[rawkode/bring-your-own-telegraf](https://github.com/rawkode/bring-your-own-telegraf)

- Uses docker

## Additional information

You might be able to further reduce the binary size of Telegraf by removing
debugging information. This is done by adding `-w` and `-s` to the linker flags
before building `LDFLAGS="-w -s"`.

However, please note that this removes information helpful for debugging issues
in Telegraf.

Additionally, you can use a binary packer such as [UPX](https://upx.github.io/)
to reduce the required *disk* space. This compresses the binary and decompresses
it again at runtime. However, this does not reduce memory footprint at runtime.
