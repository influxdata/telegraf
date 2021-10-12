import argparse
import re
import cx_Oracle
import sys


def handle_error(error_message):
    sys.stderr.write("ERROR|" + str(error_message))
    sys.exit(1)


class OracleMetrics():

    def __init__(self, user, passwd, sid):
        self.user = user
        self.passwd = passwd
        self.sid = sid
        self.delengine = "none"
        self.connection = None
        try:
            self.connection = cx_Oracle.connect(self.user, self.passwd, self.sid)
        except cx_Oracle.DatabaseError as e :
            raise
        except Exception as e :
            raise


    def getWaitClassStats(self, user, passwd, sid):
        cursor = None
        try:
            cursor = self.connection.cursor()
            cursor.execute("""
            select n.wait_class, round(m.time_waited/m.INTSIZE_CSEC,3) AAS
            from   v$waitclassmetric  m, v$system_wait_class n
            where m.wait_class_id=n.wait_class_id and n.wait_class != 'Idle'
            union
            select  'CPU', round(value/100,3) AAS
            from v$sysmetric where metric_name='CPU Usage Per Sec' and group_id=2
            union 
            select 'CPU_OS', round((prcnt.busy*parameter.cpu_count)/100,3) - aas.cpu
            from
            ( select value busy
            from v$sysmetric
            where metric_name='Host CPU Utilization (%)'
            and group_id=2 ) prcnt,
            ( select value cpu_count from v$parameter where name='cpu_count' )  parameter,
            ( select  'CPU', round(value/100,3) cpu from v$sysmetric where metric_name='CPU Usage Per Sec' and group_id=2) aas
            """)
            for wait in cursor:
                wait_name = wait[0]
                wait_value = wait[1]
                print ("oracle_wait_class,instance={0},wait_class={1} wait_value={2}".format(sid, re.sub(' ', '_', wait_name), wait_value))
        except Exception as e :
            raise
        finally:
            if cursor is not None:
                cursor.close()


    def getSysmetrics(self, user, passwd, sid):
        cursor = None
        try:
            cursor = self.connection.cursor()
            cursor.execute("""
            select METRIC_NAME,VALUE,METRIC_UNIT from v$sysmetric where group_id=2
            """)
            for metric in cursor:
                metric_name = metric[0]
                metric_value = metric[1]
                print ("oracle_sysmetric,instance={0},metric_name={1} metric_value={2}".format(sid,re.sub(' ', '_', metric_name),metric_value))
        except Exception as e :
            raise
        finally:
            if cursor is not None:
                cursor.close()
   

    def getWaitStats(self, user, passwd, sid):
        cursor = None
        try:
            cursor = self.connection.cursor()
            cursor.execute("""
            select 
            n.wait_class wait_class,
            n.name wait_name,
            m.wait_count  cnt,
            nvl(round(10*m.time_waited/nullif(m.wait_count,0),3) ,0) avg_ms
            from v$eventmetric m,
            v$event_name n
            where m.event_id=n.event_id
            and n.wait_class <> 'Idle' and m.wait_count > 0 order by 1""")
            for wait in cursor:
                wait_class = wait[0]
                wait_name = wait[1]
                wait_cnt = wait[2]
                wait_avgms = wait[3]
                print ("oracle_wait_event,instance={0},wait_class={1},wait_event={2} count={3},latency={4}".format(sid,re.sub(' ', '_', wait_class), re.sub(' ','_',wait_name),wait_cnt,wait_avgms))
        except Exception as e :
            raise
        finally:
            if cursor is not None:
                cursor.close()


    def getTableSpaceStats(self, user, passwd, sid):
        cursor = None
        try:
            cursor = self.connection.cursor()
            cursor.execute("""
            select 
                tablespace_name,
                round(used_space),
                round(max_size-used_space) free_space,
                round(max_size),
                round(used_space*100/max_size,2) percent_used
                from (
                    select m.tablespace_name,
                    m.used_space*t.block_size/1024/1024 used_space,
                    (case when t.bigfile='YES' then power(2,32)*t.block_size/1024/1024
                            else tablespace_size*t.block_size/1024/1024 end) max_size
                from dba_tablespace_usage_metrics m, dba_tablespaces t
            where m.tablespace_name=t.tablespace_name)
            """)
            for tbs in cursor:
                tbs_name = tbs[0]
                used_space_mb = tbs[1]
                free_space_mb = tbs[2]
                max_size_mb = tbs[3]
                percent_used = tbs[4]
                print ("oracle_tablespaces,instance={0},tbs_name={1} used_space_mb={2},free_space_mb={3},percent_used={4},max_size_mb={5}".format(sid, re.sub(' ', '_', tbs_name), used_space_mb,free_space_mb,percent_used,max_size_mb))
        except Exception as e :
            raise
        finally:
            if cursor is not None:
                cursor.close()


    def getMiscMetrics(self, user, passwd, sid):
        query="""select status , count(1) as connectionCount from V$SESSION group by status"""
        cursor = None
        try:
            cursor = self.connection.cursor()
            cursor.execute(query)
            for metric in cursor:
                metric_name = metric[0]
                metric_value = metric[1]
                print("oracle_connectioncount,instance={0},metric_name={1} metric_value={2}".format(sid,metric_name,metric_value))
            
            query="""SELECT 'instance_status'  metric_name,
                        CASE STATUS when 'OPEN' THEN 1
                        ELSE 0 END  metric_value
                    FROM v$instance
                    UNION
                    SELECT 'database_status'  metric_name,
                        CASE DATABASE_STATUS when 'ACTIVE' THEN 1
                        ELSE 0 END  metric_value
                    FROM v$instance
                    """
            cursor = self.connection.cursor()
            cursor.execute(query)
            for metric in cursor:
                metric_name = metric[0]
                metric_value = metric[1]
                print("oracle_status,instance={0},metric_name={1} metric_value={2}".format(sid,metric_name,metric_value))
        except Exception as e :
            raise
        finally:
            if cursor is not None:
                cursor.close()


if __name__ == "__main__":
    try:
        parser = argparse.ArgumentParser()
        parser.add_argument('-u', '--user', help="Pass the username with SELECT_CATALOG_ROLE role granted", required=True)
        parser.add_argument('-p', '--passwd', required=True)
        parser.add_argument('-s', '--sid', help="SID to connect to", required=True)
        args = parser.parse_args()
        stats = None
        try:
            stats = OracleMetrics(args.user, args.passwd, args.sid)
            stats.getWaitClassStats(args.user, args.passwd, args.sid)
            stats.getWaitStats(args.user, args.passwd, args.sid)
            stats.getSysmetrics(args.user, args.passwd, args.sid)
            stats.getTableSpaceStats(args.user, args.passwd, args.sid)
            stats.getMiscMetrics(args.user, args.passwd, args.sid)
        except Exception as e:
            handle_error(e)
        finally:
            if stats is not None:
                stats.connection.close()
    except Exception as e:
        handle_error(e)
