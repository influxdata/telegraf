#!/bin/bash

BRANCH="$(git rev-parse --abbrev-ref HEAD)"
echo $BRANCH
if [[ "$BRANCH" != "master" ]]; then
    git diff master --name-only --no-color | egrep -e "^(\.circleci\/.*)|(.*\.(go|mod|sum))|Makefile$" || circleci step halt;
fi
