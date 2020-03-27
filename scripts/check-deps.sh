#!/bin/sh

tmpdir="$(mktemp -d)"

cleanup() {
	rm -rf "$tmpdir"
}
trap cleanup EXIT

targets="$(go tool dist list)"

for target in ${targets}; do
	# only check platforms we build for
	case "${target}" in
		linux/*) ;;
		windows/*) ;;
		freebsd/*) ;;
		darwin/*) ;;
		*) continue;;
	esac

	GOOS=${target%%/*} GOARCH=${target##*/} \
		go list -deps -f '{{with .Module}}{{.Path}}{{end}}' ./cmd/telegraf/ >> "${tmpdir}/golist"
done

for dep in $(LC_ALL=C sort -u "${tmpdir}/golist"); do
	case "${dep}" in
		# ignore ourselves
		github.com/influxdata/telegraf) continue;;

		# dependency is replaced in go.mod
		github.com/satori/go.uuid) continue;;

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

grep '^-' docs/LICENSE_OF_DEPENDENCIES.md | grep -v github.com/DataDog/datadog-agent | cut -f 2 -d' ' > "${tmpdir}/LICENSE_OF_DEPENDENCIES.md"

diff -U0 "${tmpdir}/LICENSE_OF_DEPENDENCIES.md" "${tmpdir}/HEAD" ||
cat - <<EOF


The docs/LICENSE_OF_DEPENDENCIES.md file does not contain the expected entries.

Lines prefixed with '+' should be added to LICENSE_OF_DEPENDENCIES.md and '-'
lines should be removed.

Include a link to the appropriate licenses for any additions.
EOF
