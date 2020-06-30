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

curl -s -S -H Circle-Token:${CIRCLE_TOKEN} \
	"https://circleci.com/api/v2/project/gh/influxdata/telegraf/${BUILD_NUM}/artifacts" \
	-o artifacts || exit 1

cat artifacts | jq -r '.items[] | "\(.url) \(.path|ltrimstr("build/dist/"))"' > manifest

while read url path;
do
	echo $url
	curl -s -S -o "$path" "$url" &&
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

aws s3 sync ./ "s3://$BUCKET/" \
	--exclude "*" \
	--include "*.tar.gz" \
	--include "*.deb" \
	--include "*.rpm" \
	--include "*.zip" \
	--include "*.asc" \
	--acl public-read
