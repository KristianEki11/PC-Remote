# 1. Terminate all running PC Remote processes
$processNames = @("pcremote-server", "PCRemoteDashboard", "test_server", "pcremote-server-debug")
foreach ($p in $processNames) {
    Get-Process -Name $p -ErrorAction SilentlyContinue | Stop-Process -Force -ErrorAction SilentlyContinue
}
taskkill /F /IM pcremote-server.exe 2>$null
taskkill /F /IM PCRemoteDashboard.exe 2>$null
taskkill /F /IM test_server.exe 2>$null

# 2. Cleanup Windows Services (if any)
nssm stop PCRemoteServer 2>$null
nssm remove PCRemoteServer confirm 2>$null

# 3. Remove Firewall Rules
netsh advfirewall firewall delete rule name="PCRemote Server" 2>$null

# 4. Remove Shortcuts and Autostart
$userStartup = [Environment]::GetFolderPath("Startup")
$userDesktop = [Environment]::GetFolderPath("Desktop")
$userPrograms = [Environment]::GetFolderPath("Programs")
$publicDesktop = [Environment]::GetFolderPath("CommonDesktopDirectory")
$publicPrograms = [Environment]::GetFolderPath("CommonPrograms")

Remove-Item "$userStartup\PCRemoteServer.lnk" -Force -ErrorAction SilentlyContinue
Remove-Item "$userDesktop\PCRemote Dashboard.lnk" -Force -ErrorAction SilentlyContinue
Remove-Item "$publicDesktop\PCRemote Dashboard.lnk" -Force -ErrorAction SilentlyContinue
Remove-Item "$userPrograms\PCRemote" -Recurse -Force -ErrorAction SilentlyContinue
Remove-Item "$publicPrograms\PCRemote" -Recurse -Force -ErrorAction SilentlyContinue

# 5. Remove Registry Keys
Remove-ItemProperty -Path "HKLM:\Software\Microsoft\Windows\CurrentVersion\Run" -Name "PCRemoteServer" -ErrorAction SilentlyContinue
Remove-ItemProperty -Path "HKCU:\Software\Microsoft\Windows\CurrentVersion\Run" -Name "PCRemoteServer" -ErrorAction SilentlyContinue

# 6. Remove Installation and Data Folders
$installDir = "C:\Program Files\PCRemote"
$localAppDataDir = "$env:LOCALAPPDATA\PCRemote"

if (Test-Path $installDir) {
    Remove-Item $installDir -Recurse -Force -ErrorAction SilentlyContinue
}
if (Test-Path $localAppDataDir) {
    Remove-Item $localAppDataDir -Recurse -Force -ErrorAction SilentlyContinue
}

Write-Output "CLEAN_UNINSTALL_COMPLETE"
