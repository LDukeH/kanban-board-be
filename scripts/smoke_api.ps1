<#
.SYNOPSIS
  End-to-end smoke test of the Kanban API with a real cookie session.

.DESCRIPTION
  Registers a throwaway account, then walks the authenticated board-page path
  (workspace -> board -> columns -> tasks -> labels), prints the latency of
  every request, and deletes everything it created.

  It exists because the backend has no Go test suite: this is the quickest way
  to prove a change to the query/service layer still works against a real
  database. Point it at the local dev server (default) or at the deployment:

    ./scripts/smoke_api.ps1
    ./scripts/smoke_api.ps1 -BaseUrl https://kanban-board-be.vercel.app

  What it checks beyond "does it 2xx":
    - GET /auth/me must not expose password_hash.
    - Listing workspaces must report accurate member_count/board_count.
    - GET /columns/{id}/tasks must return the task's labels and assignees
      (the aggregates that used to be per-row subqueries).
    - PATCH /tasks/{id}/position across two columns on the same board must
      succeed (the cross-board check resolves both columns through authz).
    - A resource the caller cannot see must answer 403, not 500. Before the
      authz rewrite an unknown column/task surfaced as an internal error.

  The run ends with a latency table (cold first request included), which is how
  the region/round-trip problem was measured in the first place.
#>
[CmdletBinding()]
param(
    [string]$BaseUrl = "http://localhost:3000",
    [switch]$KeepAccount
)

$ErrorActionPreference = "Stop"
# Invoke-WebRequest writes a progress row per response by default; it drowns the
# report when the script runs from another shell's transcript.
$ProgressPreference = "SilentlyContinue"

$session = New-Object Microsoft.PowerShell.Commands.WebRequestSession
$failures = @()
$timings = @()

function Invoke-Api {
    param(
        [Parameter(Mandatory)][string]$Method,
        [Parameter(Mandatory)][string]$Path,
        $Body = $null
    )

    $uri = "$($BaseUrl.TrimEnd('/'))$Path"
    $params = @{
        Uri             = $uri
        Method          = $Method
        WebSession      = $session
        UseBasicParsing = $true
        TimeoutSec      = 60
    }

    if ($null -ne $Body) {
        $params["Body"] = ($Body | ConvertTo-Json -Depth 10 -Compress)
        $params["ContentType"] = "application/json"
    }

    $sw = [System.Diagnostics.Stopwatch]::StartNew()

    try {
        $response = Invoke-WebRequest @params
        $sw.Stop()
        $decoded = $null
        if ($response.Content -and $response.Content.Length -gt 0) {
            try { $decoded = $response.Content | ConvertFrom-Json } catch { $decoded = $null }
        }
        return [pscustomobject]@{
            Status = [int]$response.StatusCode
            Body   = $decoded
            Ms     = [math]::Round($sw.Elapsed.TotalMilliseconds)
            Text   = $response.Content
        }
    }
    catch {
        $sw.Stop()

        $errorResponse = $_.Exception.Response
        if ($null -eq $errorResponse) {
            return [pscustomobject]@{
                Status = 0
                Body   = $null
                Ms     = [math]::Round($sw.Elapsed.TotalMilliseconds)
                Text   = $_.Exception.Message
            }
        }

        # Windows PowerShell 5.1 does not expose the error body directly.
        $reader = New-Object System.IO.StreamReader($errorResponse.GetResponseStream())
        $text = $reader.ReadToEnd()

        return [pscustomobject]@{
            Status = [int]$errorResponse.StatusCode
            Body   = $null
            Ms     = [math]::Round($sw.Elapsed.TotalMilliseconds)
            Text   = $text
        }
    }
}
function Step {
    param(
        [Parameter(Mandatory)][string]$Label,
        [Parameter(Mandatory)][string]$Method,
        [Parameter(Mandatory)][string]$Path,
        $Body = $null,
        [int[]]$Expected = @(200)
    )

    $result = Invoke-Api -Method $Method -Path $Path -Body $Body
    $script:timings += [pscustomobject]@{ Label = $Label; Ms = $result.Ms }

    $ok = $Expected -contains $result.Status
    if (-not $ok) {
        $script:failures += "$Label ($Method $Path) -> $($result.Status), expected $($Expected -join '/'). Body: $($result.Text)"
    }

    $marker = if ($ok) { "ok  " } else { "FAIL" }
    # Write-Host, not Write-Output: the caller assigns Step's return value, and
    # anything written to the output stream would be captured alongside it.
    Write-Host ("  [{0}] {1,-46} {2,-4} {3,6} ms" -f $marker, $Label, $result.Status, $result.Ms)

    return $result
}

function Assert {
    param([Parameter(Mandatory)][bool]$Condition, [Parameter(Mandatory)][string]$Message)

    if ($Condition) {
        Write-Host ("  [ok ] {0}" -f $Message)
    }
    else {
        $script:failures += $Message
        Write-Host ("  [FAIL] {0}" -f $Message)
    }
}

$email = "smoke-$([guid]::NewGuid().ToString('N').Substring(0, 12))@example.com"
$password = "Sm0ke-$([guid]::NewGuid().ToString('N').Substring(0, 12))"

Write-Output "Smoke test against $BaseUrl"
Write-Output "Account: $email"
Write-Output ""

Write-Output "session"
$null = Step -Label "POST /auth/register" -Method Post -Path "/auth/register" -Body @{ email = $email; password = $password } -Expected 204
$null = Step -Label "POST /auth/login" -Method Post -Path "/auth/login" -Body @{ email = $email; password = $password } -Expected 204

$me = Step -Label "GET /auth/me" -Method Get -Path "/auth/me" -Expected 200
Assert -Condition ($null -ne $me.Body.id) -Message "GET /auth/me returns an id"
Assert -Condition ($null -eq $me.Body.PSObject.Properties["password_hash"]) -Message "GET /auth/me does not expose password_hash"

Write-Output ""
Write-Output "workspace"
$workspace = Step -Label "POST /workspaces/" -Method Post -Path "/workspaces/" -Body @{ name = "Smoke workspace"; description = "created by smoke_api.ps1" } -Expected 201
$workspaceID = $workspace.Body.id

$list = Step -Label "GET /workspaces/" -Method Get -Path "/workspaces/" -Expected 200
$mine = $list.Body | Where-Object { $_.id -eq $workspaceID }
Assert -Condition ($null -ne $mine) -Message "GET /workspaces/ lists the new workspace"
Assert -Condition ($mine.member_count -eq 1) -Message "member_count is 1 (creator membership)"
Assert -Condition ($mine.board_count -eq 0) -Message "board_count is 0 before any board exists"
Write-Output ""
Write-Output "board and columns"
$board = Step -Label "POST /workspaces/{id}/boards/" -Method Post -Path "/workspaces/$workspaceID/boards/" -Body @{ name = "Smoke board" } -Expected 201
$boardID = $board.Body.id

$boards = Step -Label "GET /workspaces/{id}/boards/" -Method Get -Path "/workspaces/$workspaceID/boards/" -Expected 200
Assert -Condition (@($boards.Body | Where-Object { $_.id -eq $boardID }).Count -eq 1) -Message "board appears in the workspace list"

$columnA = Step -Label "POST /boards/{id}/columns/ (A)" -Method Post -Path "/boards/$boardID/columns/" -Body @{ name = "Todo" } -Expected 201
$columnB = Step -Label "POST /boards/{id}/columns/ (B)" -Method Post -Path "/boards/$boardID/columns/" -Body @{ name = "Done" } -Expected 201
$columnAID = $columnA.Body.id
$columnBID = $columnB.Body.id

$columns = Step -Label "GET /boards/{id}/columns/" -Method Get -Path "/boards/$boardID/columns/" -Expected 200
Assert -Condition (@($columns.Body).Count -eq 2) -Message "both columns are listed"

Write-Output ""
Write-Output "task and label"
$task = Step -Label "POST /columns/{id}/tasks/" -Method Post -Path "/columns/$columnAID/tasks/" -Body @{
    title       = "Smoke task"
    description = @{ type = "doc"; content = @(@{ type = "paragraph"; content = @(@{ type = "text"; text = "hello" }) }) }
} -Expected 201
$taskID = $task.Body.id

$label = Step -Label "POST /boards/{id}/labels/" -Method Post -Path "/boards/$boardID/labels/" -Body @{ name = "smoke"; color = "#4f46e5" } -Expected 201
$labelID = $label.Body.id

$null = Step -Label "POST /tasks/{id}/labels/" -Method Post -Path "/tasks/$taskID/labels/" -Body @{ label_id = $labelID } -Expected 201

$taskLabels = Step -Label "GET /tasks/{id}/labels/" -Method Get -Path "/tasks/$taskID/labels/" -Expected 200
Assert -Condition (@($taskLabels.Body | Where-Object { $_.id -eq $labelID }).Count -eq 1) -Message "the label is attached to the task"

$columnTasks = Step -Label "GET /columns/{id}/tasks/" -Method Get -Path "/columns/$columnAID/tasks/" -Expected 200
# @(): PowerShell 5.1 turns a one-element JSON array into a bare object, and a
# bare object has no Count, so an unwrapped count comparison silently fails.
$listed = @($columnTasks.Body | Where-Object { $_.id -eq $taskID })
Assert -Condition ($listed.Count -eq 1) -Message "the task is listed in its column"
Assert -Condition ($listed[0].labels.Count -eq 1) -Message "the task carries its aggregated labels"
Assert -Condition ($listed[0].assignees.Count -eq 0) -Message "the task carries an empty (not null) assignee list"

Write-Output ""
Write-Output "mutation"
$null = Step -Label "PATCH /tasks/{id}" -Method Patch -Path "/tasks/$taskID" -Body @{ title = "Smoke task renamed" } -Expected 204
$null = Step -Label "PATCH /tasks/{id}/position (other column)" -Method Patch -Path "/tasks/$taskID/position" -Body @{ column_id = $columnBID; position = 0 } -Expected 204

$moved = Step -Label "GET /columns/{id}/tasks/ (B)" -Method Get -Path "/columns/$columnBID/tasks/" -Expected 200
Assert -Condition (@($moved.Body | Where-Object { $_.id -eq $taskID }).Count -eq 1) -Message "the task moved to the second column"
Write-Output ""
Write-Output "authorization"
$random = [guid]::NewGuid().ToString()
$unknownColumn = Step -Label "PATCH /columns/{unknown}" -Method Patch -Path "/columns/$random" -Body @{ name = "nope" } -Expected 403, 404
Assert -Condition ($unknownColumn.Status -eq 403) -Message "an unknown column answers 403 (it used to be a 500)"
$unknownTask = Step -Label "PATCH /tasks/{unknown}" -Method Patch -Path "/tasks/$random" -Body @{ title = "nope" } -Expected 403, 404
Assert -Condition ($unknownTask.Status -eq 403) -Message "an unknown task answers 403 (it used to be a 500)"
$unknownBoard = Step -Label "GET /boards/{unknown}" -Method Get -Path "/boards/$random" -Expected 404
Assert -Condition ($unknownBoard.Status -eq 404) -Message "an unknown board answers 404"

# Informational: the live frontend suite expects 400 here because the frontend
# trims and rejects a blank name; the API itself only rejects an empty string.
$blankName = Invoke-Api -Method Patch -Path "/workspaces/$workspaceID" -Body @{ name = "   "; description = "Platform roadmap" }
Write-Host ("  [info] PATCH /workspaces/<id> with a blank name -> {0} ({1} ms)" -f $blankName.Status, $blankName.Ms)

Write-Output ""
Write-Output "cleanup"
$null = Step -Label "DELETE /tasks/{id}" -Method Delete -Path "/tasks/$taskID" -Expected 204
$null = Step -Label "DELETE /boards/{id}" -Method Delete -Path "/boards/$boardID" -Expected 204
$null = Step -Label "DELETE /workspaces/{id}" -Method Delete -Path "/workspaces/$workspaceID" -Expected 204
if (-not $KeepAccount) {
    $null = Step -Label "DELETE /auth/me" -Method Delete -Path "/auth/me" -Expected 204
}

Write-Output ""
Write-Output "latency"
$sorted = $timings | Sort-Object Ms
$count = $sorted.Count
if ($count -gt 0) {
    Write-Output ("  first (cold) : {0} ms" -f $timings[0].Ms)
    Write-Output ("  p50          : {0} ms" -f $sorted[[math]::Floor($count * 0.5)].Ms)
    Write-Output ("  p95          : {0} ms" -f $sorted[[math]::Min($count - 1, [math]::Floor($count * 0.95))].Ms)
    Write-Output ("  slowest      : {0} ({1} ms)" -f $sorted[-1].Label, $sorted[-1].Ms)
}
Write-Output ("  requests     : {0}" -f $count)

Write-Output ""
if ($failures.Count -gt 0) {
    Write-Output ("FAILED ({0})" -f $failures.Count)
    foreach ($failure in $failures) { Write-Output ("  - " + $failure) }
    exit 1
}

Write-Output "PASSED"



