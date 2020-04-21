#!/bin/bash

## Example in bash using STDIN signaling

counter=0

while read LINE; do
    echo "counter_bash count=${counter}"
    counter=`expr $counter + 1`
done

(>&2 echo "terminate")