CREATE TABLE "expected_metric_one" (
    "timestamp" TIMESTAMP,
    "tag_one" VARCHAR2(4000),
    "tag_two" VARCHAR2(4000),
    "int64_one" NUMBER(38),
    "int64_two" NUMBER(38),
    "bool_one" BOOLEAN,
    "bool_two" BOOLEAN,
    "uint64_one" NUMBER(38),
    "float64_one" NUMBER
);
INSERT INTO "expected_metric_one" VALUES (TO_TIMESTAMP('2021-05-17 22:04:45', 'YYYY-MM-DD HH24:MI:SS'),'tag1','tag2',1234,2345,1,0,1000000000,3.1415);