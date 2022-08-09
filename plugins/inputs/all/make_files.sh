#!/bin/bash

#run this in the same directory as all.go

regex='^	_ "([^"]*)"$'
#line='_ "github.com/influxdata/telegraf/plugins/inputs/synproxy"'
while IFS= read -r line; do
    if [[ ! $line =~ $regex ]]; then
	#don't print errors for lines that aren't imports
	#echo no match: $line
	continue
    fi

    mod="${BASH_REMATCH[1]}"
    #echo "$mod"
    IFS='/' read -ra segments <<< "$mod"
    plugin="${segments[-1]}"
    #echo "$plugin"

    if [ ! -d "../$plugin" ]; then
	echo dir not found: "$plugin"
	continue
    fi

    IFS= cat > "$plugin.go" <<EOF
//go:build all || inputs || inputs.$plugin

package all

import (
$line
)
EOF
done < all.go-old
