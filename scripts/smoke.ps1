$ErrorActionPreference = "Stop"
Set-StrictMode -Version Latest

$root = Resolve-Path (Join-Path $PSScriptRoot "..")

Write-Host "[1/3] Running backend tests: go test ./..." -ForegroundColor Cyan
Push-Location $root
try {
  go test ./...

  Write-Host "[2/3] Running frontend lint: npm run lint" -ForegroundColor Cyan
  Push-Location (Join-Path $root "frontend")
  try {
    npm run lint

    Write-Host "[3/3] Running frontend build: npm run build" -ForegroundColor Cyan
    npm run build
  }
  finally {
    Pop-Location
  }

  Write-Host "Smoke checks passed." -ForegroundColor Green
}
finally {
  Pop-Location
}
