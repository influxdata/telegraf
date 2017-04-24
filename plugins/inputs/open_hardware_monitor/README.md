# OpenHardwareMonitor Input Plugin

This input plugin will gather sensors data provide by [Open hardware Monitor](http://openhardwaremonitor.org) application via Windows Management Instrumentation interface (WMI) 

### Configuration:

```
# # Get sensors data from Open Hardware Monitor via WMI
# [[inputs.open_hardware_monitor]]
	## Sensors to query ( if not given then all is queried )
	SensorsType = ["Temperature", "Fan", "Voltage"] # optional
	
	## Which hardware should be available
	Parent = ["intelcpu_0"]  # optional
```

### Measurements & Fields:

- All sensors provided by OpenHardwareMonitor or specify subset defined in the SensorsType configuration. 

### Tags:

- All measurements have the following tags:
	
	- name
	- parent

### Example Output:

```
* Plugin: open_hardware_monitor, Collection 1
ohm,host=Test-PC,name=Temperature_#2,parent=lpc_nct6779d Temperature=34 1469698553000000000
ohm,host=Test-PC,name=VTT,parent=lpc_nct6779d Voltage=1.056 1469698553000000000
ohm,host=Test-PC,name=Voltage_#14,parent=lpc_nct6779d Voltage=1 1469698553000000000
ohm,host=Test-PC,name=3VCC,parent=lpc_nct6779d Voltage=3.4240003 1469698553000000000
ohm,host=Test-PC,name=CPU_Core_#3,parent=intelcpu_0 Temperature=41 1469698553000000000
ohm,host=Test-PC,name=Fan_Control_#4,parent=lpc_nct6779d Control=100 1469698553000000000
ohm,host=Test-PC,name=GPU_Shader,parent=nvidiagpu_0 Clock=270 1469698553000000000
ohm,host=Test-PC,name=Bus_Speed,parent=intelcpu_0 Clock=99.99993 1469698553000000000
ohm,host=Test-PC,name=CPU_Core_#2,parent=intelcpu_0 Clock=1599.9989 1469698553000000000
ohm,host=Test-PC,name=CPU_Core_#2,parent=intelcpu_0 Load=0 1469698553000000000
ohm,host=Test-PC,name=GPU_Fan,parent=nvidiagpu_0 Control=33 1469698553000000000
ohm,host=Test-PC,name=Available_Memory,parent=ram Data=13.124931 1469698553000000000
ohm,host=Test-PC,name=3VSB,parent=lpc_nct6779d Voltage=3.4080002 1469698553000000000
ohm,host=Test-PC,name=Voltage_#12,parent=lpc_nct6779d Voltage=0.24800001 1469698553000000000
ohm,host=Test-PC,name=Voltage_#2,parent=lpc_nct6779d Voltage=1 1469698553000000000
ohm,host=Test-PC,name=Fan_Control_#1,parent=lpc_nct6779d Control=74.117645 1469698553000000000
ohm,host=Test-PC,name=CPU_Core,parent=lpc_nct6779d Temperature=46 1469698553000000000
ohm,host=Test-PC,name=Voltage_#6,parent=lpc_nct6779d Voltage=2.0400002 1469698553000000000
ohm,host=Test-PC,name=GPU_Core,parent=nvidiagpu_0 Clock=135 1469698553000000000
ohm,host=Test-PC,name=CPU_Cores,parent=intelcpu_0 Power=4.77574 1469698553000000000
ohm,host=Test-PC,name=CPU_Package,parent=intelcpu_0 Temperature=45 1469698553000000000
ohm,host=Test-PC,name=GPU_Memory_Controller,parent=nvidiagpu_0 Load=5 1469698553000000000
ohm,host=Test-PC,name=Memory,parent=ram Load=45.18159 1469698553000000000
ohm,host=Test-PC,name=CPU_Core_#4,parent=intelcpu_0 Clock=2999.998 1469698553000000000
ohm,host=Test-PC,name=CPU_Core_#4,parent=intelcpu_0 Load=3.125 1469698553000000000
ohm,host=Test-PC,name=GPU_Core,parent=nvidiagpu_0 Temperature=37 1469698553000000000
ohm,host=Test-PC,name=Temperature_#6,parent=lpc_nct6779d Temperature=-7 1469698553000000000
ohm,host=Test-PC,name=Fan_#4,parent=lpc_nct6779d Fan=672 1469698553000000000
ohm,host=Test-PC,name=CPU_Core_#1,parent=intelcpu_0 Temperature=44 1469698553000000000
ohm,host=Test-PC,name=Temperature_#4,parent=lpc_nct6779d Temperature=-25 1469698553000000000
ohm,host=Test-PC,name=Temperature_#5,parent=lpc_nct6779d Temperature=90 1469698553000000000
ohm,host=Test-PC,name=Fan_#2,parent=lpc_nct6779d Fan=1355 1469698553000000000
ohm,host=Test-PC,name=Voltage_#11,parent=lpc_nct6779d Voltage=1.8240001 1469698553000000000
ohm,host=Test-PC,name=Temperature_#3,parent=lpc_nct6779d Temperature=32 1469698553000000000
ohm,host=Test-PC,name=GPU_Video_Engine,parent=nvidiagpu_0 Load=0 1469698553000000000
ohm,host=Test-PC,name=VBAT,parent=lpc_nct6779d Voltage=3.3600001 1469698553000000000
ohm,host=Test-PC,name=Used_Space,parent=hdd_1 Load=36.86175 1469698553000000000
ohm,host=Test-PC,name=Voltage_#13,parent=lpc_nct6779d Voltage=1.016 1469698553000000000
ohm,host=Test-PC,name=AVCC,parent=lpc_nct6779d Voltage=3.4240003 1469698553000000000
ohm,host=Test-PC,name=CPU_Core_#2,parent=intelcpu_0 Temperature=41 1469698553000000000
ohm,host=Test-PC,name=Fan_Control_#3,parent=lpc_nct6779d Control=100 1469698553000000000
ohm,host=Test-PC,name=GPU_Memory,parent=nvidiagpu_0 Clock=405.00003 1469698553000000000
ohm,host=Test-PC,name=CPU_Graphics,parent=intelcpu_0 Power=0 1469698553000000000
ohm,host=Test-PC,name=CPU_Core_#1,parent=intelcpu_0 Clock=1599.9989 1469698553000000000
ohm,host=Test-PC,name=CPU_Core_#1,parent=intelcpu_0 Load=3.125 1469698553000000000
ohm,host=Test-PC,name=CPU_Total,parent=intelcpu_0 Load=2.734375 1469698553000000000
ohm,host=Test-PC,name=Used_Memory,parent=ram Data=10.817631 1469698553000000000
ohm,host=Test-PC,name=Voltage_#7,parent=lpc_nct6779d Voltage=1.5680001 1469698553000000000
ohm,host=Test-PC,name=Used_Space,parent=hdd_0 Load=48.958797 1469698553000000000
ohm,host=Test-PC,name=Temperature_#1,parent=lpc_nct6779d Temperature=39 1469698553000000000
ohm,host=Test-PC,name=CPU_VCore,parent=lpc_nct6779d Voltage=0.87200004 1469698553000000000
ohm,host=Test-PC,name=Voltage_#15,parent=lpc_nct6779d Voltage=0.216 1469698553000000000
ohm,host=Test-PC,name=Voltage_#5,parent=lpc_nct6779d Voltage=1.016 1469698553000000000
ohm,host=Test-PC,name=GPU,parent=nvidiagpu_0 Fan=1480 1469698553000000000
ohm,host=Test-PC,name=CPU_Package,parent=intelcpu_0 Power=10.8843355 1469698553000000000
ohm,host=Test-PC,name=CPU_Core_#4,parent=intelcpu_0 Temperature=32 1469698553000000000
ohm,host=Test-PC,name=Fan_Control_#5,parent=lpc_nct6779d Control=100 1469698553000000000
ohm,host=Test-PC,name=GPU_Core,parent=nvidiagpu_0 Load=0 1469698553000000000
ohm,host=Test-PC,name=CPU_Core_#3,parent=intelcpu_0 Clock=2999.998 1469698553000000000
ohm,host=Test-PC,name=CPU_Core_#3,parent=intelcpu_0 Load=4.6875 1469698553000000000
ohm,host=Test-PC,name=GPU_Memory,parent=nvidiagpu_0 Load=8.956909 1469698553000000000
ohm,host=Test-PC,name=Fan_#1,parent=lpc_nct6779d Fan=1078 1469698553000000000
ohm,host=Test-PC,name=Fan_Control_#2,parent=lpc_nct6779d Control=45.882355 1469698553000000000
ohm,host=Test-PC,name=Fan_#3,parent=lpc_nct6779d Fan=1000 1469698553000000000
