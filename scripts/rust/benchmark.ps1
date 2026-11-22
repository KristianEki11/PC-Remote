param(
    [Parameter(Mandatory=$true)][int[]]$ServerProcessIds,
    [ValidateRange(1,86400)][int]$DurationSeconds = 900,
    [ValidateRange(1,60)][int]$IntervalSeconds = 5,
    [string]$OutputDirectory = (Join-Path $PSScriptRoot 'results'),
    [switch]$IncludeTailscale
)
$ErrorActionPreference = 'Stop'
# Sampling only: never starts/stops applications or sends control commands.
$known = [Collections.Generic.HashSet[int]]::new()
$births = @{}
foreach ($processId in $ServerProcessIds) {
    $rootProcess = Get-CimInstance Win32_Process -Filter "ProcessId = $processId"
    if (-not $rootProcess) { throw "Process $processId is not running." }
    $null = $known.Add($processId)
    $births[$processId] = $rootProcess.CreationDate.ToUniversalTime().Ticks
}
$output = [IO.Path]::GetFullPath($OutputDirectory)
$null = New-Item -ItemType Directory -Force -Path $output
$stamp = Get-Date -Format 'yyyyMMdd-HHmmss'
$csv = Join-Path $output "$stamp-samples.csv"
$started = Get-Date
$previousCpu = @{}
$previousTime = $started
$allRows = [Collections.Generic.List[object]]::new()
while (((Get-Date) - $started).TotalSeconds -lt $DurationSeconds) {
    $now = Get-Date
    $inventory = @(Get-CimInstance Win32_Process | Select-Object ProcessId, ParentProcessId, Name, CreationDate)
    foreach ($item in $inventory) {
        $itemId = [int]$item.ProcessId
        if ($known.Contains($itemId) -and $births.ContainsKey($itemId) -and $births[$itemId] -ne $item.CreationDate.ToUniversalTime().Ticks) {
            $null = $known.Remove($itemId)
        }
    }
    # Recursively include descendants. Processes shorter than the sample interval may be missed.
    do {
        $added = $false
        foreach ($item in $inventory) {
            if ($known.Contains([int]$item.ParentProcessId) -and -not $known.Contains([int]$item.ProcessId)) {
                $null = $known.Add([int]$item.ProcessId)
                $births[[int]$item.ProcessId] = $item.CreationDate.ToUniversalTime().Ticks
                $added = $true
            }
        }
    } while ($added)
    $selected = @($inventory | Where-Object { $known.Contains([int]$_.ProcessId) -or ($IncludeTailscale -and $_.Name -like 'tailscale*') })
    foreach ($item in $selected) {
        $process = Get-Process -Id $item.ProcessId -ErrorAction SilentlyContinue
        if (-not $process) { continue }
        $key = "$($item.ProcessId):$($item.CreationDate.ToUniversalTime().Ticks)"
        $cpu = $process.TotalProcessorTime.TotalSeconds
        $percent = $null
        if ($previousCpu.ContainsKey($key)) {
            $percent = [Math]::Round(100 * [Math]::Max(0, $cpu - $previousCpu[$key]) / [Math]::Max(0.001, ($now - $previousTime).TotalSeconds) / [Environment]::ProcessorCount, 4)
        }
        $previousCpu[$key] = $cpu
        $row = [pscustomobject]@{
            TimestampUtc=$now.ToUniversalTime().ToString('o'); ProcessId=$item.ProcessId; Name=$item.Name
            Group=$(if ($item.Name -like 'tailscale*') { 'tailscale' } else { 'application' })
            PrivateBytes=$process.PrivateMemorySize64; WorkingSetBytes=$process.WorkingSet64
            CpuSeconds=$cpu; MachineCpuPercent=$percent
        }
        $allRows.Add($row)
        $row | Export-Csv -LiteralPath $csv -NoTypeInformation -Append
    }
    $previousTime = $now
    Start-Sleep -Seconds $IntervalSeconds
}
[pscustomobject]@{
    StartedUtc=$started.ToUniversalTime().ToString('o'); DurationSeconds=$DurationSeconds
    IntervalSeconds=$IntervalSeconds; ProcessorCount=[Environment]::ProcessorCount
    RootProcessIds=$ServerProcessIds; IncludeTailscale=[bool]$IncludeTailscale
    SampleRows=$allRows.Count
    Note='CPU is normalized to whole-machine capacity. Sum memory/CPU per timestamp and group; do not sum across time. Short-lived subprocesses may be missed. Pair with ETW/WPR for process-start counts.'
} | ConvertTo-Json | Set-Content -LiteralPath (Join-Path $output "$stamp-metadata.json")
Write-Output $csv
