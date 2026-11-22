param([switch]$Android, [switch]$Installer)
$ErrorActionPreference = 'Stop'
$repo = [IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../..'))
function Invoke-Checked([string]$Command, [string[]]$Arguments) {
    & $Command @Arguments
    if ($LASTEXITCODE -ne 0) { throw "$Command failed with exit code $LASTEXITCODE" }
}
Push-Location (Join-Path $repo 'server-rust')
try {
    Invoke-Checked cargo @('fmt', '--check')
    Invoke-Checked cargo @('clippy', '--locked', '--all-targets', '--', '-D', 'warnings')
    Invoke-Checked cargo @('test', '--locked')
    Invoke-Checked cargo @('build', '--locked', '--release')
} finally { Pop-Location }
if ($Android) {
    Push-Location (Join-Path $repo 'app')
    try {
        Invoke-Checked flutter @('pub', 'get')
        Invoke-Checked flutter @('analyze')
        Invoke-Checked flutter @('test')
        Invoke-Checked flutter @('build', 'apk', '--release')
    } finally { Pop-Location }
}
if ($Installer) {
    $compiler = Get-Command makensis -ErrorAction SilentlyContinue
    $compilerPath = if ($compiler) { $compiler.Source } else { Join-Path ${env:ProgramFiles(x86)} 'NSIS/makensis.exe' }
    if (-not (Test-Path -LiteralPath $compilerPath)) { throw 'NSIS compiler not found. Install NSIS or add makensis to PATH.' }
    Push-Location (Join-Path $repo 'installer/rust')
    try { Invoke-Checked $compilerPath @('PCRemoteRust.nsi') } finally { Pop-Location }
}
Get-FileHash -Algorithm SHA256 -LiteralPath (Join-Path $repo 'server-rust/target/release/pcremote-server.exe')
