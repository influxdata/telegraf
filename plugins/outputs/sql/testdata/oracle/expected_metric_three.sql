CREATE TABLE "expected_metric three" (
    "timestamp" TIMESTAMP,
    "tag four" VARCHAR2(4000),
    "string two" VARCHAR2(4000)
);
INSERT INTO "expected_metric three" VALUES (TO_TIMESTAMP('2021-05-17 22:04:45', 'YYYY-MM-DD HH24:MI:SS'), 'tag4', 'string2');