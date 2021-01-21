#!/bin/sh

## Example in bash using STDIN signaling

counter=0

while read LINE; do
    echo "counter_bash count=${counter}"
    counter=$((counter+1))
done

trap "echo terminate 1>&2" EXIT