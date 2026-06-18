Write-Host "========================================" -ForegroundColor Cyan
Write-Host "wx_channel build script" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

Write-Host "[1/4] Clean old files..." -ForegroundColor Yellow
Remove-Item -Path "wx_channel.exe" -ErrorAction SilentlyContinue
Remove-Item -Path "rsrc_windows_*.syso" -ErrorAction SilentlyContinue

Write-Host "[2/4] Generate Windows resources..." -ForegroundColor Yellow
$goBin = go env GOPATH
$winres = Join-Path $goBin "bin\go-winres.exe"
if (-not (Test-Path $winres)) {
    Write-Host "Installing go-winres..." -ForegroundColor Yellow
    go install github.com/tc-hib/go-winres@latest
}
& $winres make
if ($LASTEXITCODE -ne 0) {
    Write-Host "ERROR: Resource generation failed" -ForegroundColor Red
    exit 1
}

Write-Host "[3/4] Compiling..." -ForegroundColor Yellow
$env:CGO_ENABLED = "1"
go build -ldflags="-s -w" -o wx_channel.exe
if ($LASTEXITCODE -ne 0) {
    Write-Host "ERROR: Build failed" -ForegroundColor Red
    exit 1
}

Write-Host "[4/4] Verify build..." -ForegroundColor Yellow
.\wx_channel.exe version

Write-Host ""
Write-Host "========================================" -ForegroundColor Green
Write-Host "Build complete!" -ForegroundColor Green
Write-Host "Output: wx_channel.exe" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green
