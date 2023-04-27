$system = (Get-CimInstance Win32_ComputerSystem)

$MACHINE = ($system).SystemType
$PROCESSOR = ($system).SystemType
$ISVIRTUALIZATION = ($system).HypervisorPresent
$HOSTNAME = ($system).Name

$OS = (Get-CimInstance Win32_OperatingSystem).Caption
$CHASSISTYPE = (Get-CimInstance Win32_SystemEnclosure).ChassisTypes

if ($CHASSISTYPE -eq 12 -or $CHASSISTYPE-eq 21) {} #Ignore Docking Stations

else {
    switch ($CHASSISTYPE) {
        {$_ -in "8", "9", "10", "11", "12", "14", "18", "21","31"} {$chassis = "Laptop"}
        {$_ -eq "32"} {$chassis = "Tablet"}
        {$_ -in "3", "4", "5", "6", "7", "15", "16"} {$chassis = "Desktop"}
        {$_ -eq "23"}{$chassis = "Server"}
        Default {$chassis = "Unknown Type : $CHASSISTYPE" }
    }
}

Write-Host "server_information,type=platform chassis=""$CHASSIS"",hostname=""$HOSTNAME"",os=""$OS"",processor=""$PROCESSOR"",machine=""$MACHINE"",virtualization=""$isVirtualization"""
