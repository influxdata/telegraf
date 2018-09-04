# Description

This is a sample chat program implemented using PostgreSQL's listen/notify
functionality with pgx.

Start multiple instances of this program connected to the same database to chat
between them.

## Connection configuration

The database connection is configured via the standard PostgreSQL environment variables.

* PGHOST - defaults to localhost
* PGUSER - defaults to current OS user
* PGPASSWORD - defaults to empty string
* PGDATABASE - defaults to user name

You can either export them then run chat:

    export PGHOST=/private/tmp
    ./chat

Or you can prefix the chat execution with the environment variables:

    PGHOST=/private/tmp ./chat
