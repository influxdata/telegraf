CREATE TABLE IF NOT EXISTS default.metric_one (
    tag_one String,
    tag_two String,
    int64_one Int64,
    int64_two Int64,
    timestamp Int64
) ENGINE MergeTree() ORDER BY timestamp;

INSERT INTO default.metric_one (
    tag_one,
    tag_two,
    int64_one,
    int64_two,
    timestamp
) VALUES ('tag1', 'tag2', 1234, 2345, 1621289085);
