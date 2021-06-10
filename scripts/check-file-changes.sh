#!/bin/bash
# To prevent the tests/builds to run for only a doc change, this script checks what files have changed in a pull request.

BRANCH="$(git rev-parse --abbrev-ref HEAD)"
echo $BRANCH
if [[ "$BRANCH" != "master" ]] && [[ "$BRANCH" != release* ]]; then # This should never skip for master and release branches
    # Ask git for all the differences between this branch and master
    # Then use grep to look for changes in the .circleci/ directory, anything named *.go or *.mod or *.sum or *.sh or Makefile
    # If no match is found, then circleci step halt will stop the CI job but mark it successful
    git diff master --name-only --no-color | egrep -e "^(\.circleci\/.*)$|^(.*\.(go|mod|sum|sh))$|^Makefile$" || circleci step halt;
fi
