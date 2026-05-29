param(
    [string]$BaseUrl = "http://127.0.0.1:8080",
    [int]$Requests = 3000,
    [int]$Concurrency = 50,
    [int]$SeedPosts = 20,
    [switch]$SkipSetup,
    [switch]$KeepRunning
)

$ErrorActionPreference = "Stop"

function Write-Step {
    param([string]$Message)
    Write-Host "[stress-test] $Message"
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

$projectRoot = Split-Path -Parent $PSScriptRoot
$healthUrl = "$BaseUrl/check-health"
$metricsUrl = "$BaseUrl/metrics"
$apiProcess = $null
$apiStartedByScript = $false
$dbStartedByScript = $false

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
    $email = "stress-$timestamp@example.com"
    $username = "stress$timestamp"
    $password = "secret123"

    Write-Step "Criando usuario para o teste"
    $registerBody = @{
        email            = $email
        username         = $username
        password         = $password
        password_confirm = $password
    } | ConvertTo-Json

    Invoke-RestMethod -Method Post -Uri "$BaseUrl/auth/register" -ContentType "application/json" -Body $registerBody | Out-Null

    Write-Step "Autenticando usuario de carga"
    $loginBody = @{
        email    = $email
        password = $password
    } | ConvertTo-Json

    $loginResponse = Invoke-RestMethod -Method Post -Uri "$BaseUrl/auth/login" -ContentType "application/json" -Body $loginBody
    $token = $loginResponse.token

    Write-Step "Gerando $SeedPosts posts de seed"
    1..$SeedPosts | ForEach-Object {
        $postBody = @{
            title   = "Stress Post $_"
            content = "Payload generated for reusable stress test $_"
        } | ConvertTo-Json

        Invoke-RestMethod -Method Post -Uri "$BaseUrl/tweets/" -Headers @{ Authorization = $token } -ContentType "application/json" -Body $postBody | Out-Null
    }

    $heyPath = Ensure-HeyInstalled
    $targetUrl = "$BaseUrl/tweets/?page=1&limit=10"

    Write-Step "Executando carga com $Requests requests e concorrencia $Concurrency"
    & $heyPath -n $Requests -c $Concurrency $targetUrl | Out-Host

    Write-Step "Coletando snapshot de metricas"
    $metrics = (Invoke-WebRequest -Uri $metricsUrl -UseBasicParsing).Content
    $metrics |
        Select-String "go_tweets_http_requests_total|go_tweets_http_request_errors_total|go_tweets_http_request_duration_seconds_count|go_tweets_db_query_duration_seconds_count" |
        ForEach-Object { $_.Line } |
        Out-Host

    Write-Step "Teste concluido"
} finally {
    if ($apiStartedByScript -and $apiProcess -and -not $KeepRunning) {
        Write-Step "Parando API"
        Stop-Process -Id $apiProcess.Id -Force -ErrorAction SilentlyContinue
    }

    if ($dbStartedByScript -and -not $KeepRunning) {
        Write-Step "Parando docker compose"
        & docker compose down | Out-Host
    }
}
