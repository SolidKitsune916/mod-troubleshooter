# Prerequisites Check Script
Write-Host "Checking prerequisites..." -ForegroundColor Cyan
Write-Host ""

# Check Node.js
Write-Host "Checking Node.js..." -ForegroundColor Yellow
try {
    $nodeVersion = node --version 2>$null
    if ($nodeVersion) {
        Write-Host "  ✓ Node.js found: $nodeVersion" -ForegroundColor Green
    } else {
        Write-Host "  ✗ Node.js not found" -ForegroundColor Red
        Write-Host "    Download from: https://nodejs.org/" -ForegroundColor Yellow
    }
} catch {
    Write-Host "  ✗ Node.js not found" -ForegroundColor Red
    Write-Host "    Download from: https://nodejs.org/" -ForegroundColor Yellow
}

# Check npm
Write-Host "Checking npm..." -ForegroundColor Yellow
try {
    $npmVersion = npm --version 2>$null
    if ($npmVersion) {
        Write-Host "  ✓ npm found: $npmVersion" -ForegroundColor Green
    } else {
        Write-Host "  ✗ npm not found" -ForegroundColor Red
    }
} catch {
    Write-Host "  ✗ npm not found" -ForegroundColor Red
}

# Check Go
Write-Host "Checking Go..." -ForegroundColor Yellow
try {
    $goVersion = go version 2>$null
    if ($goVersion) {
        Write-Host "  ✓ Go found: $goVersion" -ForegroundColor Green
    } else {
        Write-Host "  ✗ Go not found" -ForegroundColor Red
        Write-Host "    Download from: https://go.dev/dl/" -ForegroundColor Yellow
    }
} catch {
    Write-Host "  ✗ Go not found" -ForegroundColor Red
    Write-Host "    Download from: https://go.dev/dl/" -ForegroundColor Yellow
}

Write-Host ""
Write-Host "Next Steps:" -ForegroundColor Cyan
Write-Host "1. Install missing tools (see SETUP_GUIDE.md for details)" -ForegroundColor White
Write-Host "2. Close and reopen this terminal after installation" -ForegroundColor White
Write-Host "3. Run this script again to verify installation" -ForegroundColor White
Write-Host "4. Once all tools are installed, follow START_SERVERS.md" -ForegroundColor White
