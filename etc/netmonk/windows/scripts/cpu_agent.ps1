$processor = (Get-CimInstance Win32_Processor)

$VENDORID = $processor.Manufacturer
$MODELNAME = $processor.Name
$FAMILY = $processor.Family
$CPUS = $processor.NumberOfCores #Physical Core
$MHZ = $processor.MaxClockSpeed 
$STEPPING = $processor.Stepping 

Write-Host "server_information,type=cpu vendor_id=""$VENDORID"",model_name=""$MODELNAME"",cpus=$CPUS,MHz=$MHZ,family=$FAMILY,model=""$MODEL"",stepping=""$STEPPING"""
