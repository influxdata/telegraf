#!/bin/bash
################################################################################
# Copyright 2013-2016 Aerospike, Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
################################################################################

LOG=$1
if [ ! -f $LOG ]; then
  echo "A log file does not exist at $LOG"
  exit 1
fi

i=0
while [ $i -le 12 ]
do
  sleep 1
  grep -i "there will be cake" $LOG
  if [ $? == 0 ]; then
    exit 0
  else
    i=$(($i + 1))
    echo -n "."
  fi
done
echo "the cake is a lie!"
tail -n 1000 $LOG
exit 2