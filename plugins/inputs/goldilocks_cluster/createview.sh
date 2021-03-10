



if [ $# -eq 6 ]
then
	USER="$1"
	PWD="$2"
	DSN="$3"
	VAR_CLUSTER_NAME="$4"
	VAR_GROUP_NAME="$5"
	VAR_MEMBER_NAME="$6"
else
	echo "Usage : createview.sh <id> <pwd> <dsn> <cluster_name> <group_name> <member_name>"
	exit -1
fi


source ./MonitoringView.sql

GSQL="gsqlnet $USER $PWD --dsn=$DSN"


create_view()
{
$GSQL << EOF > ${DSN}_$1.log
$2
EOF

grep ERR ${DSN}_$1.log > /dev/null


if [ $? -eq 1 ]
then
	printf "OK"
	rm ${DSN}_$1.log
else
	printf "ERR"
fi

}

ignore_list="MONITOR_SHARD_TAB_DISTRIBUTION MONITOR_SHARD_IND_DISTRIBUTION"

for view_name in `grep CREATE MonitoringView.sql | awk '{print $5 }'`
do
	is_ignore="FALSE"
	for ignore in $ignore_list
	do
		if [ $view_name = $ignore ]
	       	then
			is_ignore="TRUE"
			break
		fi

	done


	if [ "$is_ignore" = "TRUE" ] 
	then
		continue
	fi
	printf "> VIEW : %-35s ... [" $view_name
	create_view $view_name "${!view_name}" 
	printf "]\n"
done
