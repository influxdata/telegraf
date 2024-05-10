#!/usr/bin/env python3
import sys
import time

COUNTER = 0

while True:
    print("counter_python count=" + str(COUNTER))
    sys.stdout.flush()
    COUNTER += 1

    time.sleep(1)
