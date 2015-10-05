#!/bin/sh

go clean

for file in *.go
do
    echo -n "Compiling $file ..."
    go build "$file"
    echo " done."
done
