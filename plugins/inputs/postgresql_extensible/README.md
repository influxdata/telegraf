# PostgreSQL plugin

This postgresql plugin provides metrics for your postgres database. It has been designed to parse a sql query json file with some parameters.

For now the plugin only support one postgresql instance, the plan is to be able to extend easily your postgres monitoring.



View to create :
-- View: public.sessions

DROP VIEW public.sessions;

CREATE OR REPLACE VIEW public.sessions AS
 WITH proctab AS (
         SELECT pg_proctab.pid,
                CASE
                    WHEN pg_proctab.state::text = 'R'::bpchar THEN 'running'::text
                    WHEN pg_proctab.state::text = 'D'::bpchar THEN 'sleep-io'::text
                    WHEN pg_proctab.state::text = 'S'::bpchar THEN 'sleep-waiting'::text
                    WHEN pg_proctab.state::text = 'Z'::bpchar THEN 'zombie'::text
                    WHEN pg_proctab.state::text = 'T'::bpchar THEN 'stopped'::text
                    ELSE NULL::text
                END AS proc_state,
            pg_proctab.ppid,
            pg_proctab.utime,
            pg_proctab.stime,
            pg_proctab.vsize,
            pg_proctab.rss,
            pg_proctab.processor,
            pg_proctab.rchar,
            pg_proctab.wchar,
            pg_proctab.syscr,
            pg_proctab.syscw,
            pg_proctab.reads,
            pg_proctab.writes,
            pg_proctab.cwrites
           FROM pg_proctab() pg_proctab
        ), stat_activity AS (
         SELECT pg_stat_activity.datname,
            pg_stat_activity.pid,
            pg_stat_activity.usename,
                CASE
                    WHEN pg_stat_activity.query IS NULL THEN 'no query'::text
                    ELSE regexp_replace(pg_stat_activity.query, '[\n\r]+'::text, ' '::text, 'g'::text)
                END AS query
           FROM pg_stat_activity
        )
 SELECT ('"'::text || stat.datname::text) || '"'::text AS db,
    ('"'::text || stat.usename::text) || '"'::text as username,
    stat.pid AS pid,
    ('"'::text || proc.proc_state ::text) || '"'::text AS state,
    ('"'::text || stat.query::text) || '"'::text AS query,
    proc.utime AS session_usertime,
    proc.stime AS session_systemtime,
    proc.vsize AS session_virtual_memory_size,
    proc.rss AS session_resident_memory_size,
    proc.processor AS session_processor_number,
    proc.rchar AS session_bytes_read,
    proc.wchar AS session_bytes_written,
    proc.syscr AS session_read_io,
    proc.syscw AS session_write_io,
    proc.reads AS session_physical_reads,
    proc.writes AS session_physical_writes,
    proc.cwrites AS session_cancel_writes
   FROM proctab proc,
    stat_activity stat
  WHERE proc.pid = stat.pid;

ALTER TABLE public.sessions
  OWNER TO postgres;
