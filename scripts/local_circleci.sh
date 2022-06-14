#!/bin/sh

jobName=$1

circleci config process .circleci/config.yml > process.yml
circleci local execute -c process.yml --job $jobName
