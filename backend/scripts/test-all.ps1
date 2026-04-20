param(
    [string]$EnvFile = ".env.test",
    [switch]$Coverage
)

$ErrorActionPreference = "Stop"

$projectRoot = Split-Path -Parent $PSScriptRoot
Set-Location $projectRoot

$env:GOCACHE = Join-Path $projectRoot ".gocache"

function Load-EnvFile {
    param([string]$Path)

    if (-not (Test-Path $Path)) {
        Write-Error "Environment file not found: $Path"
    }

    Get-Content $Path | ForEach-Object {
        if ($_ -match '^\s*#' -or $_ -match '^\s*$') { return }
        $name, $value = $_ -split '=', 2
        Set-Item -Path ("Env:" + $name) -Value $value
    }
}

if ($Coverage) {
    Write-Host "Running unit tests with coverage..."
    go test ./... -coverprofile unit.cover.out
    if ($LASTEXITCODE -ne 0) {
        exit $LASTEXITCODE
    }

    Write-Host "Loading integration environment from $EnvFile..."
    Load-EnvFile -Path $EnvFile

    Write-Host "Running integration tests with repository coverage..."
    go test -count=1 -tags=integration -coverpkg=project-management/internal/repository -coverprofile repository.cover.out ./tests/integration/...
    if ($LASTEXITCODE -ne 0) {
        exit $LASTEXITCODE
    }

    Write-Host "Running db integration coverage..."
    go test -count=1 -tags=integration ./internal/db -coverprofile db.cover.out
    exit $LASTEXITCODE
}

Write-Host "Running unit tests..."
go test ./...
if ($LASTEXITCODE -ne 0) {
    exit $LASTEXITCODE
}

Write-Host "Loading integration environment from $EnvFile..."
Load-EnvFile -Path $EnvFile

Write-Host "Running integration tests..."
go test -tags=integration ./tests/integration/...
if ($LASTEXITCODE -ne 0) {
    exit $LASTEXITCODE
}

go test -tags=integration ./internal/repository ./internal/db
exit $LASTEXITCODE
