#!/bin/bash

plugin="${1}"

if [ ! -d "${plugin}" ]; then
    echo "ERR: ${plugin} is not a directory"
    exit 1
fi

pluginname=$(basename "${plugin}")
if [[ ":all:" =~ .*:${pluginname}:.* ]]; then
    echo "INF: ${plugin} ignored"
    exit 0
fi

# Check for the sample.conf file
if [ ! -f "${plugin}/sample.conf" ] && [ "${pluginname}" != "modbus" ]; then
    echo "ERR: ${plugin} does not contain a sample.conf file"
    exit 1
fi

# Check for the sample.conf embedding into the README.md
readme="^\`\`\`toml @sample.*\.conf\b"
if  ! grep -q "${readme}" "${plugin}/README.md" ; then
    echo "ERR: ${plugin} is missing embedding in README"
    exit 1
fi

# Check for the generator
generator="//go:generate ../../../tools/readme_config_includer/generator"
found=false
for filename in "${plugin}/"*.go; do
    if [[ "${filename}" == *_test.go ]]; then
        continue
    fi
    if ! grep -q "SampleConfig(" "${filename}"; then
        continue
    fi

    if grep -q "^${generator}\$" "${filename}"; then
        found=true
        break
    fi
done

if ! ${found}; then
    echo "ERR: ${plugin} is missing generator statement!"
    exit 1
fi

# Check for the embedding
embedding="//go:embed sample.*\.conf"
found=false
for filename in "${plugin}/"*.go; do
    if [[ "${filename}" == *_test.go ]]; then
        continue
    fi
    if ! grep -q "SampleConfig(" "${filename}"; then
        continue
    fi

    if grep -q "^${embedding}\$" "${filename}"; then
        found=true
        break
    fi
done

if ! ${found}; then
    echo "ERR: ${plugin} is missing embedding statement!"
    exit 1
fi