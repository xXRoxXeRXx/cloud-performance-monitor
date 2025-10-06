# Test script for MagentaCLOUD debugging with detailed logs

Write-Host "Setting up environment for MagentaCLOUD debugging..." -ForegroundColor Green

# Ensure .env exists
if (!(Test-Path ".env")) {
    Write-Host "Creating .env from example..." -ForegroundColor Yellow
    Copy-Item ".env.example" ".env"
}

# Set debug logging
Write-Host "Setting LOG_LEVEL=DEBUG in .env..." -ForegroundColor Yellow
$envContent = Get-Content ".env"
$envContent = $envContent -replace "LOG_LEVEL=.*", "LOG_LEVEL=DEBUG"
if ($envContent -notcontains "LOG_LEVEL=DEBUG") {
    $envContent += "LOG_LEVEL=DEBUG"
}
$envContent | Set-Content ".env"

# Build and start the container
Write-Host "Building and starting the monitor with debug logging..." -ForegroundColor Green
docker compose down
docker compose build monitor-agent
docker compose up -d monitor-agent

Write-Host "Waiting for service to start..." -ForegroundColor Yellow
Start-Sleep 5

Write-Host "Tailing logs with detailed debugging..." -ForegroundColor Green
Write-Host "Look for the following in the logs:" -ForegroundColor Cyan
Write-Host "  - [DEBUG] [http_request] entries showing exact HTTP requests" -ForegroundColor White
Write-Host "  - [DEBUG] [http_response] entries showing server responses" -ForegroundColor White
Write-Host "  - [INFO] [cleanup] entries showing pre-upload cleanup" -ForegroundColor White
Write-Host "  - [INFO] [debug] entries during 409 conflicts with directory listings" -ForegroundColor White
Write-Host "  - [WARN] [cleanup] entries if old upload directories are found" -ForegroundColor White
Write-Host ""

# Follow logs until interrupted
docker compose logs -f monitor-agent