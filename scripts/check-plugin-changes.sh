#!/bin/bash

echo "=== Plugin Change Detection Script ===" >&2
echo "Current directory: $(pwd)" >&2
echo "CIRCLE_PULL_REQUEST: ${CIRCLE_PULL_REQUEST}" >&2

PLUGINS='^plugins/(inputs|outputs|aggregators|processors)/.*$'

# Check if we're in a pull request
if [[ -z "${CIRCLE_PULL_REQUEST##*/}" ]]; then
    echo "Not in a pull request context, running all tests" >&2
    echo "./..."
    exit 0
fi

# If anything outside the supported plugins changed we need to test everything
MODIFIED_NONPLUGINS=$(git diff --name-only "$(git merge-base HEAD origin/master)" | grep -v -E "${PLUGINS}")
echo "modified: ${MODIFIED_NONPLUGINS}">&2
if [[ -n "${MODIFIED_NONPLUGINS}" ]]; then
    echo "Modified files outside plugins detected, running all tests" >&2
    echo "./..."
    exit 0
fi

# Get the plugins modified and selectively run that plugin if only one is touched
MODIFIED_PLUGINS=$(IFS='' git diff --name-only "$(git merge-base HEAD origin/master)" | grep -E "${PLUGINS}" | awk -F'/' '{print $1"/"$2"/"$3}' | sort -u)
echo "=== Changed plugins ===" >&2
echo "${MODIFIED_PLUGINS}" >&2
echo "=====================" >&2

MODIFIED_PLUGIN_COUNT="$(echo "${MODIFIED_PLUGINS}" | wc -l)"
if [[ "${MODIFIED_PLUGIN_COUNT}" -ne 1 ]]; then
    echo "Found ${MODIFIED_PLUGIN_COUNT} modified plugins, running all tests" >&2
    echo "./..."
    exit 0
fi

# Make sure we don't target a plugin import file
if [[ "${MODIFIED_PLUGINS##*/}" == "all" ]]; then
    echo "Found modified \"all\" file(s), running all tests" >&2
    echo "./..."
    exit 0
fi

# Make sure the plugin dir exists
if [ ! -d "${MODIFIED_PLUGINS}" ]; then
    echo "Warning: ${MODIFIED_PLUGINS} is not a directory or does not exist, running all tests" >&2
    echo "./..."
    exit 0
fi

echo "Changes detected in \"${MODIFIED_PLUGINS}\" only, running selective test" >&2

echo "=== Script execution completed ===" >&2
echo "./${MODIFIED_PLUGINS}"
exit 0
