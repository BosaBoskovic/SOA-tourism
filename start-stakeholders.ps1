param(
    [string]$Neo4jUri = "neo4j://localhost:7687",
    [string]$Neo4jUser = "neo4j",
    [string]$Neo4jPassword = "password",
    [string]$Neo4jDatabase = "neo4j",
    [string]$JwtSecret = "super-secret-change-me-123456",
    [switch]$PersistEnv,
    [switch]$SkipNeo4j
)

$ErrorActionPreference = "Stop"

function Require-Command {
    param([string]$Name)

    if (-not (Get-Command $Name -ErrorAction SilentlyContinue)) {
        throw "Required command '$Name' is not installed or not available in PATH."
    }
}

function Set-SessionEnv {
    param(
        [string]$Name,
        [string]$Value
    )

    Set-Item -Path "Env:$Name" -Value $Value
}

function Set-PersistentEnv {
    param(
        [string]$Name,
        [string]$Value
    )

    $null = setx $Name $Value
}

function Ensure-DockerDaemon {
    try {
        $null = docker info 2>$null
    }
    catch {
        throw "Docker is installed but daemon is not running. Start Docker Desktop and try again."
    }
}

function Wait-ForTcpPort {
    param(
        [string]$HostName,
        [int]$Port,
        [int]$TimeoutSeconds = 45
    )

    $stopwatch = [System.Diagnostics.Stopwatch]::StartNew()

    while ($stopwatch.Elapsed.TotalSeconds -lt $TimeoutSeconds) {
        $client = New-Object System.Net.Sockets.TcpClient
        try {
            $async = $client.BeginConnect($HostName, $Port, $null, $null)
            if ($async.AsyncWaitHandle.WaitOne(1000, $false)) {
                $client.EndConnect($async)
                $client.Close()
                return $true
            }
        }
        catch {
            # Keep waiting until timeout.
        }
        finally {
            $client.Close()
        }
    }

    return $false
}

Require-Command -Name go

Set-SessionEnv -Name "NEO4J_URI" -Value $Neo4jUri
Set-SessionEnv -Name "NEO4J_USER" -Value $Neo4jUser
Set-SessionEnv -Name "NEO4J_PASSWORD" -Value $Neo4jPassword
Set-SessionEnv -Name "NEO4J_DATABASE" -Value $Neo4jDatabase
Set-SessionEnv -Name "JWT_SECRET" -Value $JwtSecret

if ($PersistEnv) {
    Set-PersistentEnv -Name "NEO4J_URI" -Value $Neo4jUri
    Set-PersistentEnv -Name "NEO4J_USER" -Value $Neo4jUser
    Set-PersistentEnv -Name "NEO4J_PASSWORD" -Value $Neo4jPassword
    Set-PersistentEnv -Name "NEO4J_DATABASE" -Value $Neo4jDatabase
    Set-PersistentEnv -Name "JWT_SECRET" -Value $JwtSecret

    Write-Host "Environment variables are also saved with setx for future terminals."
}

if (-not $SkipNeo4j) {
    Require-Command -Name docker
    Ensure-DockerDaemon

    $containerName = "soa-tourism-neo4j"
    $existing = docker ps -a --filter "name=^/$containerName$" --format "{{.Names}}"

    if (-not $existing) {
        Write-Host "Creating Neo4j container '$containerName'..."
        docker run -d --name $containerName -p 7474:7474 -p 7687:7687 -e "NEO4J_AUTH=$Neo4jUser/$Neo4jPassword" neo4j:5 | Out-Null
    }
    else {
        $running = docker ps --filter "name=^/$containerName$" --format "{{.Names}}"
        if (-not $running) {
            Write-Host "Starting existing Neo4j container '$containerName'..."
            docker start $containerName | Out-Null
        }
    }

    Write-Host "Waiting for Neo4j on localhost:7687..."
    if (-not (Wait-ForTcpPort -HostName "localhost" -Port 7687 -TimeoutSeconds 60)) {
        throw "Neo4j did not become available on localhost:7687 in time."
    }
}

$servicePath = Join-Path $PSScriptRoot "services/stakeholders"
if (-not (Test-Path $servicePath)) {
    throw "Service path not found: $servicePath"
}

Push-Location $servicePath
try {
    Write-Host "Starting stakeholders service on :8081..."
    go run .
}
finally {
    Pop-Location
}
