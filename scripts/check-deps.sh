#!/bin/bash

tmpdir="$(mktemp -d)"

cleanup() {
	rm -rf "$tmpdir"
	rm telegraf
}
trap cleanup EXIT

go install github.com/JoakimSoderberg/gobindep@latest

declare -a targets=(
  "darwin/amd64"
  "darwin/arm64"
  "freebsd/amd64"
  "freebsd/arm/7"
  "freebsd/386"
  "linux/amd64"
  "linux/arm64/7"
  "linux/arm/5"
  "linux/arm/6"
  "linux/386"
  "linux/mips"
  "linux/mipsle"
  "linux/ppc64le"
  "linux/riscv64"
  "linux/s390x"
  "windows/amd64"
  "windows/arm64"
  "windows/386"
)

for target in "${targets[@]}"; do
  os="${target%%/*}"
  rest="${target#*/}"

  if [[ "$rest" == */* ]]; then
    arch="${rest%%/*}"
    arm="${rest#*/}"

    echo "GOOS=${os} GOARCH=${arch} GOARM=${arm}"
    GOOS=${os} GOARCH=${arch} GOARM=${arm} make telegraf > /dev/null 2>&1
  else
    echo "GOOS=${os} GOARCH=${rest}"
    GOOS=${os} GOARCH=${rest} make telegraf > /dev/null 2>&1
  fi
  gobindep telegraf | sed 's/ .*//' >> "${tmpdir}/golist"
done

LC_ALL=C sort -u < "${tmpdir}/golist" | while IFS= read -r dep; do
	case "${dep}" in
		# ignore ourselves
		github.com/influxdata/telegraf) continue;;

		# go-autorest has a single license for all sub modules
		github.com/Azure/go-autorest/autorest)
			dep=github.com/Azure/go-autorest;;
		github.com/Azure/go-autorest/*)
			continue;;

		# single license for all sub modules
		cloud.google.com/go/*)
			continue;;
	esac

	# Remove single and double digit version from path; these are generally not
	# actual parts of the path and instead indicate a branch or tag.
	#   example: github.com/influxdata/go-syslog/v2 -> github.com/influxdata/go-syslog
	dep="${dep%%/v[0-9]}"
	dep="${dep%%/v[0-9][0-9]}"

	echo "${dep}" >> "${tmpdir}/HEAD"
done

# If there are two versions of a library that have the same base (like
# github.com/foo/bar github.com/foo/bar/v3) there will be a duplicate
# in the list.  Remove duplicates again.
mv "${tmpdir}/HEAD" "${tmpdir}/HEAD-dup"
uniq "${tmpdir}/HEAD-dup" > "${tmpdir}/HEAD"

grep '^-' docs/LICENSE_OF_DEPENDENCIES.md | grep -v github.com/DataDog/datadog-agent | cut -f 2 -d' ' > "${tmpdir}/LICENSE_OF_DEPENDENCIES.md"

diff -U0 "${tmpdir}/LICENSE_OF_DEPENDENCIES.md" "${tmpdir}/HEAD" || {
cat - <<EOF


The docs/LICENSE_OF_DEPENDENCIES.md file does not contain the expected entries.

Lines prefixed with '+' should be added to LICENSE_OF_DEPENDENCIES.md and '-'
lines should be removed.

Include a link to the appropriate licenses for any additions.
EOF
exit 1
}
