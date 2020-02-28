#!/bin/bash

## Example in bash using STDIN signaling

counter=0

while read; do
    echo "counter_bash count=${counter}"
    let counter=counter+1
done

(>&2 echo "terminate")
