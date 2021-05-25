SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;
SET default_tablespace = '';
SET default_table_access_method = heap;
CREATE TABLE public.metric_one (
    "timestamp" timestamp without time zone,
    tag_one text,
    tag_two text,
    int64_one integer,
    int64_two integer
);
ALTER TABLE public.metric_one OWNER TO postgres;
CREATE TABLE public.metric_two (
    "timestamp" timestamp without time zone,
    tag_three text,
    string_one text
);
ALTER TABLE public.metric_two OWNER TO postgres;
COPY public.metric_one ("timestamp", tag_one, tag_two, int64_one, int64_two) FROM stdin;
2021-05-17 16:04:45	tag1	tag2	1234	2345
\.
COPY public.metric_two ("timestamp", tag_three, string_one) FROM stdin;
2021-05-17 16:04:45	tag3	string1
\.
