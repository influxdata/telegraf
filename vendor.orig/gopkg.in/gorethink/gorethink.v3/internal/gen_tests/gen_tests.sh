#!/usr/bin/env bash

set -e

if [[ $REQL_TEST_DIR == "" ]]
then
    echo "\$REQL_TEST_DIR must be specified"
    exit 1
fi

../gen_tests/gen_tests.py --test-dir=$REQL_TEST_DIR

goimports -w . > /dev/null

exit 0
