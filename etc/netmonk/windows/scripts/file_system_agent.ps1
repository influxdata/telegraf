$logicalDisks = Get-CimInstance -Class Win32_LogicalDisk
$volumeName = "unnamed volume"

if ($logicalDisks.Count -gt 1) {
    foreach ($disk in $logicalDisks) {
	if ($disk.VolumeName -ne "") {
	    $volumeName = $disk.VolumeName
	}
        Write-Host "server_information,type=filesystem key=$volumeName,size=$($disk.Size),mounted_on=$($disk.DeviceID)"
    }
} else {
    $disk = $logicalDisks[0]
    if ($disk.VolumeName -ne "") {
         $volumeName = $disk.VolumeName
    }
    Write-Host "server_information,type=filesystem key=$volumeName,size=$($disk.Size),mounted_on=$($disk.DeviceID)"
}
