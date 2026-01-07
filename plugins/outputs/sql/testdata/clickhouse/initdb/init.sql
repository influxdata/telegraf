CREATE DATABASE foo;

-- Pre-create a table to test existing table column detection
CREATE TABLE foo.pre_existing_table (
    `timestamp` DateTime,
    `tag_one` String
) ENGINE=MergeTree ORDER BY (timestamp);
