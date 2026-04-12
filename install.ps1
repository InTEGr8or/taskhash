# Powershell installation script for taskhash
$ErrorActionPreference = "Stop"

$Repo = "InTEGr8or/taskhash"
$Version = "latest"
$InstallDir = $PWD
$TaskhashBin = Join-Path $InstallDir "taskhash.exe"

Write-Host "Installing taskhash to $InstallDir..." -ForegroundColor Cyan

# Logic to detect/download or build.
# For now, let's keep it simple: assume a direct build-from-source to start until releases exist.
# Or better, let's try to fetch from release if it exists, but the user said "We haven't released a Windows binary".
# So, for now, fall back to go build if it exists.

Write-Host "Building from source..." -ForegroundColor Yellow
if (-not (Get-Command "go" -ErrorAction SilentlyContinue)) {
    Write-Error "Go is required to build from source. Please install Go."
    exit 1
}

go build -o $TaskhashBin taskhash.go

Write-Host "taskhash installed at $TaskhashBin" -ForegroundColor Green

# Initialize
Write-Host "Initializing taskhash..." -ForegroundColor Cyan
& $TaskhashBin init

Write-Host "Done! Run './taskhash.exe --help' for usage." -ForegroundColor Green
