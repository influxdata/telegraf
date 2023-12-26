# Pressure Stall Information (PSI) Input Plugin

A plugin to gather resource pressure metrics from the Linux kernel.
Pressure Stall Information (PSI) is available at
"/proc/pressure/" -- cpu, memory and io.

Examples:
/proc/pressure/cpu
some avg10=1.53 avg60=1.87 avg300=1.73 total=1088168194

/proc/pressure/memory
some avg10=0.00 avg60=0.00 avg300=0.00 total=3463792
full avg10=0.00 avg60=0.00 avg300=0.00 total=1429641

/proc/pressure/io
some avg10=0.00 avg60=0.00 avg300=0.00 total=68568296
full avg10=0.00 avg60=0.00 avg300=0.00 total=54982338
