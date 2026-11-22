param(
    [string]$Binary = (Join-Path $PSScriptRoot '../../server-rust/target/debug/pcremote-server.exe'),
    [ValidateRange(1024,65534)][int]$Port = 18443
)
$ErrorActionPreference = 'Stop'
$binaryPath = (Resolve-Path -LiteralPath $Binary).Path
if (Get-NetTCPConnection -State Listen -LocalPort $Port,($Port+1) -ErrorAction SilentlyContinue) { throw 'Smoke-test ports are already in use.' }
$testDir = Join-Path $PSScriptRoot ('results/smoke-' + [guid]::NewGuid().ToString('N'))
$null = New-Item -ItemType Directory -Force -Path $testDir
# Only this test's explicit loopback HTTPS port bypasses TLS verification.
# Never use this helper in the production client or with remote endpoints.
Add-Type -TypeDefinition @'
using System;
using System.IO;
using System.Net;
using System.Text;
public static class PCRemoteSmokeHttp {
    public static string[] Send(string url, string method, string body, string headerName, string headerValue, int testPort) {
        Uri uri = new Uri(url);
        if (uri.Host != "127.0.0.1" || (uri.Port != testPort && uri.Port != testPort + 1)) throw new Exception("Only isolated loopback test endpoints are allowed");
        var request = (HttpWebRequest)WebRequest.Create(uri);
        request.Method = method;
        request.Timeout = 10000;
        request.Proxy = null;
        request.AllowAutoRedirect = false;
        request.ServerCertificateValidationCallback = (sender, cert, chain, errors) => uri.Scheme == "https" && uri.Port == testPort;
        if (!String.IsNullOrEmpty(headerName)) request.Headers[headerName] = headerValue;
        if (method != "GET" && body != null) {
            request.ContentType = "application/json";
            byte[] bytes = Encoding.UTF8.GetBytes(body);
            request.ContentLength = bytes.Length;
            using (var stream = request.GetRequestStream()) stream.Write(bytes, 0, bytes.Length);
        }
        HttpWebResponse response;
        try { response = (HttpWebResponse)request.GetResponse(); }
        catch (WebException error) { if (error.Response == null) throw; response = (HttpWebResponse)error.Response; }
        using (response) using (var reader = new StreamReader(response.GetResponseStream())) {
            return new string[] { ((int)response.StatusCode).ToString(), reader.ReadToEnd() };
        }
    }
}
'@
[Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12
function Request([string]$Url, [string]$Method='GET', $Body=$null, [string]$HeaderName='', [string]$HeaderValue='') {
    $json = if ($null -eq $Body) { $null } else { ConvertTo-Json -InputObject $Body -Compress }
    $result = [PCRemoteSmokeHttp]::Send($Url,$Method,$json,$HeaderName,$HeaderValue,$Port)
    [pscustomobject]@{Status=[int]$result[0]; Body=$result[1]}
}
function Assert-Status($Response,[int]$Expected,[string]$Label) {
    if ($Response.Status -ne $Expected) { throw "$Label returned $($Response.Status), expected $Expected" }
}
$start = New-Object Diagnostics.ProcessStartInfo
$start.FileName = $binaryPath
$start.Arguments = '--background'
$start.UseShellExecute = $false
$start.CreateNoWindow = $true
$start.WindowStyle = [Diagnostics.ProcessWindowStyle]::Hidden
$start.EnvironmentVariables['PCREMOTE_PORT'] = "$Port"
$start.EnvironmentVariables['PCREMOTE_DATA_DIR'] = [IO.Path]::GetFullPath($testDir)
$process = [Diagnostics.Process]::Start($start)
$api = "https://127.0.0.1:$Port"
$admin = "http://127.0.0.1:$($Port+1)"
try {
    $ready = $false
    for ($attempt=0; $attempt -lt 30; $attempt++) {
        if ($process.HasExited) { throw "Server exited during startup ($($process.ExitCode))" }
        try { $health=Request "$api/health"; if ($health.Status -eq 200) { $ready=$true; break } } catch { }
        Start-Sleep -Milliseconds 500
    }
    if (-not $ready) { throw 'Server failed to become ready' }
    $page=Request "$admin/"
    Assert-Status $page 200 'Local dashboard'
    $match=[regex]::Match($page.Body,"const token='([a-f0-9]+)'")
    if (-not $match.Success) { throw 'Could not extract local setup nonce' }
    $setupToken=$match.Groups[1].Value
    Assert-Status (Request "$admin/setup" 'POST' @{action='status'}) 403 'Unauthenticated setup'
    Assert-Status (Request "$admin/setup" 'POST' @{action='pin';pin='19283746'} 'X-Setup-Token' $setupToken) 200 'Set test PIN'
    $login=Request "$api/auth/login" 'POST' @{pin='19283746'}
    Assert-Status $login 200 'PIN login'
    $credentials=$login.Body | ConvertFrom-Json
    if (-not $credentials.token -or -not $credentials.device_id) { throw 'Missing credentials' }
    $bearer='Bearer ' + $credentials.token
    Assert-Status (Request "$api/auth/verify" 'GET' $null 'Authorization' $bearer) 200 'Verify token'
    $connRes = Request "$api/connection" 'GET' $null 'Authorization' $bearer
    Assert-Status $connRes 200 'Connection metadata'
    $conn = $connRes.Body | ConvertFrom-Json
    foreach ($field in @('hosts', 'tailscale_state', 'tailscale_urls', 'server_id', 'fingerprint')) {
        if ($null -eq $conn.PSObject.Properties[$field]) { throw "Connection metadata missing field: $field" }
    }
    if ($null -eq $conn.hosts) { throw 'Connection metadata hosts is null' }
    if ($null -eq $conn.tailscale_urls) { throw 'Connection metadata tailscale_urls is null' }
    if ([string]::IsNullOrWhiteSpace($conn.tailscale_state)) { throw 'Connection metadata tailscale_state is empty' }
    if ([string]::IsNullOrWhiteSpace($conn.server_id)) { throw 'Connection metadata server_id is empty' }
    if ([string]::IsNullOrWhiteSpace($conn.fingerprint)) { throw 'Connection metadata fingerprint is empty' }
    Assert-Status (Request "$api/auth/verify" 'GET' $null 'Authorization' 'Bearer wrong') 401 'Unknown token'
    # A read-only native call; hardware may legitimately have no audio endpoint.
    $audio=Request "$api/audio/status" 'GET' $null 'Authorization' $bearer
    if ($audio.Status -ne 200 -and $audio.Status -ne 400) { throw "Audio read returned unexpected status $($audio.Status)" }
    Assert-Status (Request "$admin/setup" 'POST' @{action='revoke';device_id=$credentials.device_id} 'X-Setup-Token' $setupToken) 200 'Revoke device'
    Assert-Status (Request "$api/auth/verify" 'GET' $null 'Authorization' $bearer) 401 'Revoked token'
    Write-Output "PASS: isolated HTTPS setup, login, verification, endpoint metadata, unknown-token rejection and revocation. Audio status: $($audio.Status)."
} finally {
    if (-not $process.HasExited) {
        if ($setupToken) { try { $null=Request "$admin/setup" 'POST' @{action='stop'} 'X-Setup-Token' $setupToken } catch { } }
        if (-not $process.WaitForExit(5000)) { $process.Kill(); $process.WaitForExit() }
    }
    $process.Dispose()
}
