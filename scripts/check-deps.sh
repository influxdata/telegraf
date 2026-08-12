#!/bin/bash

tmpdir="$(mktemp -d)"

cleanup() {
	rm -rf "$tmpdir"
}
trap cleanup EXIT

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
    CGO_ENABLED=0 GOOS=${os} GOARCH=${arch} GOARM=${arm} \
      go list -f '{{with .Module}}{{.Path}}{{end}}' -deps ./cmd/telegraf >> "${tmpdir}/golist"
  else
    echo "GOOS=${os} GOARCH=${rest}"
    CGO_ENABLED=0 GOOS=${os} GOARCH=${rest} \
      go list -f '{{with .Module}}{{.Path}}{{end}}' -deps ./cmd/telegraf >> "${tmpdir}/golist"
  fi
done

LC_ALL=C sort -u < "${tmpdir}/golist" | while IFS= read -r dep; do
	case "${dep}" in
		# Denylist: modules that must never appear in the shipped binary.
		# These are test-only and shipping them bloats the binary and can
		# add startup cost (e.g. an init() that scans the process table).
		github.com/testcontainers/*|github.com/stretchr/testify/*)
			echo "FORBIDDEN DEPENDENCY: ${dep}" >> "${tmpdir}/denylist_hits";;

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

# Fail the build if any test-only module leaked into the shipped binary.
if [[ -s "${tmpdir}/denylist_hits" ]]; then
	cat - <<EOF


The telegraf binary must not depend on test-only modules. The following
modules were found in the dependency graph and are forbidden:

$(sort -u < "${tmpdir}/denylist_hits")

Move the imports into _test.go files or a test-only package so they are
not linked into the shipped binary.
EOF
	exit 1
fi

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
