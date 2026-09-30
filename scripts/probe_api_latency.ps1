<#
.SYNOPSIS
  Measures round-trip latency of the Kanban API and reports which region served
  each request.

.DESCRIPTION
  Every request prints its total wall-clock time plus the `x-vercel-id` header,
  which has the shape `<edge-pop>::<function-region>::<request-id>` (with empty
  segments between the colons). The function region is the second non-empty
  segment: if it is `iad1` (Washington DC) while the database is in
  `ap-southeast-1` (Singapore), every request pays an intercontinental round trip
  per database query, which is what this probe was written to catch.

  The first request to a cold instance is printed separately: a value close to
  the platform function timeout (rather than the p50) is a cold start, not a
  slow query.

.EXAMPLE
  ./probe_api_latency.ps1
  ./probe_api_latency.ps1 -BaseUrl http://localhost:3000 -Path /auth/me,/workspaces -Iterations 20
#>
[CmdletBinding()]
param(
    [string]$BaseUrl = "https://kanban-board-be.vercel.app",
    [string[]]$Path = @("/auth/me"),
    [int]$Iterations = 10,
    [int]$Warmup = 2
)

$ErrorActionPreference = "Stop"
$ProgressPreference = "SilentlyContinue"

function Get-HeaderValue {
    param($Headers, [string]$Name)

    if (-not $Headers) { return $null }

    foreach ($key in $Headers.AllKeys) {
        if ($key -and $key.ToLowerInvariant() -eq $Name.ToLowerInvariant()) {
            return ($Headers[$key] -join ",")
        }
    }

    return $null
}

function Invoke-Probe {
    param([string]$Url)

    $sw = [System.Diagnostics.Stopwatch]::StartNew()
    $status = 0
    $headers = $null

    try {
        # -UseBasicParsing keeps this working on Windows PowerShell 5.1, where
        # the non-basic parser needs Internet Explorer.
        $response = Invoke-WebRequest -Uri $Url -Method Get -UseBasicParsing -TimeoutSec 60
        $status = $response.StatusCode
        $headers = $response.Headers
    }
    catch {
        $webResponse = $_.Exception.Response
        if ($null -eq $webResponse) {
            $sw.Stop()
            return [pscustomobject]@{
                Ms     = [math]::Round($sw.Elapsed.TotalMilliseconds)
                Status = "ERR"
                Region = $_.Exception.Message
            }
        }
        $status = [int]$webResponse.StatusCode
        $headers = $webResponse.Headers
    }

    $sw.Stop()

    # PowerShell 5.1 hands back different header collections for success and
    # error responses, and indexer lookups silently return $null for a name the
    # collection does not expose directly, so match keys case-insensitively.
    $region = "-"
    $vercelID = Get-HeaderValue -Headers $headers -Name "x-vercel-id"
    if ($vercelID) {
        # The header is `hkg1::iad1::q562d-...`: the doubled colons are empty
        # placeholders, so drop empty segments before reading the region
        # (splitting on ":" alone yields ['hkg1', '', 'iad1', '', 'q562d-...']).
        $segments = @($vercelID -split ':' | Where-Object { $_ })
        if ($segments.Count -ge 2) { $region = $segments[1] }
    }

    return [pscustomobject]@{
        Ms     = [math]::Round($sw.Elapsed.TotalMilliseconds)
        Status = $status
        Region = $region
    }
}

Write-Output "Target: $BaseUrl"
Write-Output ""

foreach ($p in $Path) {
    $url = "$($BaseUrl.TrimEnd('/'))$p"

    foreach ($i in 1..$Warmup) {
        $null = Invoke-Probe -Url $url
    }

    $samples = @()
    foreach ($i in 1..$Iterations) {
        $samples += Invoke-Probe -Url $url
    }

    $timings = $samples | Where-Object { $_.Status -ne "ERR" } | Select-Object -ExpandProperty Ms
    $sorted = $timings | Sort-Object

    $p50 = "-"
    $p95 = "-"
    if ($sorted.Count -gt 0) {
        $p50 = $sorted[[math]::Floor($sorted.Count * 0.5)]
        $p95 = $sorted[[math]::Min($sorted.Count - 1, [math]::Floor($sorted.Count * 0.95))]
    }

    $statuses = ($samples | Select-Object -ExpandProperty Status | Sort-Object -Unique) -join ", "
    $regions = ($samples | Select-Object -ExpandProperty Region | Where-Object { $_ -ne "-" } | Sort-Object -Unique) -join ", "
    if (-not $regions) { $regions = "-" }

    Write-Output "$p"
    Write-Output ("  status   : " + $statuses)
    Write-Output ("  region   : " + $regions)
    Write-Output ("  first    : " + $samples[0].Ms + " ms (cold instance if it is the outlier)")
    Write-Output ("  p50      : " + $p50 + " ms")
    Write-Output ("  p95      : " + $p95 + " ms")
    Write-Output ("  min/max  : " + ($sorted | Select-Object -First 1) + " / " + ($sorted | Select-Object -Last 1) + " ms")
    Write-Output ""
}
