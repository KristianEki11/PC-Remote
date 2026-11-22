[CmdletBinding(SupportsShouldProcess=$true)]
param(
    [string]$Binary = (Join-Path $env:LOCALAPPDATA 'PCRemoteRust/pcremote-server.exe'),
    [ValidateRange(1024,65534)][int]$Port = 8443,
    [switch]$Lan,
    [switch]$Tailscale
)
$ErrorActionPreference = 'Stop'
if (-not $Lan -and -not $Tailscale) { throw 'Choose -Lan, -Tailscale, or both.' }
$binaryPath = (Resolve-Path -LiteralPath $Binary).Path
$identity = [Security.Principal.WindowsIdentity]::GetCurrent()
$principal = New-Object Security.Principal.WindowsPrincipal($identity)
if (-not $principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)) {
    throw 'Run this explicit firewall setup command in an elevated PowerShell window. The installer never changes firewall rules automatically.'
}
function Add-AppRule([string]$Name,[string]$DisplayName,[hashtable]$Scope) {
    if (Get-NetFirewallRule -Name $Name -ErrorAction SilentlyContinue) { throw "Rule $Name already exists. Review it in Windows Firewall rather than creating duplicate access rules." }
    if ($PSCmdlet.ShouldProcess($DisplayName, 'Create narrowly scoped inbound TCP rule')) {
        New-NetFirewallRule -Name $Name -DisplayName $DisplayName -Direction Inbound -Action Allow -Protocol TCP -LocalPort $Port -Program $binaryPath @Scope | Out-Null
    }
}
if ($Lan) {
    Add-AppRule "PCRemoteRust-LAN-$Port" 'PC Remote Rust: private LAN' @{Profile=@('Private','Domain'); RemoteAddress='LocalSubnet'}
}
if ($Tailscale) {
    $adapters = @(Get-NetAdapter | Where-Object { $_.Name -match 'Tailscale' -or $_.InterfaceDescription -match 'Tailscale' })
    if ($adapters.Count -ne 1) { throw 'Could not identify one Tailscale adapter. No broad public-interface rule will be added. Review the adapter name manually.' }
    Add-AppRule "PCRemoteRust-Tailscale-$Port" 'PC Remote Rust: Tailscale interface' @{Profile='Any'; RemoteAddress='100.64.0.0/10'; InterfaceAlias=$adapters[0].Name}
}
