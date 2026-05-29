param(
    [string]$BaseUrl = "http://127.0.0.1:8080",
    [int]$TotalRequests = 3000,
    [int]$TotalConcurrency = 60,
    [int]$ReadPercent = 70,
    [int]$LoginPercent = 20,
    [int]$WritePercent = 10,
    [int]$SeedPosts = 20,
    [switch]$SkipSetup,
    [switch]$KeepRunning
)

$ErrorActionPreference = "Stop"

function Write-Step {
    param([string]$Message)
    Write-Host "[stress-test-realistic] $Message"
}

function Test-EndpointHealthy {
    param([string]$Url)

    try {
        $response = Invoke-WebRequest -Uri $Url -UseBasicParsing -TimeoutSec 2
        return $response.StatusCode -eq 200
    } catch {
        return $false
    }
}

function Wait-ForHealth {
    param(
        [string]$Url,
        [int]$TimeoutSeconds = 30
    )

    $startedAt = Get-Date
    while (((Get-Date) - $startedAt).TotalSeconds -lt $TimeoutSeconds) {
        if (Test-EndpointHealthy -Url $Url) {
            return $true
        }

        Start-Sleep -Seconds 1
    }

    return $false
}

function Ensure-HeyInstalled {
    $gopath = (& go env GOPATH).Trim()
    $heyPath = Join-Path $gopath "bin\hey.exe"
    if (Test-Path $heyPath) {
        return $heyPath
    }

    Write-Step "Instalando hey"
    & go install github.com/rakyll/hey@latest
    if (-not (Test-Path $heyPath)) {
        throw "Nao foi possivel localizar hey em $heyPath"
    }

    return $heyPath
}

function New-TempBodyFile {
    param([string]$Content)

    $tempFile = Join-Path ([System.IO.Path]::GetTempPath()) ("go-tweets-" + [guid]::NewGuid().ToString() + ".json")
    $utf8WithoutBom = New-Object System.Text.UTF8Encoding($false)
    [System.IO.File]::WriteAllText($tempFile, $Content, $utf8WithoutBom)
    return $tempFile
}

function Get-WeightedShare {
    param(
        [int]$Total,
        [int]$Percent,
        [int]$Minimum = 1
    )

    if ($Total -le 0 -or $Percent -le 0) {
        return 0
    }

    $value = [Math]::Floor(($Total * $Percent) / 100)
    if ($value -lt $Minimum) {
        return $Minimum
    }

    return [int]$value
}

function Start-HeyScenario {
    param(
        [string]$HeyPath,
        [string[]]$Arguments,
        [string]$ScenarioName
    )

    $stdoutFile = Join-Path ([System.IO.Path]::GetTempPath()) ("go-tweets-" + $ScenarioName + "-" + [guid]::NewGuid().ToString() + ".out.txt")
    $stderrFile = Join-Path ([System.IO.Path]::GetTempPath()) ("go-tweets-" + $ScenarioName + "-" + [guid]::NewGuid().ToString() + ".err.txt")

    $process = Start-Process -FilePath $HeyPath `
        -ArgumentList $Arguments `
        -WorkingDirectory $projectRoot `
        -PassThru `
        -WindowStyle Hidden `
        -RedirectStandardOutput $stdoutFile `
        -RedirectStandardError $stderrFile

    return @{
        Process = $process
        Stdout  = $stdoutFile
        Stderr  = $stderrFile
        Name    = $ScenarioName
    }
}

if (($ReadPercent + $LoginPercent + $WritePercent) -ne 100) {
    throw "A soma de ReadPercent, LoginPercent e WritePercent deve ser 100."
}

$projectRoot = Split-Path -Parent $PSScriptRoot
$healthUrl = "$BaseUrl/check-health"
$metricsUrl = "$BaseUrl/metrics"
$apiProcess = $null
$apiStartedByScript = $false
$dbStartedByScript = $false
$tempFiles = @()
$scenarioRuns = @()

try {
    Set-Location $projectRoot

    if (-not $SkipSetup) {
        Write-Step "Subindo banco com docker compose"
        & docker compose up -d | Out-Host
        $dbStartedByScript = $true

        Write-Step "Aplicando migrations"
        & npx dbmate up | Out-Host
    }

    if (-not (Test-EndpointHealthy -Url $healthUrl)) {
        Write-Step "Subindo API"
        $apiProcess = Start-Process -FilePath "go" -ArgumentList @("run", "./cmd") -WorkingDirectory $projectRoot -PassThru -WindowStyle Hidden
        $apiStartedByScript = $true

        if (-not (Wait-ForHealth -Url $healthUrl -TimeoutSeconds 40)) {
            throw "A API nao ficou saudavel em tempo habil."
        }
    }

    $timestamp = [DateTimeOffset]::UtcNow.ToUnixTimeSeconds()
    $email = "stress-realistic-$timestamp@example.com"
    $username = "stressrealistic$timestamp"
    $password = "secret123"

    Write-Step "Criando usuario base"
    $registerBody = @{
        email            = $email
        username         = $username
        password         = $password
        password_confirm = $password
    } | ConvertTo-Json

    Invoke-RestMethod -Method Post -Uri "$BaseUrl/auth/register" -ContentType "application/json" -Body $registerBody | Out-Null

    Write-Step "Autenticando usuario base"
    $loginBody = @{
        email    = $email
        password = $password
    } | ConvertTo-Json

    $loginResponse = Invoke-RestMethod -Method Post -Uri "$BaseUrl/auth/login" -ContentType "application/json" -Body $loginBody
    $token = $loginResponse.token

    Write-Step "Gerando $SeedPosts posts de seed"
    1..$SeedPosts | ForEach-Object {
        $postBody = @{
            title   = "Realistic Seed Post $_"
            content = "Payload generated for realistic stress test $_"
        } | ConvertTo-Json

        Invoke-RestMethod -Method Post -Uri "$BaseUrl/tweets/" -Headers @{ Authorization = $token } -ContentType "application/json" -Body $postBody | Out-Null
    }

    $heyPath = Ensure-HeyInstalled

    $loginBodyFile = New-TempBodyFile -Content $loginBody
    $tempFiles += $loginBodyFile

    $writeBody = @{
        title   = "Realistic stress title"
        content = "Realistic stress content payload"
    } | ConvertTo-Json
    $writeBodyFile = New-TempBodyFile -Content $writeBody
    $tempFiles += $writeBodyFile

    $readRequests = Get-WeightedShare -Total $TotalRequests -Percent $ReadPercent
    $loginRequests = Get-WeightedShare -Total $TotalRequests -Percent $LoginPercent
    $writeRequests = [Math]::Max(1, $TotalRequests - $readRequests - $loginRequests)

    $readConcurrency = Get-WeightedShare -Total $TotalConcurrency -Percent $ReadPercent
    $loginConcurrency = Get-WeightedShare -Total $TotalConcurrency -Percent $LoginPercent
    $writeConcurrency = [Math]::Max(1, $TotalConcurrency - $readConcurrency - $loginConcurrency)

    Write-Step "Distribuicao: read=$readRequests/$readConcurrency login=$loginRequests/$loginConcurrency write=$writeRequests/$writeConcurrency"

    $scenarioRuns += Start-HeyScenario -HeyPath $heyPath -ScenarioName "read" -Arguments @(
        "-n", "$readRequests",
        "-c", "$readConcurrency",
        "$BaseUrl/tweets/?page=1&limit=10"
    )

    $scenarioRuns += Start-HeyScenario -HeyPath $heyPath -ScenarioName "login" -Arguments @(
        "-n", "$loginRequests",
        "-c", "$loginConcurrency",
        "-m", "POST",
        "-T", "application/json",
        "-D", $loginBodyFile,
        "$BaseUrl/auth/login"
    )

    $scenarioRuns += Start-HeyScenario -HeyPath $heyPath -ScenarioName "write" -Arguments @(
        "-n", "$writeRequests",
        "-c", "$writeConcurrency",
        "-m", "POST",
        "-T", "application/json",
        "-H", "Authorization: $token",
        "-D", $writeBodyFile,
        "$BaseUrl/tweets/"
    )

    Write-Step "Cenarios em execucao concorrente"
    $scenarioRuns.Process | Wait-Process

    foreach ($scenarioRun in $scenarioRuns) {
        Write-Step "Resultado do cenario $($scenarioRun.Name)"
        if (Test-Path $scenarioRun.Stdout) {
            Get-Content -Path $scenarioRun.Stdout | Out-Host
            $tempFiles += $scenarioRun.Stdout
        }

        if (Test-Path $scenarioRun.Stderr) {
            $stderrContent = Get-Content -Path $scenarioRun.Stderr
            if ($stderrContent) {
                $stderrContent | Out-Host
            }

            $tempFiles += $scenarioRun.Stderr
        }
    }

    Write-Step "Coletando snapshot de metricas"
    $metrics = (Invoke-WebRequest -Uri $metricsUrl -UseBasicParsing).Content
    $metrics |
        Select-String "go_tweets_http_requests_total|go_tweets_http_request_errors_total|go_tweets_http_request_duration_seconds_count|go_tweets_db_query_duration_seconds_count" |
        ForEach-Object { $_.Line } |
        Out-Host

    Write-Step "Teste realista concluido"
} finally {
    foreach ($tempFile in $tempFiles) {
        if (Test-Path $tempFile) {
            Remove-Item $tempFile -Force -ErrorAction SilentlyContinue
        }
    }

    if ($apiStartedByScript -and $apiProcess -and -not $KeepRunning) {
        Write-Step "Parando API"
        Stop-Process -Id $apiProcess.Id -Force -ErrorAction SilentlyContinue
    }

    if ($dbStartedByScript -and -not $KeepRunning) {
        Write-Step "Parando docker compose"
        & docker compose down | Out-Host
    }
}
