#!/bin/bash
# Warns when metric field names are removed or renamed in plugin READMEs.
# Metric removals break user dashboards and alerting rules; this check
# surfaces the impact so contributors can acknowledge it before merging.
#
# Always exits 0 — this is an informational warning, not a hard gate.

set -euo pipefail

BASE_REF="${GITHUB_BASE_REF:-master}"
BASE="origin/${BASE_REF}"

# Extract backtick-quoted field names from the first column of markdown tables.
# Matches lines like:  | `field_name`  | ...
extract_fields() {
    # shellcheck disable=SC2016
    grep -oE '^\s*\|\s*`[a-zA-Z0-9_]+`' "$1" 2>/dev/null \
        | awk -F'`' '{print $2}' \
        || true
}

extract_fields_from_stdin() {
    # shellcheck disable=SC2016
    grep -oE '^\s*\|\s*`[a-zA-Z0-9_]+`' 2>/dev/null \
        | awk -F'`' '{print $2}' \
        || true
}

CHANGED_READMES=$(git diff --name-only "${BASE}...HEAD" -- 'plugins/*/*/README.md' 2>/dev/null || true)

if [[ -z "${CHANGED_READMES}" ]]; then
    echo "No plugin READMEs changed — nothing to check."
    exit 0
fi

WARNED=0

while IFS= read -r readme; do
    [[ -z "$readme" ]] && continue

    if ! git show "${BASE}:${readme}" &>/dev/null; then
        # New plugin README — no baseline to compare against.
        continue
    fi

    old_fields=$(git show "${BASE}:${readme}" | extract_fields_from_stdin | sort -u)
    new_fields=$(extract_fields "$readme" | sort -u)

    removed=$(comm -23 <(echo "$old_fields") <(echo "$new_fields"))

    if [[ -n "$removed" ]]; then
        echo "::warning file=${readme}::Metric fields removed or renamed in ${readme}"
        echo ""
        echo "  Plugin: ${readme}"
        echo "  Removed fields:"
        while IFS= read -r field; do
            [[ -z "$field" ]] && continue
            echo "    - ${field}"
        done <<< "$removed"
        echo ""
        echo "  Removing or renaming existing metric fields is a breaking change."
        echo "  Users relying on these names in dashboards, alerts, or pipelines"
        echo "  will see data gaps after upgrading. If this change is intentional:"
        echo "    1. Add a deprecation notice in the README alongside the old name."
        echo "    2. Keep the old name available (e.g. via a 'legacy_*' include option)."
        echo "    3. Document the migration path in your PR description."
        echo ""
        WARNED=1
    fi
done <<< "$CHANGED_READMES"

if [[ $WARNED -eq 0 ]]; then
    echo "No metric fields removed — backward compatibility preserved."
fi

exit 0
