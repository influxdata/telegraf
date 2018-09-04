#!/usr/bin/env bash
set -eux

mv conn_config_test.go.travis conn_config_test.go

if [ "${PGVERSION-}" != "" ]
then
  # The tricky test user, below, has to actually exist so that it can be used in a test
  # of aclitem formatting. It turns out aclitems cannot contain non-existing users/roles.
  psql -U postgres -c 'create database pgx_test'
  psql -U postgres pgx_test -c 'create extension hstore'
  psql -U postgres -c "create user pgx_ssl SUPERUSER PASSWORD 'secret'"
  psql -U postgres -c "create user pgx_md5 SUPERUSER PASSWORD 'secret'"
  psql -U postgres -c "create user pgx_pw  SUPERUSER PASSWORD 'secret'"
  psql -U postgres -c "create user pgx_replication with replication password 'secret'"
  psql -U postgres -c "create user \" tricky, ' } \"\" \\ test user \" superuser password 'secret'"
fi
