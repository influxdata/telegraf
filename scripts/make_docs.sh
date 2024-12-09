#!/bin/sh

make docs

if [ "$(git status --porcelain | wc -l)" -eq "0" ]; then
  echo "🟢 Git repo is clean."
else
  echo "🔴 Git repo dirty. Please run \"make docs\" and push the updated README. Failing CI."
  exit 1
fi
