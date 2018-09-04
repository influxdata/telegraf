#!/bin/bash

./asinfo -p 3000 -v "namespace/test" | grep -q ";client_proxy_complete=0;"
if [ $? -ne 0 ]
then
	exit 1
fi

./asinfo -p 3010 -v "namespace/test" | grep -q ";client_proxy_complete=0;"
if [ $? -ne 0 ]
then
	exit 1
fi

exit 0