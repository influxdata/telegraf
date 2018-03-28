#!/bin/bash

TELEGRAF_VERSION=1.5.3-azmon
CONTAINER_URI=https://masdiagstore.blob.core.windows.net/share/

make

tar -cvfz telegraf-${TELEGRAF_VERSION}-l_amd64.tar.gz ./telegraf

azcopy \
    --source ./telegraf-${TELEGRAF_VERSION}-l_amd64.tar.gz \
    --destination $CONTAINER_URL/telegraf-${TELEGRAF_VERSION}-l_amd64.tar.gz \
    --dest-key $AZURE_STORAGE_KEY
