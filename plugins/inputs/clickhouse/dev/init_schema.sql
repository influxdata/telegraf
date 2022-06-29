DROP TABLE IF EXISTS default.test;
CREATE TABLE default.test(
    Nom String,
    Code Nullable(String) DEFAULT Null,
    Cur Nullable(String) DEFAULT Null
) ENGINE=MergeTree() ORDER BY tuple();
