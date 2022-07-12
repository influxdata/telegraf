#!/bin/sh

make docs

if [ "$(git status --porcelain | wc -l)" -eq "0" ]; then
  echo "🟢 Git repo is clean."
else
  echo "🔴 Git repo dirty. Quit."
  exit 1
fi
