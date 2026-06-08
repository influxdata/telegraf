CREATE TABLE "expected_metric_two" (
    "timestamp" TIMESTAMP,
    "tag_three" VARCHAR2(4000),
    "string_one" VARCHAR2(4000)
);
INSERT INTO "expected_metric_two" VALUES (TO_TIMESTAMP('2021-05-17 22:04:45', 'YYYY-MM-DD HH24:MI:SS'), 'tag3', 'string1');