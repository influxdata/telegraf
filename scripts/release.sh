#!/bin/sh
#
# usage: release.sh BUILD_NUM
#
# Requirements:
# - curl
# - jq
# - sha256sum
# - awscli
# - gpg
#
# CIRCLE_TOKEN set to a CircleCI API token that can list the artifacts.
#
# AWS cli setup to be able to write to the BUCKET.
#
# GPG setup with a signing key.

BUILD_NUM="${1:?usage: release.sh BUILD_NUM}"
BUCKET="${2:-dl.influxdata.com/telegraf/releases}"

: ${CIRCLE_TOKEN:?"Must set CIRCLE_TOKEN"}

tmpdir="$(mktemp -d -t telegraf.XXXXXXXXXX)"

on_exit() {
	rm -rf "$tmpdir"
}
trap on_exit EXIT

echo "${tmpdir}"
cd "${tmpdir}" || exit 1

curl -s -S -L -H Circle-Token:${CIRCLE_TOKEN} \
	"https://circleci.com/api/v2/project/gh/influxdata/telegraf/${BUILD_NUM}/artifacts" \
	-o artifacts || exit 1

cat artifacts | jq -r '.items[] | "\(.url) \(.path|ltrimstr("build/dist/"))"' > manifest

while read url path;
do
	echo $url
	curl -s -S -L -o "$path" "$url" &&
	sha256sum "$path" > "$path.DIGESTS" &&
	gpg --armor --detach-sign "$path.DIGESTS" &&
	gpg --armor --detach-sign "$path" || exit 1
done < manifest

echo
cat *.DIGESTS
echo

arch() {
	case ${1} in
		*i386.*)
			echo i386;;
		*armel.*)
			echo armel;;
		*armv6hl.*)
			echo armv6hl;;
		*armhf.*)
			echo armhf;;
		*arm64.* | *aarch64.*)
			echo arm64;;
		*amd64.* | *x86_64.*)
			echo amd64;;
		*s390x.*)
			echo s390x;;
		*mipsel.*)
			echo mipsel;;
		*mips.*)
			echo mips;;
		*)
			echo unknown
	esac
}

platform() {
	case ${1} in
		*".rpm")
			echo Centos;;
		*".deb")
			echo Debian;;
		*"linux"*)
			echo Linux;;
		*"freebsd"*)
			echo FreeBSD;;
		*"darwin"*)
			echo Mac OS X;;
		*"windows"*)
			echo Windows;;
		*)
			echo unknown;;
	esac
}

echo "Arch | Platform | Package | SHA256"
echo "---| --- | --- | ---"
while read url path;
do
	echo "$(arch ${path}) | $(platform ${path}) | [\`${path}\`](https://dl.influxdata.com/telegraf/releases/${path}) | \`$(sha256sum ${path} | cut -f1 -d' ')\`"
done < manifest
echo ""

package="$(grep *_amd64.deb manifest | cut -f2 -d' ')"
cat -<<EOF
      {
        "platform": "Ubuntu &amp; Debian",
        "sha256":"$(sha256sum $package | cut -f1 -d' ')",
        "code":[
          "wget https://dl.influxdata.com/telegraf/releases/$package",
          "sudo dpkg -i $package"
        ]
      },
EOF
package="$(grep *.x86_64.rpm manifest | cut -f2 -d' ')"
cat -<<EOF
      {
        "platform": "RedHat &amp; CentOS",
        "sha256":"$(sha256sum $package | cut -f1 -d' ')",
        "code":[
          "wget https://dl.influxdata.com/telegraf/releases/$package",
          "sudo yum localinstall $package"
        ]
      },
EOF
package="$(grep *windows_amd64.zip manifest | cut -f2 -d' ')"
cat -<<EOF
      {
        "platform": "Windows Binaries (64-bit)",
        "sha256":"$(sha256sum $package | cut -f1 -d' ')",
        "code":[
          "wget https://dl.influxdata.com/telegraf/releases/$package",
          "unzip $package"
        ]
      },
EOF
package="$(grep *_linux_amd64.tar.gz manifest | cut -f2 -d' ')"
cat -<<EOF
      {
        "platform": "Linux Binaries (64-bit)",
        "sha256":"$(sha256sum $package | cut -f1 -d' ')",
        "code":[
          "wget https://dl.influxdata.com/telegraf/releases/$package",
          "tar xf $package"
        ]
      },
EOF
package="$(grep *linux_i386.tar.gz manifest | cut -f2 -d' ')"
cat -<<EOF
      {
        "platform": "Linux Binaries (32-bit)",
        "sha256":"$(sha256sum $package | cut -f1 -d' ')",
        "code":[
          "wget https://dl.influxdata.com/telegraf/releases/$package",
          "tar xf $package"
        ]
      },
EOF
package="$(grep *linux_armhf.tar.gz manifest | cut -f2 -d' ')"
cat -<<EOF
      {
        "platform": "Linux Binaries (ARM)",
        "sha256":"$(sha256sum $package | cut -f1 -d' ')",
        "code":[
          "wget https://dl.influxdata.com/telegraf/releases/$package",
          "tar xf $package"
        ]
      }
EOF

aws s3 sync ./ "s3://$BUCKET/" \
	--exclude "*" \
	--include "*.tar.gz" \
	--include "*.deb" \
	--include "*.rpm" \
	--include "*.zip" \
	--include "*.DIGESTS" \
	--include "*.asc" \
	--acl public-read
