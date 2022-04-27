#!/bin/bash
set -euo pipefail
IFS=$'\n\t'

usage () {
    echo "Check that no dynamic symbols provided by glibc are newer than a given version"
    echo "Usage:"
    echo "   $0 program version"
    echo "where program is the elf binary to check and version is a dotted version string like 2.3.4"
    exit 1
}

#validate input and display help
[[ $# = 2 ]] || usage
prog=$1
max=$2

#make sure dependencies are installed
have_deps=true
for i in objdump grep sort uniq sed; do
    if ! command -v "$i" > /dev/null; then
	echo "$i not in path"
	have_deps=false
    fi
done
if [[ $have_deps = false ]]; then
    exit 1
fi

#compare dotted versions
#see https://stackoverflow.com/questions/4023830/how-to-compare-two-strings-in-dot-separated-version-format-in-bash
vercomp () {
    if [[ $1 == $2 ]]
    then
        return 0
    fi
    local IFS=.
    local i ver1=($1) ver2=($2)
    # fill empty fields in ver1 with zeros
    for ((i=${#ver1[@]}; i<${#ver2[@]}; i++))
    do
        ver1[i]=0
    done
    for ((i=0; i<${#ver1[@]}; i++))
    do
        if [[ -z ${ver2[i]} ]]
        then
            # fill empty fields in ver2 with zeros
            ver2[i]=0
        fi
        if ((10#${ver1[i]} > 10#${ver2[i]}))
        then
            return 1
        fi
        if ((10#${ver1[i]} < 10#${ver2[i]}))
        then
            return 2
        fi
    done
    return 0
}

if ! objdump -p "$prog" | grep -q NEEDED; then
    echo "$prog doesn't have dynamic library dependencies"
    exit 0
fi

objdump -T "$prog" | # get the dynamic symbol table
    sed -n "s/.* GLIBC_\([0-9.]\+\).*/\1/p" | # find the entries for glibc and grab the version
    sort | uniq | # remove duplicates
    while read v; do
        set +e
        vercomp "$v" "$max" # fail if any version is newer than our max
        comp=$?
        set -e
        if [[ $comp -eq 1 ]]; then
            echo "$v is newer than $max"
            exit 1
        fi
    done

exit 0
