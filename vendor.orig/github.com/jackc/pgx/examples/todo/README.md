# Description

This is a sample todo list implemented using pgx as the connector to a
PostgreSQL data store.

# Usage

Create a PostgreSQL database and run structure.sql into it to create the
necessary data schema.

Example:

    createdb todo
    psql todo < structure.sql

Build todo:

    go build

## Connection configuration

The database connection is configured via enviroment variables.

* PGHOST - defaults to localhost
* PGUSER - defaults to current OS user
* PGPASSWORD - defaults to empty string
* PGDATABASE - defaults to user name

You can either export them then run todo:

    export PGDATABASE=todo
    ./todo list

Or you can prefix the todo execution with the environment variables:

    PGDATABASE=todo ./todo list

## Add a todo item

    ./todo add 'Learn go'

## List tasks

    ./todo list

## Update a task

    ./todo add 1 'Learn more go'

## Delete a task

    ./todo remove 1

# Example Setup and Execution

    jack@hk-47~/dev/go/src/github.com/jackc/pgx/examples/todo$ createdb todo
    jack@hk-47~/dev/go/src/github.com/jackc/pgx/examples/todo$ psql todo < structure.sql
    Expanded display is used automatically.
    Timing is on.
    CREATE TABLE
    Time: 6.363 ms
    jack@hk-47~/dev/go/src/github.com/jackc/pgx/examples/todo$ go build
    jack@hk-47~/dev/go/src/github.com/jackc/pgx/examples/todo$ export PGDATABASE=todo
    jack@hk-47~/dev/go/src/github.com/jackc/pgx/examples/todo$ ./todo list
    jack@hk-47~/dev/go/src/github.com/jackc/pgx/examples/todo$ ./todo add 'Learn Go'
    jack@hk-47~/dev/go/src/github.com/jackc/pgx/examples/todo$ ./todo list
    1. Learn Go
    jack@hk-47~/dev/go/src/github.com/jackc/pgx/examples/todo$ ./todo update 1 'Learn more Go'
    jack@hk-47~/dev/go/src/github.com/jackc/pgx/examples/todo$ ./todo list
    1. Learn more Go
    jack@hk-47~/dev/go/src/github.com/jackc/pgx/examples/todo$ ./todo remove 1
    jack@hk-47~/dev/go/src/github.com/jackc/pgx/examples/todo$ ./todo list
