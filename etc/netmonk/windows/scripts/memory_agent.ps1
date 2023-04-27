$MEMTOTAL = (Get-CimInstance Win32_PhysicalMemory | Measure-Object -Property capacity -Sum).sum

Write-Host "server_information,type=memory mem_total=${MEMTOTAL}"
