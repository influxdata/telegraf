#!/bin/sh

export INFLUX_HOME=$HOME/influx_home/
export PATH=$PATH:$INFLUX_HOME/usr/bin


rec_goldilocks()
{
echo ""
echo "#---------------------------#"
echo "#- CREATE MONITORING VIEW  -#"
echo ""
sh createview.sh $1 $2 $3 $4 $5 $6 

echo ""
echo "#---------------------------#"
echo "#- INIT TELEGRAF METRICS   -#"
echo ""
gsqlnet $1 $2 --dsn=$3 -i InitData.sql > InitData.log

printf "> InitData ... ["
grep ERR InitData.log

if [ $? -eq 1 ]
then
	printf "OK"
	rm InitData.log
else
	printf "ERR"
	cat InitData.log
fi
printf "]\n"
}

rec_influx()
{
influx -database 'telegraf' << EOF
DROP MEASUREMENT goldilocks_session_stat;
DROP MEASUREMENT goldilocks_instance_stat;
DROP MEASUREMENT goldilocks_sql_stat;
DROP MEASUREMENT goldilocks_cluster_dispatcher_stat;
DROP MEASUREMENT goldilocks_tablespace_stat;
DROP MEASUREMENT goldilocks_ager_stat;
DROP MEASUREMENT goldilocks_session_detail;
DROP MEASUREMENT goldilocks_statement_detail;
DROP MEASUREMENT goldilocks_transaction_detail;
DROP MEASUREMENT goldilocks_ssa_stat;
DROP MEASUREMENT goldilocks_shard_table_distibution;
DROP MEASUREMENT goldilocks_shard_index_distibution;
DROP MEASUREMENT goldilocks_tech_shard;
EOF
}

rec_goldilocks sys gliese GOLDILOCKS MATCHING DUMMY DUMMY
rec_goldilocks sys gliese G3         STAND    G3    G3N2
rec_influx

echo ""
