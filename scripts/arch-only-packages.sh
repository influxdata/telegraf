#!/bin/bash

# Print the packages containing Go files that build for the given GOOS/GOARCH
# but not for GOOS/amd64, so a lint run for the architecture can be limited to
# what the amd64 run does not cover. The selection is based on the build
# constraints the Go tool evaluates, so it catches files selected by a
# filename suffix (e.g. pdh_arm64.go) as well as by a "//go:build" line.
set -euo pipefail

if [[ $# -ne 2 ]]; then
    echo "usage: $0 GOOS GOARCH" >&2
    exit 1
fi
os="$1"
arch="$2"

# List all Go files (including tests) of all packages built for the architecture.
# The "$d" below is a Go template variable, not a shell one.
files() {
    # shellcheck disable=SC2016
    GOOS="${os}" GOARCH="$1" go list -e -f '{{$d:=.Dir}}{{range .GoFiles}}{{$d}}/{{.}}{{"\n"}}{{end}}{{range .TestGoFiles}}{{$d}}/{{.}}{{"\n"}}{{end}}{{range .XTestGoFiles}}{{$d}}/{{.}}{{"\n"}}{{end}}' ./... | sort
}

comm -13 <(files amd64) <(files "${arch}") | while read -r file; do
    dirname "${file}"
done | sort -u | sed "s|^$(pwd)|.|"
